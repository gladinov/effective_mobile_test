package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gladinov/e"
	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	errInvalidRequestBody      error = errors.New("invalid request body")
	errInvalidUserIDQueryParam error = errors.New("invalid userID query param")

	errGetData     error = errors.New("internal error")
	errInvalidUUID error = errors.New("invalid uuid")
	errNotFound    error = errors.New("not Found")
)

type Handler interface {
	Routes() *echo.Echo
}

type handler struct {
	logger  *slog.Logger
	service Service
}

func NewHandler(logger *slog.Logger, service Service) *handler {
	return &handler{
		logger:  logger,
		service: service,
	}
}

func (h *handler) Routes() *echo.Echo {
	router := echo.New()

	router.Use(middleware.CORS())
	router.Use(h.LoggerMiddleWare)
	router.HTTPErrorHandler = HTTPErrorHandler(h.logger)
	// TODO: Добавить healthcheck
	router.POST("/subscriptions/create", h.Create)
	router.GET("/subscriptions/get/:id", h.Get)
	router.PUT("/subscriptions/update/:id", h.Update)
	router.DELETE("/subscriptions/delete/:id", h.Delete)
	router.GET("/subscriptions/list", h.List)
	router.GET("/subscriptions/total", h.Total)

	return router
}

type Service interface {
	Create(ctx context.Context, sub domain.Subscription) (uuid.UUID, error)
	GetByID(ctx context.Context, subID uuid.UUID) (domain.Subscription, error)
	UpdateByID(ctx context.Context, subID uuid.UUID, sub domain.Subscription) error
	DeleteByID(ctx context.Context, subID uuid.UUID) error
	List(ctx context.Context) ([]domain.Subscription, error)
	GetTotal(ctx context.Context, filter domain.FilterTotal) (int, error)
}

func (h *handler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	// TODO: Подобрать таймауты
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var subsReq subscriptionRequest

	err := c.Bind(&subsReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}

	domainSub := subsReq.ToDomain()

	subID, err := h.service.Create(ctx, domainSub)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	resp := CreateResponce{SubID: subID}

	return c.JSON(http.StatusCreated, resp)
}

func (h *handler) Get(c echo.Context) error {
	ctx := c.Request().Context()
	// TODO: Подобрать таймауты
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	id := c.Param("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidUUID)
	}

	sub, err := h.service.GetByID(ctx, uuid)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, errNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	resp := mapDomainSubToDTOSubResponce(sub)

	return c.JSON(http.StatusOK, resp)
}

func (h *handler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	// TODO: Подобрать таймауты
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	id := c.Param("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidUUID)
	}

	var subsReq subscriptionRequest

	err = c.Bind(&subsReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}
	domainSub := subsReq.ToDomain()

	err = h.service.UpdateByID(ctx, uuid, domainSub)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, errNotFound)
		}
		// TODO: Логирование внутренних ошибок
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	// TODO: Подобрать таймауты
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	id := c.Param("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidUUID)
	}

	err = h.service.DeleteByID(ctx, uuid)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, errNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *handler) List(c echo.Context) error {
	ctx := c.Request().Context()
	// TODO: Подобрать таймауты
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	domainSubs, err := h.service.List(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, errNotFound)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	subs := make([]subscriptionResponce, 0, len(domainSubs))
	for i := range domainSubs {
		subs = append(subs, mapDomainSubToDTOSubResponce(domainSubs[i]))
	}

	return c.JSON(http.StatusOK, subs)
}

func (h *handler) Total(c echo.Context) error {
	ctx := c.Request().Context()
	// TODO: Подобрать таймауты
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	filterTotal, err := h.getQueryForTotal(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidUserIDQueryParam)
	}

	domainFilter := filterTotal.ToDomain()

	total, err := h.service.GetTotal(ctx, domainFilter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	totalResponce := totalResponce{
		Total: total,
	}
	return c.JSON(http.StatusOK, totalResponce)
}

func (h *handler) getQueryForTotal(c echo.Context) (filterTotal, error) {
	var filter filterTotal

	userID, err := userIdFromQueryParams(c)
	if err != nil {
		// TODO: логируем внутреннюю ошибку на этом уровне и возвращаем из функции ошибку для http ответа
		return filterTotal{}, errInvalidUserIDQueryParam
	}
	filter.UserID = userID

	serviceName := serviceNameFromQueryParams(c)
	filter.ServiceName = serviceName

	from, err := getYearMonthFromQueryParams(fromYear, fromMonth, c)
	if err != nil {
		// TODO: Обработать нормально ошибки
		return filterTotal{}, err
	}

	filter.From = from

	to, err := getYearMonthFromQueryParams(toYear, toMonth, c)
	if err != nil {
		return filterTotal{}, err
	}

	filter.To = to

	return filter, nil
}

func userIdFromQueryParams(c echo.Context) (*uuid.UUID, error) {
	value, exist := c.QueryParams()[userID]
	if !exist {
		return nil, nil
	}

	if len(value) == 0 || value[0] == "" {
		return nil, errors.New("user_id param can't be empty")
	}
	uuid, err := uuid.Parse(value[0])
	if err != nil {
		return nil, e.WrapIfErr("failed to parse uuid from string", err)
	}

	return &uuid, nil
}

func serviceNameFromQueryParams(c echo.Context) *string {
	value, exist := c.QueryParams()[serviceName]
	if !exist {
		return nil
	}

	res := value[0]

	return &res
}

func getYearMonthFromQueryParams(yearConst, monthConst string, c echo.Context) (*yearMonth, error) {
	valueYear, existYear := c.QueryParams()[yearConst]
	valueMonth, existMonth := c.QueryParams()[monthConst]

	switch {
	case !existYear && !existMonth:
		return nil, nil
	case !existYear:
		return nil, errors.New("year is empty, but month exist")
	case !existMonth:
		return nil, errors.New("month is empty, but year exist")
	}
	if len(valueYear) == 0 || valueYear[0] == "" {
		return nil, errors.New("year can't be empty")
	}
	if len(valueMonth) == 0 || valueMonth[0] == "" {
		return nil, errors.New("month can't be empty")
	}

	year, err := strconv.Atoi(valueYear[0])
	if err != nil {
		return nil, errors.New("failed to conv year query param to int")
	}
	month, err := strconv.Atoi(valueMonth[0])
	if err != nil {
		return nil, errors.New("failed to conv month query param to int")
	}

	res := yearMonth{
		Year:  year,
		Month: time.Month(month),
	}

	return &res, nil
}

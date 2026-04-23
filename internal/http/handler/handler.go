package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler interface {
	RegisterRoutes(router *echo.Echo)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=Service
type Service interface {
	Create(ctx context.Context, sub domain.Subscription) (uuid.UUID, error)
	GetByID(ctx context.Context, subID uuid.UUID) (domain.Subscription, error)
	UpdateByID(ctx context.Context, subID uuid.UUID, sub domain.Subscription) error
	DeleteByID(ctx context.Context, subID uuid.UUID) error
	List(ctx context.Context) ([]domain.Subscription, error)
	GetTotal(ctx context.Context, filter domain.FilterTotal) (int, error)
}

type handler struct {
	service        Service
	requestTimeout time.Duration
	logger         *slog.Logger
}

func NewHandler(logger *slog.Logger,
	service Service,
	requestTimeout time.Duration,
) *handler {
	return &handler{
		logger:         logger,
		service:        service,
		requestTimeout: requestTimeout,
	}
}

func (h *handler) RegisterRoutes(router *echo.Echo) {
	// TODO: Добавить healthcheck
	router.POST("/subscriptions/create", h.Create)
	router.GET("/subscriptions/get/:id", h.Get)
	router.PUT("/subscriptions/update/:id", h.Update)
	router.DELETE("/subscriptions/delete/:id", h.Delete)
	router.GET("/subscriptions/list", h.List)
	router.GET("/subscriptions/total", h.Total)
}

func (h *handler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	ctx, cancel := context.WithTimeout(ctx, h.requestTimeout)
	defer cancel()

	var subsReq subscriptionRequest

	err := c.Bind(&subsReq)
	if err != nil {
		return mapSubscriptionRequestError(err)
	}

	domainSub, err := subsReq.ToDomain()
	if err != nil {
		return mapSubscriptionRequestError(err)
	}

	subID, err := h.service.Create(ctx, domainSub)
	if err != nil {
		h.logger.Error("failed to create subscription",
			slog.Any("error", err),
			slog.Any("subscription", domainSub),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	resp := CreateResponse{SubID: subID}

	return c.JSON(http.StatusCreated, resp)
}

func (h *handler) Get(c echo.Context) error {
	ctx := c.Request().Context()
	ctx, cancel := context.WithTimeout(ctx, h.requestTimeout)
	defer cancel()

	id := c.Param("id")
	subID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidUUID)
	}

	sub, err := h.service.GetByID(ctx, subID)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, errNotFound)
		}

		h.logger.Error("failed to get subscription by id",
			slog.Any("error", err),
			slog.String("subscription_id", subID.String()),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	resp := mapDomainSubToDTOSubResponse(sub)

	return c.JSON(http.StatusOK, resp)
}

func (h *handler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	ctx, cancel := context.WithTimeout(ctx, h.requestTimeout)
	defer cancel()

	id := c.Param("id")
	subID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidUUID)
	}

	var subsReq subscriptionRequest

	err = c.Bind(&subsReq)
	if err != nil {
		return mapSubscriptionRequestError(err)
	}
	domainSub, err := subsReq.ToDomain()
	if err != nil {
		return mapSubscriptionRequestError(err)
	}

	err = h.service.UpdateByID(ctx, subID, domainSub)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, errNotFound)
		}

		h.logger.Error("failed to update subscription",
			slog.Any("error", err),
			slog.String("subscription_id", subID.String()),
			slog.Any("subscription", domainSub),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) Delete(c echo.Context) error {
	ctx := c.Request().Context()
	ctx, cancel := context.WithTimeout(ctx, h.requestTimeout)
	defer cancel()

	id := c.Param("id")
	subID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidUUID)
	}

	err = h.service.DeleteByID(ctx, subID)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, errNotFound)
		}

		h.logger.Error("failed to delete subscription",
			slog.Any("error", err),
			slog.String("subscription_id", subID.String()),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *handler) List(c echo.Context) error {
	ctx := c.Request().Context()
	ctx, cancel := context.WithTimeout(ctx, h.requestTimeout)
	defer cancel()

	domainSubs, err := h.service.List(ctx)
	if err != nil {
		h.logger.Error("failed to list subscriptions",
			slog.Any("error", err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	subs := make([]subscriptionResponse, 0, len(domainSubs))
	for i := range domainSubs {
		subs = append(subs, mapDomainSubToDTOSubResponse(domainSubs[i]))
	}

	return c.JSON(http.StatusOK, subs)
}

func (h *handler) Total(c echo.Context) error {
	ctx := c.Request().Context()
	ctx, cancel := context.WithTimeout(ctx, h.requestTimeout)
	defer cancel()

	domainFilter, err := getQueryForTotal(c)
	if err != nil {
		return mapTotalQueryError(err)
	}

	total, err := h.service.GetTotal(ctx, domainFilter)
	if err != nil {
		h.logger.Error("failed to get total",
			slog.Any("error", err),
			slog.Any("filter", domainFilter),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	totalResponse := totalResponse{
		Total: total,
	}
	return c.JSON(http.StatusOK, totalResponse)
}

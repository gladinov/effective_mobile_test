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
	router.GET("/health", h.Healthcheck)
	router.POST("/subscriptions", h.Create)
	router.GET("/subscriptions/:id", h.Get)
	router.PUT("/subscriptions/:id", h.Update)
	router.DELETE("/subscriptions/:id", h.Delete)
	router.GET("/subscriptions", h.List)
	router.GET("/subscriptions/total", h.Total)
}

// Healthcheck returns the application health status.
// @Summary Healthcheck
// @Description Returns service health status
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *handler) Healthcheck(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}

// Create creates a new subscription.
// @Summary Create subscription
// @Description Creates a new user subscription record. user_id must be UUID, price must be zero or greater, start_date and end_date use MM-YYYY format.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body SubscriptionRequest true "Subscription payload. start_date and end_date use MM-YYYY format, user_id must be UUID."
// @Success 201 {object} CreateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [post]
func (h *handler) Create(c echo.Context) error {
	ctx := c.Request().Context()
	ctx, cancel := context.WithTimeout(ctx, h.requestTimeout)
	defer cancel()

	var subsReq SubscriptionRequest

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
		h.logger.Error("create subscription",
			slog.Any("error", err),
			slog.Any("subscription", domainSub),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	resp := CreateResponse{SubID: subID}

	return c.JSON(http.StatusCreated, resp)
}

// Get returns a subscription by ID.
// @Summary Get subscription
// @Description Returns a subscription by its ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Success 200 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [get]
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

		h.logger.Error("get subscription by id",
			slog.Any("error", err),
			slog.String("subscription_id", subID.String()),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	resp := mapDomainSubToDTOSubResponse(sub)

	return c.JSON(http.StatusOK, resp)
}

// Update updates a subscription by ID.
// @Summary Update subscription
// @Description Updates an existing subscription by its ID. user_id must be UUID, price must be zero or greater, start_date and end_date use MM-YYYY format.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Param subscription body SubscriptionRequest true "Subscription payload. start_date and end_date use MM-YYYY format, user_id must be UUID."
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *handler) Update(c echo.Context) error {
	ctx := c.Request().Context()
	ctx, cancel := context.WithTimeout(ctx, h.requestTimeout)
	defer cancel()

	id := c.Param("id")
	subID, err := uuid.Parse(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidUUID)
	}

	var subsReq SubscriptionRequest

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

		h.logger.Error("update subscription",
			slog.Any("error", err),
			slog.String("subscription_id", subID.String()),
			slog.Any("subscription", domainSub),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	return c.NoContent(http.StatusNoContent)
}

// Delete removes a subscription by ID.
// @Summary Delete subscription
// @Description Deletes a subscription by its ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [delete]
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

		h.logger.Error("delete subscription",
			slog.Any("error", err),
			slog.String("subscription_id", subID.String()),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	return c.NoContent(http.StatusNoContent)
}

// List returns all subscriptions.
// @Summary List subscriptions
// @Description Returns all subscriptions
// @Tags subscriptions
// @Produce json
// @Success 200 {array} SubscriptionResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [get]
func (h *handler) List(c echo.Context) error {
	ctx := c.Request().Context()
	ctx, cancel := context.WithTimeout(ctx, h.requestTimeout)
	defer cancel()

	domainSubs, err := h.service.List(ctx)
	if err != nil {
		h.logger.Error("list subscriptions",
			slog.Any("error", err),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}
	subs := make([]SubscriptionResponse, 0, len(domainSubs))
	for i := range domainSubs {
		subs = append(subs, mapDomainSubToDTOSubResponse(domainSubs[i]))
	}

	return c.JSON(http.StatusOK, subs)
}

// Total calculates the total subscription cost for a selected period.
// @Summary Get total subscriptions cost
// @Description Calculates the total cost of subscriptions for the selected period with optional filters
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by user ID (UUID)"
// @Param service_name query string false "Filter by service name"
// @Param from query string false "Start period in MM-YYYY format"
// @Param to query string false "End period in MM-YYYY format"
// @Success 200 {object} TotalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/total [get]
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
		h.logger.Error("get total",
			slog.Any("error", err),
			slog.Any("filter", domainFilter),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, errGetData)
	}

	totalResponse := TotalResponse{
		Total: total,
	}
	return c.JSON(http.StatusOK, totalResponse)
}

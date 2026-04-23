package handler

import (
	"errors"
	"net/http"

	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/labstack/echo/v4"
)

func mapSubscriptionRequestError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidUUID):
		return echo.NewHTTPError(http.StatusBadRequest, "user_id must be a valid UUID")
	case errors.Is(err, ErrServiceNameRequired):
		return echo.NewHTTPError(http.StatusBadRequest, "service_name must not be empty")
	case errors.Is(err, ErrPriceInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, "price must be greater than zero")
	case errors.Is(err, ErrStartDateRequired):
		return echo.NewHTTPError(http.StatusBadRequest, "start_date is required")
	case errors.Is(err, ErrDateEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, "date must not be empty")
	case errors.Is(err, ErrDateInvalidFormat):
		return echo.NewHTTPError(http.StatusBadRequest, "date must be in MM-YYYY format")
	case errors.Is(err, ErrMonthOutOfRange):
		return echo.NewHTTPError(http.StatusBadRequest, "month must be between 1 and 12")
	case errors.Is(err, ErrEndDateBeforeStart):
		return echo.NewHTTPError(http.StatusBadRequest, "end_date must not be before start_date")
	default:
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}
}

func mapDomainSubToDTOSubResponce(domainSub domain.Subscription) subscriptionResponce {
	return subscriptionResponce{
		ID:          domainSub.ID,
		ServiceName: domainSub.ServiceName,
		Price:       domainSub.Price,
		UserID:      domainSub.UserID,
		StartDate:   formatYearMonth(domainSub.StartDate),
		EndDate:     formatYearMonthPtr(domainSub.EndDate),
	}
}

func mapTotalQueryError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidUUID):
		return echo.NewHTTPError(http.StatusBadRequest, "user_id must be a valid UUID")

	case errors.Is(err, ErrUserIDEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, "user_id must not be empty")

	case errors.Is(err, ErrUserIDMultipleValues):
		return echo.NewHTTPError(http.StatusBadRequest, "user_id must be specified once")

	case errors.Is(err, ErrServiceNameEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, "service_name must not be empty")

	case errors.Is(err, ErrServiceNameMultipleValues):
		return echo.NewHTTPError(http.StatusBadRequest, "service_name must be specified once")

	case errors.Is(err, ErrDateEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, "date must not be empty")

	case errors.Is(err, ErrDateInvalidFormat):
		return echo.NewHTTPError(http.StatusBadRequest, "date must be in MM-YYYY format")

	case errors.Is(err, ErrDateMultipleValues):
		return echo.NewHTTPError(http.StatusBadRequest, "date query param must be specified once")

	case errors.Is(err, ErrMonthOutOfRange):
		return echo.NewHTTPError(http.StatusBadRequest, "month must be between 1 and 12")

	case errors.Is(err, ErrEndDateBeforeStart):
		return echo.NewHTTPError(http.StatusBadRequest, "to must not be before from")

	default:
		return echo.NewHTTPError(http.StatusBadRequest, "invalid query parameters")
	}
}

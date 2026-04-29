package handler

import (
	"errors"
	"net/http"

	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/labstack/echo/v4"
)

func mapSubscriptionRequestError(err error) error {
	switch {
	case errors.Is(err, errUserIDEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, "user_id must not be empty")
	case errors.Is(err, errInvalidUUID):
		return echo.NewHTTPError(http.StatusBadRequest, "user_id must be a valid UUID")
	case errors.Is(err, errServiceNameRequired):
		return echo.NewHTTPError(http.StatusBadRequest, "service_name must not be empty")
	case errors.Is(err, errPriceInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, "price must not be negative")
	case errors.Is(err, errStartDateRequired):
		return echo.NewHTTPError(http.StatusBadRequest, "start_date is required")
	case errors.Is(err, errDateEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, "date must not be empty")
	case errors.Is(err, errDateInvalidFormat):
		return echo.NewHTTPError(http.StatusBadRequest, "date must be in MM-YYYY format")
	case errors.Is(err, errMonthOutOfRange):
		return echo.NewHTTPError(http.StatusBadRequest, "month must be between 1 and 12")
	case errors.Is(err, errYearOutOfRange):
		return echo.NewHTTPError(http.StatusBadRequest, "year must be greater than zero")
	case errors.Is(err, errEndDateBeforeStart):
		return echo.NewHTTPError(http.StatusBadRequest, "end_date must not be before start_date")
	default:
		return echo.NewHTTPError(http.StatusBadRequest, errInvalidRequestBody)
	}
}

func mapSubscriptionUpdateRequestError(err error) error {
	switch {
	case errors.Is(err, errUpdateEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, "update payload must contain at least one field")
	case errors.Is(err, errEndDateUpdateConflict):
		return echo.NewHTTPError(http.StatusBadRequest, "end_date and clear_end_date=true cannot be used together")
	default:
		return mapSubscriptionRequestError(err)
	}
}

func mapDomainSubToDTOSubResponse(domainSub domain.Subscription) SubscriptionResponse {
	return SubscriptionResponse{
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
	case errors.Is(err, errInvalidUUID):
		return echo.NewHTTPError(http.StatusBadRequest, "user_id must be a valid UUID")

	case errors.Is(err, errUserIDEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, "user_id must not be empty")

	case errors.Is(err, errUserIDMultipleValues):
		return echo.NewHTTPError(http.StatusBadRequest, "user_id must be specified once")

	case errors.Is(err, errServiceNameEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, "service_name must not be empty")

	case errors.Is(err, errServiceNameMultipleValues):
		return echo.NewHTTPError(http.StatusBadRequest, "service_name must be specified once")

	case errors.Is(err, errDateEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, "date must not be empty")

	case errors.Is(err, errDateInvalidFormat):
		return echo.NewHTTPError(http.StatusBadRequest, "date must be in MM-YYYY format")

	case errors.Is(err, errDateMultipleValues):
		return echo.NewHTTPError(http.StatusBadRequest, "date query param must be specified once")

	case errors.Is(err, errMonthOutOfRange):
		return echo.NewHTTPError(http.StatusBadRequest, "month must be between 1 and 12")

	case errors.Is(err, errYearOutOfRange):
		return echo.NewHTTPError(http.StatusBadRequest, "year must be greater than zero")

	case errors.Is(err, errEndDateBeforeStart):
		return echo.NewHTTPError(http.StatusBadRequest, "to must not be before from")

	default:
		return echo.NewHTTPError(http.StatusBadRequest, "invalid query parameters")
	}
}

func mapPaginationQueryError(err error) error {
	switch {
	case errors.Is(err, errLimitEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, "limit must not be empty")

	case errors.Is(err, errLimitMultipleValues):
		return echo.NewHTTPError(http.StatusBadRequest, "limit must be specified once")

	case errors.Is(err, errLimitInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, "limit must be a positive integer")

	case errors.Is(err, errLimitTooLarge):
		return echo.NewHTTPError(http.StatusBadRequest, "limit must be less than or equal to 1000")

	case errors.Is(err, errOffsetEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, "offset must not be empty")

	case errors.Is(err, errOffsetMultipleValues):
		return echo.NewHTTPError(http.StatusBadRequest, "offset must be specified once")

	case errors.Is(err, errOffsetInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, "offset must be a non-negative integer")

	default:
		return echo.NewHTTPError(http.StatusBadRequest, "invalid pagination query parameters")
	}
}

package handler

import (
	"errors"
	"net/http"

	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/labstack/echo/v4"
)

func mapYearMonthPtrToDomain(y *yearMonth) *domain.YearMonth {
	if y == nil {
		return nil
	}
	return &domain.YearMonth{
		Year:  y.Year,
		Month: y.Month,
	}
}

func mapDomainSubToDTOSubResponce(domainSub domain.Subscription) subscriptionResponce {
	return subscriptionResponce{
		ID:          domainSub.ID,
		ServiceName: domainSub.ServiceName,
		Price:       domainSub.Price,
		UserID:      domainSub.UserID,
		StartDate:   MapDomainYearMonthToDto(domainSub.StartDate),
		EndDate:     MapDomainYearMonthPtrToDtoPtr(domainSub.EndDate),
	}
}

func MapDomainYearMonthToDto(date domain.YearMonth) yearMonth {
	return yearMonth{
		Year:  date.Year,
		Month: date.Month,
	}
}

func MapDomainYearMonthPtrToDtoPtr(date *domain.YearMonth) *yearMonth {
	if date == nil {
		return nil
	}
	return &yearMonth{
		Year:  date.Year,
		Month: date.Month,
	}
}

func mapYearMonthPtrToDomainPtr(date *yearMonth) *domain.YearMonth {
	if date == nil {
		return nil
	}
	return &domain.YearMonth{
		Year:  date.Year,
		Month: date.Month,
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

	case errors.Is(err, ErrYearRequired):
		return echo.NewHTTPError(http.StatusBadRequest, "year is required when month is provided")

	case errors.Is(err, ErrMonthRequired):
		return echo.NewHTTPError(http.StatusBadRequest, "month is required when year is provided")

	case errors.Is(err, ErrYearEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, "year must not be empty")

	case errors.Is(err, ErrMonthEmpty):
		return echo.NewHTTPError(http.StatusBadRequest, "month must not be empty")

	case errors.Is(err, ErrYearInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, "year must be a valid integer")

	case errors.Is(err, ErrMonthInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, "month must be a valid integer")

	case errors.Is(err, ErrYearMultipleValues):
		return echo.NewHTTPError(http.StatusBadRequest, "year must be specified once")

	case errors.Is(err, ErrMonthMultipleValues):
		return echo.NewHTTPError(http.StatusBadRequest, "month must be specified once")

	case errors.Is(err, ErrMonthOutOfRange):
		return echo.NewHTTPError(http.StatusBadRequest, "month must be between 1 and 12")

	default:
		return echo.NewHTTPError(http.StatusBadRequest, "invalid query parameters")
	}
}

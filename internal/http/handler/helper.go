package handler

import (
	"strconv"
	"strings"

	"github.com/gladinov/e"
	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/labstack/echo/v4"
)

func getPaginationQuery(c echo.Context) (domain.Pagination, error) {
	limitValue, err := limitFromQueryParam(c)
	if err != nil {
		return domain.Pagination{}, e.WrapIfErr("parse limit query param", err)
	}

	offsetValue, err := offsetFromQueryParam(c)
	if err != nil {
		return domain.Pagination{}, e.WrapIfErr("parse offset query param", err)
	}

	return domain.Pagination{
		Limit:  limitValue,
		Offset: offsetValue,
	}, nil
}

func limitFromQueryParam(c echo.Context) (uint64, error) {
	values, exists := c.QueryParams()[limit]
	if !exists {
		return defaultLimit, nil
	}

	if len(values) > 1 {
		return 0, errLimitMultipleValues
	}

	if len(values) == 0 {
		return 0, errLimitEmpty
	}

	value := strings.TrimSpace(values[0])
	if value == "" {
		return 0, errLimitEmpty
	}

	parsedLimit, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsedLimit == 0 {
		return 0, errLimitInvalid
	}

	if parsedLimit > maxLimit {
		return 0, errLimitTooLarge
	}

	return parsedLimit, nil
}

func offsetFromQueryParam(c echo.Context) (uint64, error) {
	values, exists := c.QueryParams()[offset]
	if !exists {
		return defaultOffset, nil
	}

	if len(values) > 1 {
		return 0, errOffsetMultipleValues
	}

	if len(values) == 0 {
		return 0, errOffsetEmpty
	}

	value := strings.TrimSpace(values[0])
	if value == "" {
		return 0, errOffsetEmpty
	}

	parsedOffset, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, errOffsetInvalid
	}

	return parsedOffset, nil
}

func getQueryForTotal(c echo.Context) (domain.FilterTotal, error) {
	var filter filterTotal

	userID, err := stringFromQueryParam(userID, c, errUserIDEmpty, errUserIDMultipleValues)
	if err != nil {
		return domain.FilterTotal{}, e.WrapIfErr("parse user_id query param", err)
	}
	filter.UserID = userID

	serviceName, err := stringFromQueryParam(serviceName, c, errServiceNameEmpty, errServiceNameMultipleValues)
	if err != nil {
		return domain.FilterTotal{}, e.WrapIfErr("parse service_name query param", err)
	}
	filter.ServiceName = serviceName

	fromPeriod, err := stringFromQueryParam(fromDate, c, errDateEmpty, errDateMultipleValues)
	if err != nil {
		return domain.FilterTotal{}, e.WrapIfErr("parse from query param", err)
	}
	filter.From = fromPeriod

	toPeriod, err := stringFromQueryParam(toDate, c, errDateEmpty, errDateMultipleValues)
	if err != nil {
		return domain.FilterTotal{}, e.WrapIfErr("parse to query param", err)
	}
	filter.To = toPeriod

	domainFilter, err := filter.ToDomain()
	if err != nil {
		return domain.FilterTotal{}, err
	}

	if domainFilter.From != nil && domainFilter.To != nil &&
		domainFilter.To.CountOfMonth() < domainFilter.From.CountOfMonth() {
		return domain.FilterTotal{}, domain.ErrEndDateBeforeStart
	}

	return domainFilter, nil
}

func stringFromQueryParam(param string, c echo.Context, emptyErr, multipleErr error) (*string, error) {
	values, exists := c.QueryParams()[param]
	if !exists {
		return nil, nil
	}

	if len(values) > 1 {
		return nil, multipleErr
	}

	if len(values) == 0 {
		return nil, emptyErr
	}

	value := strings.TrimSpace(values[0])
	if value == "" {
		return nil, emptyErr
	}

	return &value, nil
}

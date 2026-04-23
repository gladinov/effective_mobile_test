package handler

import (
	"strings"

	"github.com/gladinov/e"
	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/labstack/echo/v4"
)

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
		return domain.FilterTotal{}, errEndDateBeforeStart
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

package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gladinov/e"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func getQueryForTotal(c echo.Context) (filterTotal, error) {
	var filter filterTotal

	userID, err := userIdFromQueryParams(c)
	if err != nil {
		return filterTotal{}, e.WrapIfErr("parse user_id query param", err)
	}
	filter.UserID = userID

	serviceName, err := serviceNameFromQueryParam(c)
	if err != nil {
		return filterTotal{}, e.WrapIfErr("parse service_name query param", err)
	}
	filter.ServiceName = serviceName

	fromPeriod, err := getYearMonthFromQueryParams(fromYear, fromMonth, c)
	if err != nil {
		return filterTotal{}, e.WrapIfErr("parse from period query params", err)
	}

	filter.From = fromPeriod

	toPeriod, err := getYearMonthFromQueryParams(toYear, toMonth, c)
	if err != nil {
		return filterTotal{}, e.WrapIfErr("parse to period query params", err)
	}

	filter.To = toPeriod

	return filter, nil
}

func userIdFromQueryParams(c echo.Context) (*uuid.UUID, error) {
	values, exist := c.QueryParams()[userID]
	if !exist {
		return nil, nil
	}

	if len(values) > 1 {
		return nil, ErrUserIDMultipleValues
	}

	if len(values) == 0 {
		return nil, ErrUserIDEmpty
	}

	v := strings.TrimSpace(values[0])
	if v == "" {
		return nil, ErrUserIDEmpty
	}

	id, err := uuid.Parse(v)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	return &id, nil
}

func serviceNameFromQueryParam(c echo.Context) (*string, error) {
	values, exist := c.QueryParams()[serviceName]
	if !exist {
		return nil, nil
	}
	if len(values) > 1 {
		return nil, ErrServiceNameMultipleValues
	}

	if len(values) == 0 {
		return nil, ErrServiceNameEmpty
	}

	v := strings.TrimSpace(values[0])
	if v == "" {
		return nil, ErrServiceNameEmpty
	}

	return &v, nil
}

func getYearMonthFromQueryParams(yearParam, monthParam string, c echo.Context) (*yearMonth, error) {
	yearValues, yearExists := c.QueryParams()[yearParam]
	monthValues, monthExists := c.QueryParams()[monthParam]

	switch {
	case !yearExists && !monthExists:
		return nil, nil
	case !yearExists:
		return nil, ErrYearRequired
	case !monthExists:
		return nil, ErrMonthRequired
	}

	year, err := parseYear(yearValues)
	if err != nil {
		return nil, err
	}

	month, err := parseMonth(monthValues)
	if err != nil {
		return nil, err
	}

	return &yearMonth{
		Year:  year,
		Month: time.Month(month),
	}, nil
}

func parseYear(values []string) (int, error) {
	if len(values) > 1 {
		return 0, ErrYearMultipleValues
	}

	if len(values) == 0 {
		return 0, ErrYearEmpty
	}

	v := strings.TrimSpace(values[0])
	if v == "" {
		return 0, ErrYearEmpty
	}

	year, err := strconv.Atoi(v)
	if err != nil {
		return 0, ErrYearInvalid
	}

	return year, nil
}

func parseMonth(values []string) (int, error) {
	if len(values) > 1 {
		return 0, ErrMonthMultipleValues
	}

	if len(values) == 0 {
		return 0, ErrMonthEmpty
	}

	v := strings.TrimSpace(values[0])
	if v == "" {
		return 0, ErrMonthEmpty
	}

	month, err := strconv.Atoi(v)
	if err != nil {
		return 0, ErrMonthInvalid
	}

	if month < 1 || month > 12 {
		return 0, ErrMonthOutOfRange
	}

	return month, nil
}

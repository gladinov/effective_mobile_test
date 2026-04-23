package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/google/uuid"
)

const (
	userID      = "user_id"
	serviceName = "service_name"
	fromDate    = "from"
	toDate      = "to"
)

var (
	errInvalidRequestBody      error = errors.New("invalid request body")
	errInvalidUserIDQueryParam error = errors.New("invalid userID query param")
	errGetData                 error = errors.New("internal error")
	errInvalidUUID             error = errors.New("invalid uuid")
	errNotFound                error = errors.New("not Found")
)

var (
	errDateEmpty          = errors.New("date must not be empty")
	errDateInvalidFormat  = errors.New("date must be in MM-YYYY format")
	errDateMultipleValues = errors.New("date must be specified once")
	errMonthOutOfRange    = errors.New("month must be between 1 and 12")
	errYearOutOfRange     = errors.New("year must be greater than zero")
)

var (
	errServiceNameMultipleValues = errors.New("service_name must be specified once")
	errServiceNameEmpty          = errors.New("service name is empty")
)

var (
	errUserIDEmpty          = errors.New("user_id must not be empty")
	errUserIDMultipleValues = errors.New("user_id must be specified once")
	errServiceNameRequired  = errors.New("service_name must not be empty")
	errPriceInvalid         = errors.New("price must be greater than zero")
	errStartDateRequired    = errors.New("start_date is required")
	errEndDateBeforeStart   = errors.New("end_date must not be before start_date")
)

type filterTotal struct {
	UserID      *string
	ServiceName *string
	From        *string
	To          *string
}

func (f *filterTotal) ToDomain() (domain.FilterTotal, error) {
	var userID *uuid.UUID
	if f.UserID != nil {
		parsedUserID, err := uuid.Parse(strings.TrimSpace(*f.UserID))
		if err != nil {
			return domain.FilterTotal{}, errInvalidUUID
		}
		userID = &parsedUserID
	}

	from, err := mapStringYearMonthToDomainPtr(f.From)
	if err != nil {
		return domain.FilterTotal{}, err
	}

	to, err := mapStringYearMonthToDomainPtr(f.To)
	if err != nil {
		return domain.FilterTotal{}, err
	}

	return domain.FilterTotal{
		UserID:      userID,
		ServiceName: f.ServiceName,
		From:        from,
		To:          to,
	}, nil
}

type subscriptionRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

func (s *subscriptionRequest) ToDomain() (domain.Subscription, error) {
	if strings.TrimSpace(s.ServiceName) == "" {
		return domain.Subscription{}, errServiceNameRequired
	}

	if s.Price <= 0 {
		return domain.Subscription{}, errPriceInvalid
	}

	trimmedUserID := strings.TrimSpace(s.UserID)
	if trimmedUserID == "" {
		return domain.Subscription{}, errUserIDEmpty
	}

	userID, err := uuid.Parse(trimmedUserID)
	if err != nil {
		return domain.Subscription{}, errInvalidUUID
	}

	startDate, err := parseYearMonth(s.StartDate)
	if err != nil {
		return domain.Subscription{}, err
	}

	endDate, err := mapStringYearMonthToDomainPtr(s.EndDate)
	if err != nil {
		return domain.Subscription{}, err
	}

	if endDate != nil && endDate.CountOfMonth() < startDate.CountOfMonth() {
		return domain.Subscription{}, errEndDateBeforeStart
	}

	return domain.Subscription{
		ServiceName: strings.TrimSpace(s.ServiceName),
		Price:       s.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}

type subscriptionResponce struct {
	ID          uuid.UUID `json:"subscription_id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date"`
}

func parseYearMonth(raw string) (domain.YearMonth, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return domain.YearMonth{}, errDateEmpty
	}

	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return domain.YearMonth{}, errDateInvalidFormat
	}

	month, err := strconv.Atoi(parts[0])
	if err != nil {
		return domain.YearMonth{}, errDateInvalidFormat
	}

	year, err := strconv.Atoi(parts[1])
	if err != nil {
		return domain.YearMonth{}, errDateInvalidFormat
	}

	if month < 1 || month > 12 {
		return domain.YearMonth{}, errMonthOutOfRange
	}

	if year <= 0 {
		return domain.YearMonth{}, errYearOutOfRange
	}

	return domain.YearMonth{
		Year:  year,
		Month: time.Month(month),
	}, nil
}

func formatYearMonth(date domain.YearMonth) string {
	return fmt.Sprintf("%02d-%04d", date.Month, date.Year)
}

func formatYearMonthPtr(date *domain.YearMonth) *string {
	if date == nil {
		return nil
	}

	formatted := formatYearMonth(*date)
	return &formatted
}

func mapStringYearMonthToDomainPtr(value *string) (*domain.YearMonth, error) {
	if value == nil {
		return nil, nil
	}

	parsed, err := parseYearMonth(*value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

type CreateResponce struct {
	SubID uuid.UUID `json:"subscription_id"`
}

type totalResponce struct {
	Total int `json:"total"`
}

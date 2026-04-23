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
	ErrInvalidUUID             error = errors.New("invalid uuid")
	errNotFound                error = errors.New("not Found")
)

var (
	ErrDateEmpty          = errors.New("date must not be empty")
	ErrDateInvalidFormat  = errors.New("date must be in MM-YYYY format")
	ErrDateMultipleValues = errors.New("date must be specified once")
	ErrMonthOutOfRange    = errors.New("month must be between 1 and 12")
)

var (
	ErrServiceNameMultipleValues = errors.New("service_name must be specified once")
	ErrServiceNameEmpty          = errors.New("service name is empty")
)

var (
	ErrUserIDEmpty          = errors.New("user_id must not be empty")
	ErrUserIDMultipleValues = errors.New("user_id must be specified once")
	ErrServiceNameRequired  = errors.New("service_name must not be empty")
	ErrPriceInvalid         = errors.New("price must be greater than zero")
	ErrStartDateRequired    = errors.New("start_date is required")
	ErrEndDateBeforeStart   = errors.New("end_date must not be before start_date")
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
			return domain.FilterTotal{}, ErrInvalidUUID
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
	userID, err := uuid.Parse(strings.TrimSpace(s.UserID))
	if err != nil {
		return domain.Subscription{}, ErrInvalidUUID
	}

	startDate, err := parseYearMonth(s.StartDate)
	if err != nil {
		return domain.Subscription{}, err
	}

	endDate, err := mapStringYearMonthToDomainPtr(s.EndDate)
	if err != nil {
		return domain.Subscription{}, err
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

func (s *subscriptionRequest) Validate() error {
	if strings.TrimSpace(s.ServiceName) == "" {
		return ErrServiceNameRequired
	}

	if s.Price <= 0 {
		return ErrPriceInvalid
	}

	if strings.TrimSpace(s.UserID) == "" {
		return ErrInvalidUUID
	}

	if strings.TrimSpace(s.StartDate) == "" {
		return ErrStartDateRequired
	}

	startDate, err := parseYearMonth(s.StartDate)
	if err != nil {
		return err
	}

	if _, err := uuid.Parse(strings.TrimSpace(s.UserID)); err != nil {
		return ErrInvalidUUID
	}

	endDate, err := mapStringYearMonthToDomainPtr(s.EndDate)
	if err != nil {
		return err
	}

	if endDate != nil && endDate.CountOfMonth() < startDate.CountOfMonth() {
		return ErrEndDateBeforeStart
	}

	return nil
}

func parseYearMonth(raw string) (domain.YearMonth, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return domain.YearMonth{}, ErrDateEmpty
	}

	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return domain.YearMonth{}, ErrDateInvalidFormat
	}

	month, err := strconv.Atoi(parts[0])
	if err != nil {
		return domain.YearMonth{}, ErrDateInvalidFormat
	}

	year, err := strconv.Atoi(parts[1])
	if err != nil {
		return domain.YearMonth{}, ErrDateInvalidFormat
	}

	if month < 1 || month > 12 {
		return domain.YearMonth{}, ErrMonthOutOfRange
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

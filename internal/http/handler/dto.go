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
	userID               = "user_id"
	serviceName          = "service_name"
	fromDate             = "from"
	toDate               = "to"
	limit                = "limit"
	offset               = "offset"
	defaultLimit  uint64 = 100
	maxLimit      uint64 = 1000
	defaultOffset uint64 = 0
)

var (
	errInvalidRequestBody error = errors.New("invalid request body")
	errGetData            error = errors.New("internal error")
	errInvalidUUID        error = errors.New("invalid uuid")
	errNotFound           error = errors.New("not Found")
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
	errUserIDEmpty           = errors.New("user_id must not be empty")
	errUserIDMultipleValues  = errors.New("user_id must be specified once")
	errServiceNameRequired   = errors.New("service_name must not be empty")
	errPriceInvalid          = errors.New("price must not be negative")
	errStartDateRequired     = errors.New("start_date is required")
	errUpdateEmpty           = errors.New("update payload must contain at least one field")
	errEndDateUpdateConflict = errors.New("end_date and clear_end_date cannot be used together")
)

var (
	errLimitEmpty           = errors.New("limit must not be empty")
	errLimitMultipleValues  = errors.New("limit must be specified once")
	errLimitInvalid         = errors.New("limit must be a positive integer")
	errLimitTooLarge        = errors.New("limit exceeds maximum")
	errOffsetEmpty          = errors.New("offset must not be empty")
	errOffsetMultipleValues = errors.New("offset must be specified once")
	errOffsetInvalid        = errors.New("offset must be a non-negative integer")
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

type SubscriptionRequest struct {
	ServiceName string  `json:"service_name" example:"Yandex Plus"`
	Price       int     `json:"price" example:"0" minimum:"0"`
	UserID      string  `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba" format:"uuid"`
	StartDate   string  `json:"start_date" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"09-2025"`
}

func (s *SubscriptionRequest) ToDomain() (domain.Subscription, error) {
	if strings.TrimSpace(s.ServiceName) == "" {
		return domain.Subscription{}, errServiceNameRequired
	}

	if s.Price < 0 {
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
		return domain.Subscription{}, domain.ErrEndDateBeforeStart
	}

	return domain.Subscription{
		ServiceName: strings.TrimSpace(s.ServiceName),
		Price:       s.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}

type SubscriptionUpdateRequest struct {
	ServiceName  *string `json:"service_name,omitempty" example:"Yandex Plus"`
	Price        *int    `json:"price,omitempty" example:"500" minimum:"0"`
	UserID       *string `json:"user_id,omitempty" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba" format:"uuid"`
	StartDate    *string `json:"start_date,omitempty" example:"07-2025"`
	EndDate      *string `json:"end_date,omitempty" example:"09-2025"`
	ClearEndDate *bool   `json:"clear_end_date,omitempty" example:"true"`
}

func (s *SubscriptionUpdateRequest) ToDomain() (domain.SubscriptionUpdate, error) {
	var serviceName *string
	if s.ServiceName != nil {
		trimmedServiceName := strings.TrimSpace(*s.ServiceName)
		if trimmedServiceName == "" {
			return domain.SubscriptionUpdate{}, errServiceNameRequired
		}
		serviceName = &trimmedServiceName
	}

	if s.Price != nil && *s.Price < 0 {
		return domain.SubscriptionUpdate{}, errPriceInvalid
	}

	var userID *uuid.UUID
	if s.UserID != nil {
		trimmedUserID := strings.TrimSpace(*s.UserID)
		if trimmedUserID == "" {
			return domain.SubscriptionUpdate{}, errUserIDEmpty
		}

		parsedUserID, err := uuid.Parse(trimmedUserID)
		if err != nil {
			return domain.SubscriptionUpdate{}, errInvalidUUID
		}
		userID = &parsedUserID
	}

	startDate, err := mapStringYearMonthToDomainPtr(s.StartDate)
	if err != nil {
		return domain.SubscriptionUpdate{}, err
	}

	endDate, err := mapStringYearMonthToDomainPtr(s.EndDate)
	if err != nil {
		return domain.SubscriptionUpdate{}, err
	}

	clearEndDate := s.ClearEndDate != nil && *s.ClearEndDate
	if endDate != nil && clearEndDate {
		return domain.SubscriptionUpdate{}, errEndDateUpdateConflict
	}

	if endDate != nil && startDate != nil && endDate.CountOfMonth() < startDate.CountOfMonth() {
		return domain.SubscriptionUpdate{}, domain.ErrEndDateBeforeStart
	}

	update := domain.SubscriptionUpdate{
		ServiceName: serviceName,
		Price:       s.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate: domain.EndDateUpdate{
			Value: endDate,
			Clear: clearEndDate,
		},
	}

	if !update.HasChanges() {
		return domain.SubscriptionUpdate{}, errUpdateEmpty
	}

	return update, nil
}

type SubscriptionResponse struct {
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

type CreateResponse struct {
	SubID uuid.UUID `json:"subscription_id"`
}

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

type TotalResponse struct {
	Total int `json:"total"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

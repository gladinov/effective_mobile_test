package handler

import (
	"errors"
	"time"

	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/google/uuid"
)

const (
	userID      = "user_id"
	serviceName = "service_name"
	fromYear    = "from_year"
	fromMonth   = "from_month"
	toYear      = "to_year"
	toMonth     = "to_month"
)

var (
	errInvalidRequestBody      error = errors.New("invalid request body")
	errInvalidUserIDQueryParam error = errors.New("invalid userID query param")
	errGetData                 error = errors.New("internal error")
	ErrInvalidUUID             error = errors.New("invalid uuid")
	errNotFound                error = errors.New("not Found")
)

var (
	ErrYearRequired        = errors.New("year is required when month is provided")
	ErrMonthRequired       = errors.New("month is required when year is provided")
	ErrYearEmpty           = errors.New("year is empty")
	ErrMonthEmpty          = errors.New("month is empty")
	ErrYearInvalid         = errors.New("year must be an integer")
	ErrMonthInvalid        = errors.New("month must be an integer")
	ErrMonthOutOfRange     = errors.New("month must be between 1 and 12")
	ErrYearMultipleValues  = errors.New("year must be specified once")
	ErrMonthMultipleValues = errors.New("month must be specified once")
)

var (
	ErrServiceNameMultipleValues = errors.New("service_name must be specified once")
	ErrServiceNameEmpty          = errors.New("service name is empty")
)

var (
	ErrUserIDEmpty          = errors.New("user_id must not be empty")
	ErrUserIDMultipleValues = errors.New("user_id must be specified once")
)

type filterTotal struct {
	UserID      *uuid.UUID
	ServiceName *string
	From        *yearMonth
	To          *yearMonth
}

func (f *filterTotal) ToDomain() domain.FilterTotal {
	return domain.FilterTotal{
		UserID:      f.UserID,
		ServiceName: f.ServiceName,
		From:        mapYearMonthPtrToDomainPtr(f.From),
		To:          mapYearMonthPtrToDomainPtr(f.To),
	}
}

type subscriptionRequest struct {
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   yearMonth  `json:"start_date"`
	EndDate     *yearMonth `json:"end_date"`
}

func (s *subscriptionRequest) ToDomain() domain.Subscription {
	return domain.Subscription{
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID,
		StartDate:   s.StartDate.ToDomain(),
		EndDate:     mapYearMonthPtrToDomain(s.EndDate),
	}
}

type subscriptionResponce struct {
	ID          uuid.UUID  `json:"subscription_id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   yearMonth  `json:"start_date"`
	EndDate     *yearMonth `json:"end_date"`
}

type yearMonth struct {
	Year  int        `json:"year"`
	Month time.Month `json:"month"`
}

func (y *yearMonth) ToDomain() domain.YearMonth {
	return domain.YearMonth{
		Year:  y.Year,
		Month: y.Month,
	}
}

type CreateResponce struct {
	SubID uuid.UUID `json:"subscription_id"`
}

type totalResponce struct {
	Total int `json:"total"`
}

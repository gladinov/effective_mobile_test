package handler

import (
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

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateResponce struct {
	SubID uuid.UUID `json:"subscription_id"`
}

type totalResponce struct {
	Total int `json:"total"`
}

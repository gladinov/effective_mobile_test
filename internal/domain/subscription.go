package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrSubscriptionNotFound error = errors.New("not found subscription by this ID")

type Subscription struct {
	ID          uuid.UUID
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   YearMonth
	EndDate     *YearMonth
}

type SubscriptionUpdate struct {
	ServiceName *string
	Price       *int
	UserID      *uuid.UUID
	StartDate   *YearMonth
	EndDate     EndDateUpdate
}

func (s SubscriptionUpdate) HasChanges() bool {
	return s.ServiceName != nil ||
		s.Price != nil ||
		s.UserID != nil ||
		s.StartDate != nil ||
		s.EndDate.IsSet()
}

type EndDateUpdate struct {
	Value *YearMonth
	Clear bool
}

func (e EndDateUpdate) IsSet() bool {
	return e.Value != nil || e.Clear
}

type YearMonth struct {
	Year  int
	Month time.Month
}

func (y *YearMonth) CountOfMonth() int {
	return y.Year*12 + int(y.Month)
}

func (to *YearMonth) Sub(from YearMonth) int {
	fromMonth := from.Year*12 + int(from.Month)
	toMonth := to.Year*12 + int(to.Month)
	// Указываем конечный месяц включительно
	return toMonth - fromMonth + 1
}

type FilterTotal struct {
	UserID      *uuid.UUID
	ServiceName *string
	From        *YearMonth
	To          *YearMonth
}

type Pagination struct {
	Limit  uint64
	Offset uint64
}

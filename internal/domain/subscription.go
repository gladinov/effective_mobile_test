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

type YearMonth struct {
	Year  int
	Month time.Month
}

func (y *YearMonth) CountOfMonth() int {
	return y.Year*12 + int(y.Month)
}

// TODO: Должен ли я включать первый и последний месяц.
// Даже вопрос в том. "до" или "по" последний месяц подписка?
func (to *YearMonth) Sub(from YearMonth) int {
	fromMonth := from.Year*12 + int(from.Month)
	toMonth := to.Year*12 + int(to.Month)
	return toMonth - fromMonth
}

type FilterTotal struct {
	UserID      *uuid.UUID
	ServiceName *string
	From        *YearMonth
	To          *YearMonth
}

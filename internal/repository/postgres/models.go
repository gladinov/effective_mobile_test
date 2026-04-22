package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrSubscriptionNotFound error = errors.New("not found subscription by this ID")

const (
	subscriptionTable = "subscriptions"
	colID             = "id"
	colServiceName    = "service_name"
	colPrice          = "price"
	colUserID         = "user_id"
	colStartDate      = "start_date"
	colEndDate        = "end_date"
)

type subscriptionRow struct {
	ID          uuid.UUID  `db:"id"`
	ServiceName string     `db:"service_name"`
	Price       int        `db:"price"`
	UserID      uuid.UUID  `db:"user_id"`
	StartDate   time.Time  `db:"start_date"`
	EndDate     *time.Time `db:"end_date"`
}

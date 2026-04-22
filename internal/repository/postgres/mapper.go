package postgres

import (
	"time"

	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
)

func startDateFromSql(yearMonth time.Time) domain.YearMonth {
	year := yearMonth.Year()
	month := yearMonth.Month()
	return domain.YearMonth{Year: year, Month: month}
}

func endDateFromSql(yearMonth *time.Time) *domain.YearMonth {
	if yearMonth == nil {
		return nil
	}
	year := yearMonth.Year()
	month := yearMonth.Month()

	return &domain.YearMonth{Year: year, Month: month}
}

func mapYearMonthToSql(m domain.YearMonth) time.Time {
	return time.Date(m.Year, m.Month, 1, 0, 0, 0, 0, time.UTC)
}

// TODO: Корректно ли это?
func mapPtrYearMonthToSql(m *domain.YearMonth) *time.Time {
	if m != nil {
		date := time.Date(m.Year, m.Month, 1, 0, 0, 0, 0, time.UTC)
		return &date
	}
	return nil
}

func (s *subscriptionRow) ToDomain() domain.Subscription {
	return domain.Subscription{
		ID:          s.ID,
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID,
		StartDate:   startDateFromSql(s.StartDate),
		EndDate:     endDateFromSql(s.EndDate),
	}
}

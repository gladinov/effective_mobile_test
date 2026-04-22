package service

import (
	"time"

	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
)

func (s *Service) pricePaidInPeriod(filter domain.FilterTotal, sub domain.Subscription) int {
	return s.countOfMonthSubInPeriod(filter, sub) * sub.Price
}

func (s *Service) countOfMonthSubInPeriod(filter domain.FilterTotal, sub domain.Subscription) int {
	startPayMonth := s.setStartPayMonth(filter, sub)
	endPayMonth := s.setEndPayMonth(filter, sub)
	dur := endPayMonth.Sub(startPayMonth)
	return dur
}

func (s *Service) setStartPayMonth(filter domain.FilterTotal, sub domain.Subscription) domain.YearMonth {
	var startPayMonth domain.YearMonth
	if filter.From != nil {
		startPayMonth = maxYearMonth(sub.StartDate, *filter.From)
	}
	startPayMonth = sub.StartDate
	return startPayMonth
}

func (s *Service) setEndPayMonth(filter domain.FilterTotal, sub domain.Subscription) domain.YearMonth {
	var endPayMonth domain.YearMonth
	switch {
	case filter.To != nil && sub.EndDate != nil:
		endPayMonth = minYearMonth(*sub.EndDate, *filter.To)
	case filter.To != nil:
		endPayMonth = *filter.To
	case sub.EndDate != nil:
		endPayMonth = *sub.EndDate
	case filter.To == nil && sub.EndDate == nil:
		endPayMonth = todayInYearMonth(s.now())
	}
	return endPayMonth
}

func maxYearMonth(first domain.YearMonth, second domain.YearMonth) domain.YearMonth {
	if first.CountOfMonth() >= second.CountOfMonth() {
		return first
	}
	return second
}

func minYearMonth(first domain.YearMonth, second domain.YearMonth) domain.YearMonth {
	if first.CountOfMonth() <= second.CountOfMonth() {
		return first
	}
	return second
}

func todayInYearMonth(t time.Time) domain.YearMonth {
	return domain.YearMonth{
		Year:  t.Year(),
		Month: t.Month(),
	}
}

package service

import (
	"testing"
	"time"

	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestServiceSetStartPayMonth(t *testing.T) {
	t.Parallel()

	july2025 := domain.YearMonth{Year: 2025, Month: time.July}
	august2025 := domain.YearMonth{Year: 2025, Month: time.August}

	tests := []struct {
		name   string
		filter domain.FilterTotal
		sub    domain.Subscription
		want   domain.YearMonth
	}{
		{
			name:   "without from filter uses subscription start",
			filter: domain.FilterTotal{},
			sub: domain.Subscription{
				StartDate: july2025,
			},
			want: july2025,
		},
		{
			name: "with later from filter uses filter from",
			filter: domain.FilterTotal{
				From: &august2025,
			},
			sub: domain.Subscription{
				StartDate: july2025,
			},
			want: august2025,
		},
		{
			name: "with earlier from filter keeps subscription start",
			filter: domain.FilterTotal{
				From: &july2025,
			},
			sub: domain.Subscription{
				StartDate: august2025,
			},
			want: august2025,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := setStartPayMonth(tt.filter, tt.sub)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServiceSetEndPayMonth(t *testing.T) {
	t.Parallel()

	july2025 := domain.YearMonth{Year: 2025, Month: time.July}
	august2025 := domain.YearMonth{Year: 2025, Month: time.August}
	september2025 := domain.YearMonth{Year: 2025, Month: time.September}
	october2025 := domain.YearMonth{Year: 2025, Month: time.October}

	now := func() time.Time {
		return time.Date(2025, time.November, 17, 0, 0, 0, 0, time.UTC)
	}

	tests := []struct {
		name   string
		filter domain.FilterTotal
		sub    domain.Subscription
		want   domain.YearMonth
	}{
		{
			name: "with subscription end and filter to uses earliest month",
			filter: domain.FilterTotal{
				To: &september2025,
			},
			sub: domain.Subscription{
				EndDate: yearMonthHelperPtr(2025, time.October),
			},
			want: september2025,
		},
		{
			name: "with only filter to uses filter to",
			filter: domain.FilterTotal{
				To: &september2025,
			},
			sub:  domain.Subscription{},
			want: september2025,
		},
		{
			name:   "with only subscription end uses subscription end",
			filter: domain.FilterTotal{},
			sub: domain.Subscription{
				EndDate: yearMonthHelperPtr(2025, time.October),
			},
			want: october2025,
		},
		{
			name:   "open ended subscription without filter to uses now",
			filter: domain.FilterTotal{},
			sub:    domain.Subscription{StartDate: july2025},
			want:   domain.YearMonth{Year: 2025, Month: time.November},
		},
		{
			name: "with earlier subscription end than filter to uses subscription end",
			filter: domain.FilterTotal{
				To: &october2025,
			},
			sub: domain.Subscription{
				EndDate: &august2025,
			},
			want: august2025,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := setEndPayMonth(now, tt.filter, tt.sub)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServiceCountOfMonthSubInPeriod(t *testing.T) {
	t.Parallel()

	now := func() time.Time {
		return time.Date(2025, time.September, 15, 0, 0, 0, 0, time.UTC)
	}

	july2025 := domain.YearMonth{Year: 2025, Month: time.July}
	august2025 := domain.YearMonth{Year: 2025, Month: time.August}
	september2025 := domain.YearMonth{Year: 2025, Month: time.September}
	october2025 := domain.YearMonth{Year: 2025, Month: time.October}

	tests := []struct {
		name   string
		filter domain.FilterTotal
		sub    domain.Subscription
		want   int
	}{
		{
			name:   "counts inclusive months",
			filter: domain.FilterTotal{},
			sub: domain.Subscription{
				StartDate: july2025,
				EndDate:   &september2025,
			},
			want: 3,
		},
		{
			name: "applies from and to bounds",
			filter: domain.FilterTotal{
				From: &august2025,
				To:   &september2025,
			},
			sub: domain.Subscription{
				StartDate: july2025,
				EndDate:   &october2025,
			},
			want: 2,
		},
		{
			name:   "uses now for open ended subscription",
			filter: domain.FilterTotal{},
			sub: domain.Subscription{
				StartDate: july2025,
				EndDate:   nil,
			},
			want: 3,
		},
		{
			name: "returns zero when effective end is before effective start",
			filter: domain.FilterTotal{
				From: &october2025,
				To:   &august2025,
			},
			sub: domain.Subscription{
				StartDate: july2025,
				EndDate:   &september2025,
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := countOfMonthSubInPeriod(now, tt.filter, tt.sub)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServicePricePaidInPeriod(t *testing.T) {
	t.Parallel()

	now := func() time.Time {
		return time.Date(2025, time.September, 15, 0, 0, 0, 0, time.UTC)
	}

	got := pricePaidInPeriod(now, domain.FilterTotal{}, domain.Subscription{
		Price:     400,
		StartDate: domain.YearMonth{Year: 2025, Month: time.July},
		EndDate:   yearMonthHelperPtr(2025, time.September),
	})

	require.Equal(t, 1200, got)
}

func TestMaxYearMonth(t *testing.T) {
	t.Parallel()

	first := domain.YearMonth{Year: 2025, Month: time.July}
	second := domain.YearMonth{Year: 2025, Month: time.September}

	require.Equal(t, second, maxYearMonth(first, second))
	require.Equal(t, second, maxYearMonth(second, first))
}

func TestMinYearMonth(t *testing.T) {
	t.Parallel()

	first := domain.YearMonth{Year: 2025, Month: time.July}
	second := domain.YearMonth{Year: 2025, Month: time.September}

	require.Equal(t, first, minYearMonth(first, second))
	require.Equal(t, first, minYearMonth(second, first))
}

func TestTodayInYearMonth(t *testing.T) {
	t.Parallel()

	got := todayInYearMonth(time.Date(2026, time.February, 24, 15, 30, 0, 0, time.UTC))
	require.Equal(t, domain.YearMonth{Year: 2026, Month: time.February}, got)
}

func yearMonthHelperPtr(year int, month time.Month) *domain.YearMonth {
	return &domain.YearMonth{
		Year:  year,
		Month: month,
	}
}

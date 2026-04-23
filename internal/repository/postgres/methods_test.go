package postgres

import (
	"testing"
	"time"

	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestApplyFilters(t *testing.T) {
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	serviceName := "music"
	from := domain.YearMonth{Year: 2024, Month: time.January}
	to := domain.YearMonth{Year: 2024, Month: time.March}

	tests := []struct {
		name     string
		filter   domain.FilterTotal
		wantSQL  string
		wantArgs []any
	}{
		{
			name:    "without filters",
			filter:  domain.FilterTotal{},
			wantSQL: "SELECT id FROM subscriptions",
			wantArgs: nil,
		},
		{
			name: "with user and service filters",
			filter: domain.FilterTotal{
				UserID:      &userID,
				ServiceName: &serviceName,
			},
			wantSQL:  "SELECT id FROM subscriptions WHERE user_id = $1 AND service_name = $2",
			wantArgs: []any{userID.String(), serviceName},
		},
		{
			name: "with from and to filters",
			filter: domain.FilterTotal{
				From: &from,
				To:   &to,
			},
			wantSQL: "SELECT id FROM subscriptions WHERE start_date <= $1 AND (end_date IS NULL OR end_date >= $2)",
			wantArgs: []any{
				time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "with from filter only",
			filter: domain.FilterTotal{
				From: &from,
			},
			wantSQL: "SELECT id FROM subscriptions WHERE (end_date IS NULL OR end_date >= $1)",
			wantArgs: []any{
				time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "with to filter only",
			filter: domain.FilterTotal{
				To: &to,
			},
			wantSQL: "SELECT id FROM subscriptions WHERE start_date <= $1",
			wantArgs: []any{
				time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := psql.Select(colID).From(subscriptionTable)

			gotSQL, gotArgs, err := applyFilters(tt.filter, query).ToSql()
			require.NoError(t, err)
			require.Equal(t, tt.wantSQL, gotSQL)
			require.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}

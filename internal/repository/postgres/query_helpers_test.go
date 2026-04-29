package postgres

import (
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
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
			name:     "without filters",
			filter:   domain.FilterTotal{},
			wantSQL:  "SELECT id FROM subscriptions",
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

func TestApplyPagination(t *testing.T) {
	pagination := domain.Pagination{Limit: 10, Offset: 20}

	gotSQL, gotArgs, err := psql.
		Select(colID).
		From(subscriptionTable).
		OrderBy(colID).
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		ToSql()

	require.NoError(t, err)
	require.Equal(t, "SELECT id FROM subscriptions ORDER BY id LIMIT 10 OFFSET 20", gotSQL)
	require.Nil(t, gotArgs)
}

func TestApplySubscriptionUpdate(t *testing.T) {
	serviceName := "Netflix"
	price := 1200
	userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
	startDate := domain.YearMonth{Year: 2026, Month: time.January}
	endDate := domain.YearMonth{Year: 2026, Month: time.March}

	update := domain.SubscriptionUpdate{
		ServiceName: &serviceName,
		Price:       &price,
		UserID:      &userID,
		StartDate:   &startDate,
		EndDate: domain.EndDateUpdate{
			Value: &endDate,
		},
	}

	gotSQL, gotArgs, err := applySubscriptionUpdate(update, psql.Update(subscriptionTable)).
		Where(sq.Eq{colID: uuid.Nil}).
		ToSql()

	require.NoError(t, err)
	require.Equal(t, "UPDATE subscriptions SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5 WHERE id = $6", gotSQL)
	require.Equal(t, []any{
		serviceName,
		price,
		userID,
		time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC),
		uuid.Nil.String(),
	}, gotArgs)
}

func TestApplySubscriptionUpdate_ClearEndDate(t *testing.T) {
	gotSQL, gotArgs, err := applySubscriptionUpdate(
		domain.SubscriptionUpdate{EndDate: domain.EndDateUpdate{Clear: true}},
		psql.Update(subscriptionTable),
	).
		Where(sq.Eq{colID: uuid.Nil}).
		ToSql()

	require.NoError(t, err)
	require.Equal(t, "UPDATE subscriptions SET end_date = $1 WHERE id = $2", gotSQL)
	require.Equal(t, []any{nil, uuid.Nil.String()}, gotArgs)
}

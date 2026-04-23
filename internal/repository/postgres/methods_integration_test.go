//go:build integration
// +build integration

package postgres

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/stretchr/testify/require"
)

const createSubscriptionsMigrationPath = "../../../deployments/migrations/postgreSQL/0001_create_subscriptions_table.up.sql"

func TestStorageCreateIntegration(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool := newTestPostgresPool(ctx, t)
	storage := NewStorage(pool, 5*time.Second)

	userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
	endDate := domain.YearMonth{Year: 2025, Month: time.September}
	sub := domain.Subscription{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      userID,
		StartDate:   domain.YearMonth{Year: 2025, Month: time.July},
		EndDate:     &endDate,
	}

	createdID, err := storage.Create(ctx, sub)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, createdID)

	var got subscriptionRow
	err = pool.QueryRow(
		ctx,
		`SELECT id, service_name, price, user_id, start_date, end_date
		 FROM subscriptions
		 WHERE id = $1`,
		createdID,
	).Scan(
		&got.ID,
		&got.ServiceName,
		&got.Price,
		&got.UserID,
		&got.StartDate,
		&got.EndDate,
	)
	require.NoError(t, err)

	require.Equal(t, createdID, got.ID)
	require.Equal(t, sub.ServiceName, got.ServiceName)
	require.Equal(t, sub.Price, got.Price)
	require.Equal(t, sub.UserID, got.UserID)
	require.Equal(t, mapYearMonthToSql(sub.StartDate), got.StartDate)
	require.Equal(t, mapPtrYearMonthToSql(sub.EndDate), got.EndDate)
}

func TestStorageGetByIDIntegration(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool := newTestPostgresPool(ctx, t)
	storage := NewStorage(pool, 5*time.Second)

	subID := uuid.MustParse("5dc2f79b-8606-4a8f-88fe-d0d2289de8a2")
	userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
	startDate := time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, time.September, 1, 0, 0, 0, 0, time.UTC)

	_, err := pool.Exec(
		ctx,
		`INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		subID,
		"Yandex Plus",
		400,
		userID,
		startDate,
		&endDate,
	)
	require.NoError(t, err)

	got, err := storage.GetByID(ctx, subID)
	require.NoError(t, err)

	require.Equal(t, domain.Subscription{
		ID:          subID,
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      userID,
		StartDate: domain.YearMonth{
			Year:  2025,
			Month: time.July,
		},
		EndDate: &domain.YearMonth{
			Year:  2025,
			Month: time.September,
		},
	}, got)
}

func TestStorageGetByIDIntegration_NotFound(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool := newTestPostgresPool(ctx, t)
	storage := NewStorage(pool, 5*time.Second)

	missingID := uuid.MustParse("97cbd2bb-f735-4668-a49d-482da1a165ae")

	_, err := storage.GetByID(ctx, missingID)
	require.ErrorIs(t, err, ErrSubscriptionNotFound)
}

func TestStorageUpdateByIDIntegration(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool := newTestPostgresPool(ctx, t)
	storage := NewStorage(pool, 5*time.Second)

	subID := uuid.MustParse("4bc65e58-1a0f-4f49-8a6b-79cf9f3f857f")
	seedSubscription(t, ctx, pool, domain.Subscription{
		ID:          subID,
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   domain.YearMonth{Year: 2025, Month: time.July},
		EndDate:     yearMonthPtr(2025, time.September),
	})

	updated := domain.Subscription{
		ServiceName: "Netflix",
		Price:       1200,
		UserID:      uuid.MustParse("b7582d0c-aad8-40d2-8362-265a18173964"),
		StartDate:   domain.YearMonth{Year: 2026, Month: time.January},
		EndDate:     yearMonthPtr(2026, time.March),
	}

	err := storage.UpdateByID(ctx, subID, updated)
	require.NoError(t, err)

	got, err := storage.GetByID(ctx, subID)
	require.NoError(t, err)
	require.Equal(t, domain.Subscription{
		ID:          subID,
		ServiceName: updated.ServiceName,
		Price:       updated.Price,
		UserID:      updated.UserID,
		StartDate:   updated.StartDate,
		EndDate:     updated.EndDate,
	}, got)
}

func TestStorageUpdateByIDIntegration_NotFound(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool := newTestPostgresPool(ctx, t)
	storage := NewStorage(pool, 5*time.Second)

	err := storage.UpdateByID(ctx, uuid.MustParse("87198191-209d-44bf-ba0d-2d6fc0f2eb1a"), domain.Subscription{
		ServiceName: "Spotify",
		Price:       300,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   domain.YearMonth{Year: 2025, Month: time.May},
	})
	require.ErrorIs(t, err, ErrSubscriptionNotFound)
}

func TestStorageDeleteByIDIntegration(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool := newTestPostgresPool(ctx, t)
	storage := NewStorage(pool, 5*time.Second)

	subID := uuid.MustParse("f29572c4-d337-45d6-bc15-1909ca6e06b6")
	seedSubscription(t, ctx, pool, domain.Subscription{
		ID:          subID,
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   domain.YearMonth{Year: 2025, Month: time.July},
		EndDate:     yearMonthPtr(2025, time.September),
	})

	err := storage.DeleteByID(ctx, subID)
	require.NoError(t, err)

	_, err = storage.GetByID(ctx, subID)
	require.ErrorIs(t, err, ErrSubscriptionNotFound)
}

func TestStorageDeleteByIDIntegration_NotFound(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool := newTestPostgresPool(ctx, t)
	storage := NewStorage(pool, 5*time.Second)

	err := storage.DeleteByID(ctx, uuid.MustParse("e5b22ff1-6c0f-4138-b3a2-f86adbb0f764"))
	require.ErrorIs(t, err, ErrSubscriptionNotFound)
}

func TestStorageListIntegration(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool := newTestPostgresPool(ctx, t)
	storage := NewStorage(pool, 5*time.Second)

	first := domain.Subscription{
		ID:          uuid.MustParse("18e60ce7-ae34-4628-981d-5508c53c1244"),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   domain.YearMonth{Year: 2025, Month: time.July},
		EndDate:     yearMonthPtr(2025, time.September),
	}
	second := domain.Subscription{
		ID:          uuid.MustParse("d0a53149-7910-4ba2-b5e2-5c6fd38bef99"),
		ServiceName: "Netflix",
		Price:       1200,
		UserID:      uuid.MustParse("b7582d0c-aad8-40d2-8362-265a18173964"),
		StartDate:   domain.YearMonth{Year: 2026, Month: time.January},
		EndDate:     nil,
	}

	seedSubscription(t, ctx, pool, first)
	seedSubscription(t, ctx, pool, second)

	got, err := storage.List(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, []domain.Subscription{first, second}, got)
}

func TestStorageGetFiltredSubsIntegration(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool := newTestPostgresPool(ctx, t)
	storage := NewStorage(pool, 5*time.Second)

	targetUserID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
	match := domain.Subscription{
		ID:          uuid.MustParse("2ccb4d14-7ad9-4a64-bf8c-05bff5a79113"),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      targetUserID,
		StartDate:   domain.YearMonth{Year: 2025, Month: time.July},
		EndDate:     yearMonthPtr(2025, time.September),
	}
	openEndedMatch := domain.Subscription{
		ID:          uuid.MustParse("9880b581-bf5e-4776-b15d-9f4949dbb8f4"),
		ServiceName: "Yandex Plus",
		Price:       500,
		UserID:      targetUserID,
		StartDate:   domain.YearMonth{Year: 2025, Month: time.August},
		EndDate:     nil,
	}
	otherService := domain.Subscription{
		ID:          uuid.MustParse("a2164efb-ae53-4565-ad95-e37d20e9f369"),
		ServiceName: "Netflix",
		Price:       1000,
		UserID:      targetUserID,
		StartDate:   domain.YearMonth{Year: 2025, Month: time.July},
		EndDate:     yearMonthPtr(2025, time.December),
	}
	otherUser := domain.Subscription{
		ID:          uuid.MustParse("6b5ddb63-6314-440c-8ebe-6f779f3307b3"),
		ServiceName: "Yandex Plus",
		Price:       700,
		UserID:      uuid.MustParse("b7582d0c-aad8-40d2-8362-265a18173964"),
		StartDate:   domain.YearMonth{Year: 2025, Month: time.July},
		EndDate:     yearMonthPtr(2025, time.September),
	}

	seedSubscription(t, ctx, pool, match)
	seedSubscription(t, ctx, pool, openEndedMatch)
	seedSubscription(t, ctx, pool, otherService)
	seedSubscription(t, ctx, pool, otherUser)

	from := domain.YearMonth{Year: 2025, Month: time.August}
	to := domain.YearMonth{Year: 2025, Month: time.September}

	got, err := storage.GetFiltredSubs(ctx, domain.FilterTotal{
		UserID:      &targetUserID,
		ServiceName: stringPtr("Yandex Plus"),
		From:        &from,
		To:          &to,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []domain.Subscription{match, openEndedMatch}, got)
}

func newTestPostgresPool(ctx context.Context, t *testing.T) *pgxpool.Pool {
	t.Helper()

	container, err := tcpostgres.Run(
		ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("subscriptions_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		if shouldSkipTestcontainers(err) {
			t.Skipf("docker is unavailable for integration test: %v", err)
		}
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(context.Background()))
	})

	connString, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)

	t.Cleanup(pool.Close)

	require.NoError(t, pool.Ping(ctx))
	_, err = pool.Exec(ctx, mustReadCreateSubscriptionMigration(t))
	require.NoError(t, err)

	return pool
}

func seedSubscription(t *testing.T, ctx context.Context, pool *pgxpool.Pool, sub domain.Subscription) {
	t.Helper()

	_, err := pool.Exec(
		ctx,
		`INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		mapYearMonthToSql(sub.StartDate),
		mapPtrYearMonthToSql(sub.EndDate),
	)
	require.NoError(t, err)
}

func mustReadCreateSubscriptionMigration(t *testing.T) string {
	t.Helper()

	sqlBytes, err := os.ReadFile(filepath.Clean(createSubscriptionsMigrationPath))
	require.NoError(t, err)

	return string(sqlBytes)
}

func shouldSkipTestcontainers(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled) ||
		strings.Contains(strings.ToLower(err.Error()), "docker") ||
		strings.Contains(strings.ToLower(err.Error()), "container runtime") ||
		strings.Contains(strings.ToLower(err.Error()), "socket")
}

func yearMonthPtr(year int, month time.Month) *domain.YearMonth {
	return &domain.YearMonth{
		Year:  year,
		Month: month,
	}
}

func stringPtr(value string) *string {
	return &value
}

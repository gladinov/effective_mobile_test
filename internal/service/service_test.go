package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/gladinov/effective_mobile_test_assignment/internal/service/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceGetTotal(t *testing.T) {
	t.Parallel()

	july2025 := domain.YearMonth{Year: 2025, Month: time.July}
	august2025 := domain.YearMonth{Year: 2025, Month: time.August}
	september2025 := domain.YearMonth{Year: 2025, Month: time.September}
	october2025 := domain.YearMonth{Year: 2025, Month: time.October}
	november2025 := domain.YearMonth{Year: 2025, Month: time.November}

	tests := []struct {
		name      string
		now       time.Time
		filter    domain.FilterTotal
		subs      []domain.Subscription
		wantTotal int
	}{
		{
			name: "counts inclusive months without filters",
			now:  time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC),
			subs: []domain.Subscription{
				{
					ServiceName: "Yandex Plus",
					Price:       400,
					UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
					StartDate:   july2025,
					EndDate:     &september2025,
				},
			},
			wantTotal: 1200,
		},
		{
			name: "starts counting from filter from when it is later than subscription start",
			now:  time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC),
			filter: domain.FilterTotal{
				From: &august2025,
			},
			subs: []domain.Subscription{
				{
					ServiceName: "Yandex Plus",
					Price:       400,
					UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
					StartDate:   july2025,
					EndDate:     &september2025,
				},
			},
			wantTotal: 800,
		},
		{
			name: "ends counting at filter to when it is earlier than subscription end",
			now:  time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC),
			filter: domain.FilterTotal{
				To: &august2025,
			},
			subs: []domain.Subscription{
				{
					ServiceName: "Yandex Plus",
					Price:       400,
					UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
					StartDate:   july2025,
					EndDate:     &september2025,
				},
			},
			wantTotal: 800,
		},
		{
			name: "uses now for open ended subscription when filter to is absent",
			now:  time.Date(2025, time.September, 15, 0, 0, 0, 0, time.UTC),
			subs: []domain.Subscription{
				{
					ServiceName: "Netflix",
					Price:       1000,
					UserID:      uuid.MustParse("b7582d0c-aad8-40d2-8362-265a18173964"),
					StartDate:   july2025,
					EndDate:     nil,
				},
			},
			wantTotal: 3000,
		},
		{
			name: "uses both from and to for open ended subscription",
			now:  time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC),
			filter: domain.FilterTotal{
				From: &august2025,
				To:   &october2025,
			},
			subs: []domain.Subscription{
				{
					ServiceName: "Netflix",
					Price:       1000,
					UserID:      uuid.MustParse("b7582d0c-aad8-40d2-8362-265a18173964"),
					StartDate:   july2025,
					EndDate:     nil,
				},
			},
			wantTotal: 3000,
		},
		{
			name: "sums several subscriptions",
			now:  time.Date(2025, time.December, 1, 0, 0, 0, 0, time.UTC),
			subs: []domain.Subscription{
				{
					ServiceName: "Yandex Plus",
					Price:       400,
					UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
					StartDate:   july2025,
					EndDate:     &september2025,
				},
				{
					ServiceName: "Spotify",
					Price:       300,
					UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
					StartDate:   october2025,
					EndDate:     &november2025,
				},
			},
			wantTotal: 1800,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			storage := mocks.NewStorage(t)
			storage.On("GetFilteredSubs", mock.Anything, tt.filter).Return(tt.subs, nil).Once()

			svc := NewService(storage)
			svc.now = func() time.Time { return tt.now }

			got, err := svc.GetTotal(context.Background(), tt.filter)
			require.NoError(t, err)
			require.Equal(t, tt.wantTotal, got)
		})
	}
}

func TestServiceGetTotal_StorageError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("storage failed")
	storage := mocks.NewStorage(t)
	storage.On("GetFilteredSubs", mock.Anything, domain.FilterTotal{}).Return(([]domain.Subscription)(nil), wantErr).Once()

	svc := NewService(storage)

	_, err := svc.GetTotal(context.Background(), domain.FilterTotal{})
	require.Error(t, err)
	require.ErrorIs(t, err, wantErr)
	require.Contains(t, err.Error(), "get filtered subscriptions from storage")
}

func TestServiceGetByID(t *testing.T) {
	t.Parallel()

	subID := uuid.MustParse("5a012338-ae9e-45df-b657-c6b0a28d829a")
	wantSub := domain.Subscription{
		ID:          subID,
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		StartDate:   domain.YearMonth{Year: 2025, Month: time.July},
		EndDate:     yearMonthPtr(2025, time.September),
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		storage := mocks.NewStorage(t)
		storage.On("GetByID", mock.Anything, subID).Return(wantSub, nil).Once()

		svc := NewService(storage)

		got, err := svc.GetByID(context.Background(), subID)
		require.NoError(t, err)
		require.Equal(t, wantSub, got)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		storage := mocks.NewStorage(t)
		storage.On("GetByID", mock.Anything, subID).Return(domain.Subscription{}, domain.ErrSubscriptionNotFound).Once()

		svc := NewService(storage)

		_, err := svc.GetByID(context.Background(), subID)
		require.ErrorIs(t, err, domain.ErrSubscriptionNotFound)
	})

	t.Run("unexpected storage error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("storage failed")
		storage := mocks.NewStorage(t)
		storage.On("GetByID", mock.Anything, subID).Return(domain.Subscription{}, wantErr).Once()

		svc := NewService(storage)

		_, err := svc.GetByID(context.Background(), subID)
		require.Error(t, err)
		require.ErrorIs(t, err, wantErr)
		require.Contains(t, err.Error(), "get sub by id from storage")
	})
}

func TestServiceUpdateByID(t *testing.T) {
	t.Parallel()

	subID := uuid.MustParse("e994010b-55df-41d7-95b0-93a33f7011b4")
	sub := domain.Subscription{
		ServiceName: "Netflix",
		Price:       1200,
		UserID:      uuid.MustParse("b7582d0c-aad8-40d2-8362-265a18173964"),
		StartDate:   domain.YearMonth{Year: 2026, Month: time.January},
		EndDate:     yearMonthPtr(2026, time.March),
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		storage := mocks.NewStorage(t)
		storage.On("UpdateByID", mock.Anything, subID, sub).Return(nil).Once()

		svc := NewService(storage)

		err := svc.UpdateByID(context.Background(), subID, sub)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		storage := mocks.NewStorage(t)
		storage.On("UpdateByID", mock.Anything, subID, sub).Return(domain.ErrSubscriptionNotFound).Once()

		svc := NewService(storage)

		err := svc.UpdateByID(context.Background(), subID, sub)
		require.ErrorIs(t, err, domain.ErrSubscriptionNotFound)
	})

	t.Run("unexpected storage error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("storage failed")
		storage := mocks.NewStorage(t)
		storage.On("UpdateByID", mock.Anything, subID, sub).Return(wantErr).Once()

		svc := NewService(storage)

		err := svc.UpdateByID(context.Background(), subID, sub)
		require.Error(t, err)
		require.ErrorIs(t, err, wantErr)
		require.Contains(t, err.Error(), "update sub by id in storage")
	})
}

func TestServiceUpdatePartialByID(t *testing.T) {
	t.Parallel()

	subID := uuid.MustParse("e994010b-55df-41d7-95b0-93a33f7011b4")
	price := 500
	update := domain.SubscriptionUpdate{
		Price: &price,
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		storage := mocks.NewStorage(t)
		storage.On("UpdatePartialByID", mock.Anything, subID, update).Return(nil).Once()

		svc := NewService(storage)

		err := svc.UpdatePartialByID(context.Background(), subID, update)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		storage := mocks.NewStorage(t)
		storage.On("UpdatePartialByID", mock.Anything, subID, update).Return(domain.ErrSubscriptionNotFound).Once()

		svc := NewService(storage)

		err := svc.UpdatePartialByID(context.Background(), subID, update)
		require.ErrorIs(t, err, domain.ErrSubscriptionNotFound)
	})

	t.Run("unexpected storage error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("storage failed")
		storage := mocks.NewStorage(t)
		storage.On("UpdatePartialByID", mock.Anything, subID, update).Return(wantErr).Once()

		svc := NewService(storage)

		err := svc.UpdatePartialByID(context.Background(), subID, update)
		require.Error(t, err)
		require.ErrorIs(t, err, wantErr)
		require.Contains(t, err.Error(), "update partial sub by id in storage")
	})

	t.Run("returns error when end date is before existing start date", func(t *testing.T) {
		t.Parallel()

		endDate := domain.YearMonth{Year: 2025, Month: time.June}
		update := domain.SubscriptionUpdate{
			EndDate: domain.EndDateUpdate{Value: &endDate},
		}
		existingSub := domain.Subscription{
			ID:          subID,
			ServiceName: "Yandex Plus",
			Price:       400,
			UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
			StartDate:   domain.YearMonth{Year: 2025, Month: time.July},
			EndDate:     yearMonthPtr(2025, time.September),
		}

		storage := mocks.NewStorage(t)
		storage.On("GetByID", mock.Anything, subID).Return(existingSub, nil).Once()

		svc := NewService(storage)

		err := svc.UpdatePartialByID(context.Background(), subID, update)
		require.ErrorIs(t, err, domain.ErrEndDateBeforeStart)
	})

	t.Run("returns error when start date is after existing end date", func(t *testing.T) {
		t.Parallel()

		startDate := domain.YearMonth{Year: 2025, Month: time.October}
		update := domain.SubscriptionUpdate{
			StartDate: &startDate,
		}
		existingSub := domain.Subscription{
			ID:          subID,
			ServiceName: "Yandex Plus",
			Price:       400,
			UserID:      uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
			StartDate:   domain.YearMonth{Year: 2025, Month: time.July},
			EndDate:     yearMonthPtr(2025, time.September),
		}

		storage := mocks.NewStorage(t)
		storage.On("GetByID", mock.Anything, subID).Return(existingSub, nil).Once()

		svc := NewService(storage)

		err := svc.UpdatePartialByID(context.Background(), subID, update)
		require.ErrorIs(t, err, domain.ErrEndDateBeforeStart)
	})

	t.Run("empty update returns error", func(t *testing.T) {
		t.Parallel()

		storage := mocks.NewStorage(t)
		svc := NewService(storage)

		err := svc.UpdatePartialByID(context.Background(), subID, domain.SubscriptionUpdate{})
		require.ErrorIs(t, err, domain.ErrUpdateEmpty)
	})
}

func TestServiceDeleteByID(t *testing.T) {
	t.Parallel()

	subID := uuid.MustParse("b0cd6d63-75d0-4540-900f-b72c5d7c0153")

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		storage := mocks.NewStorage(t)
		storage.On("DeleteByID", mock.Anything, subID).Return(nil).Once()

		svc := NewService(storage)

		err := svc.DeleteByID(context.Background(), subID)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		storage := mocks.NewStorage(t)
		storage.On("DeleteByID", mock.Anything, subID).Return(domain.ErrSubscriptionNotFound).Once()

		svc := NewService(storage)

		err := svc.DeleteByID(context.Background(), subID)
		require.ErrorIs(t, err, domain.ErrSubscriptionNotFound)
	})

	t.Run("unexpected storage error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("storage failed")
		storage := mocks.NewStorage(t)
		storage.On("DeleteByID", mock.Anything, subID).Return(wantErr).Once()

		svc := NewService(storage)

		err := svc.DeleteByID(context.Background(), subID)
		require.Error(t, err)
		require.ErrorIs(t, err, wantErr)
		require.Contains(t, err.Error(), "delete sub by id in storage")
	})
}

func yearMonthPtr(year int, month time.Month) *domain.YearMonth {
	return &domain.YearMonth{
		Year:  year,
		Month: month,
	}
}

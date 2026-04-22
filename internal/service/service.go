package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gladinov/e"
	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/gladinov/effective_mobile_test_assignment/internal/repository/postgres"
	"github.com/google/uuid"
)

type Service struct {
	logger  *slog.Logger
	storage Storage
	now     func() time.Time
}

func NewService(logger *slog.Logger, storage Storage) *Service {
	return &Service{
		logger:  logger,
		storage: storage,
		now:     time.Now,
	}
}

type Storage interface {
	Create(ctx context.Context, sub domain.Subscription) (uuid.UUID, error)
	GetByID(ctx context.Context, subID uuid.UUID) (domain.Subscription, error)
	UpdateByID(ctx context.Context, subID uuid.UUID, sub domain.Subscription) error
	DeleteByID(ctx context.Context, subID uuid.UUID) error
	List(ctx context.Context) ([]domain.Subscription, error)
	GetFiltredSubs(ctx context.Context, filter domain.FilterTotal) ([]domain.Subscription, error)
}

func (s *Service) Create(ctx context.Context, sub domain.Subscription) (uuid.UUID, error) {
	return s.storage.Create(ctx, sub)
}

func (s *Service) GetByID(ctx context.Context, subID uuid.UUID) (domain.Subscription, error) {
	sub, err := s.storage.GetByID(ctx, subID)
	if err != nil {
		if errors.Is(err, postgres.ErrSubscriptionNotFound) {
			return sub, domain.ErrSubscriptionNotFound
		}
		return sub, e.WrapIfErr("failed to get sub by id from storage", err)
	}
	return sub, nil
}

func (s *Service) List(ctx context.Context) ([]domain.Subscription, error) {
	return s.storage.List(ctx)
}

func (s *Service) UpdateByID(ctx context.Context, subID uuid.UUID, sub domain.Subscription) error {
	err := s.storage.UpdateByID(ctx, subID, sub)
	if err != nil {
		if errors.Is(err, postgres.ErrSubscriptionNotFound) {
			return domain.ErrSubscriptionNotFound
		}
		return e.WrapIfErr("failed to update sub by id in storage", err)
	}
	return nil
}

func (s *Service) DeleteByID(ctx context.Context, subID uuid.UUID) error {
	err := s.storage.DeleteByID(ctx, subID)
	if err != nil {
		if errors.Is(err, postgres.ErrSubscriptionNotFound) {
			return domain.ErrSubscriptionNotFound
		}
		return e.WrapIfErr("failed to update sub by id in storage", err)
	}
	return nil
}

func (s *Service) GetTotal(ctx context.Context, filter domain.FilterTotal) (int, error) {
	subs, err := s.storage.GetFiltredSubs(ctx, filter)
	if err != nil {
		return 0, e.WrapIfErr("failed to get filtered subscribtions from strorage", err)
	}
	var sum int
	for i := range subs {
		sum += s.pricePaidInPeriod(filter, subs[i])
	}
	return sum, nil
}

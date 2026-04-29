package service

import (
	"context"
	"errors"
	"time"

	"github.com/gladinov/e"
	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/google/uuid"
)

type Service struct {
	storage Storage
	now     func() time.Time
}

func NewService(storage Storage) *Service {
	return &Service{
		storage: storage,
		now:     time.Now,
	}
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=Storage
type Storage interface {
	Create(ctx context.Context, sub domain.Subscription) (uuid.UUID, error)
	GetByID(ctx context.Context, subID uuid.UUID) (domain.Subscription, error)
	UpdateByID(ctx context.Context, subID uuid.UUID, sub domain.Subscription) error
	UpdatePartialByID(ctx context.Context, subID uuid.UUID, update domain.SubscriptionUpdate) error
	DeleteByID(ctx context.Context, subID uuid.UUID) error
	List(ctx context.Context, pagination domain.Pagination) ([]domain.Subscription, error)
	GetFilteredSubs(ctx context.Context, filter domain.FilterTotal) ([]domain.Subscription, error)
}

func (s *Service) Create(ctx context.Context, sub domain.Subscription) (uuid.UUID, error) {
	return s.storage.Create(ctx, sub)
}

func (s *Service) GetByID(ctx context.Context, subID uuid.UUID) (domain.Subscription, error) {
	sub, err := s.storage.GetByID(ctx, subID)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return sub, domain.ErrSubscriptionNotFound
		}
		return sub, e.WrapIfErr("get sub by id from storage", err)
	}
	return sub, nil
}

func (s *Service) List(ctx context.Context, pagination domain.Pagination) ([]domain.Subscription, error) {
	return s.storage.List(ctx, pagination)
}

func (s *Service) UpdateByID(ctx context.Context, subID uuid.UUID, sub domain.Subscription) error {
	err := s.storage.UpdateByID(ctx, subID, sub)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return domain.ErrSubscriptionNotFound
		}
		return e.WrapIfErr("update sub by id in storage", err)
	}
	return nil
}

func (s *Service) UpdatePartialByID(ctx context.Context, subID uuid.UUID, update domain.SubscriptionUpdate) error {
	err := s.storage.UpdatePartialByID(ctx, subID, update)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return domain.ErrSubscriptionNotFound
		}
		return e.WrapIfErr("update partial sub by id in storage", err)
	}
	return nil
}

func (s *Service) DeleteByID(ctx context.Context, subID uuid.UUID) error {
	err := s.storage.DeleteByID(ctx, subID)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			return domain.ErrSubscriptionNotFound
		}
		return e.WrapIfErr("delete sub by id in storage", err)
	}
	return nil
}

func (s *Service) GetTotal(ctx context.Context, filter domain.FilterTotal) (int, error) {
	subs, err := s.storage.GetFilteredSubs(ctx, filter)
	if err != nil {
		return 0, e.WrapIfErr("get filtered subscriptions from storage", err)
	}
	var sum int
	for i := range subs {
		sum += pricePaidInPeriod(s.now, filter, subs[i])
	}
	return sum, nil
}

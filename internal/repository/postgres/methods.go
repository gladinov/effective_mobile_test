package postgres

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/gladinov/e"
	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (s *Storage) Create(ctx context.Context, sub domain.Subscription) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, s.dbQueryTimeout)
	defer cancel()
	startDate := mapYearMonthToSql(sub.StartDate)
	endDate := mapPtrYearMonthToSql(sub.EndDate)

	insertSQL, insertArgs, err := psql.
		Insert(subscriptionTable).
		Columns(colServiceName, colPrice, colUserID, colStartDate, colEndDate).
		Values(sub.ServiceName, sub.Price, sub.UserID, startDate, endDate).
		Suffix("RETURNING " + colID).
		ToSql()
	if err != nil {
		return uuid.UUID{}, e.WrapIfErr("failed to create query", err)
	}

	var newID uuid.UUID
	if err := s.db.QueryRow(ctx, insertSQL, insertArgs...).Scan(&newID); err != nil {
		return uuid.UUID{}, e.WrapIfErr("failed to query row", err)
	}

	return newID, nil
}

func (s *Storage) GetByID(ctx context.Context, subID uuid.UUID) (domain.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, s.dbQueryTimeout)
	defer cancel()
	subRow, err := s.getRowByID(ctx, subID)
	if err != nil {
		return domain.Subscription{}, err
	}

	return subRow.ToDomain(), nil
}

func (s *Storage) getRowByID(ctx context.Context, subID uuid.UUID) (subscriptionRow, error) {
	query := psql.
		Select(colID, colServiceName, colPrice, colUserID, colStartDate, colEndDate).
		From(subscriptionTable).
		Where(sq.Eq{colID: subID})

	selectSQL, selectArgs, err := query.ToSql()
	if err != nil {
		return subscriptionRow{}, e.WrapIfErr("failed to build select query", err)
	}

	rows, err := s.db.Query(ctx, selectSQL, selectArgs...)
	if err != nil {
		return subscriptionRow{}, e.WrapIfErr("failed to execute select query", err)
	}
	defer rows.Close()

	sub, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[subscriptionRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscriptionRow{}, ErrSubscriptionNotFound
		}
		return subscriptionRow{}, e.WrapIfErr("failed to collect one row", err)
	}

	return sub, nil
}

func (s *Storage) UpdateByID(ctx context.Context, subID uuid.UUID, sub domain.Subscription) error {
	ctx, cancel := context.WithTimeout(ctx, s.dbQueryTimeout)
	defer cancel()
	updateSQL, updateArgs, err := psql.
		Update(subscriptionTable).
		Set(colServiceName, sub.ServiceName).
		Set(colPrice, sub.Price).
		Set(colUserID, sub.UserID).
		Set(colStartDate, mapYearMonthToSql(sub.StartDate)).
		Set(colEndDate, mapPtrYearMonthToSql(sub.EndDate)).
		Where(sq.Eq{colID: subID}).
		ToSql()
	if err != nil {
		return e.WrapIfErr("failed to build update query", err)
	}
	tag, err := s.db.Exec(ctx, updateSQL, updateArgs...)
	if err != nil {
		return e.WrapIfErr("failed to execute update query", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}

func (s *Storage) DeleteByID(ctx context.Context, subID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, s.dbQueryTimeout)
	defer cancel()
	deleteSQL, deleteArgs, err := psql.
		Delete(subscriptionTable).
		Where(sq.Eq{colID: subID}).
		ToSql()
	if err != nil {
		return e.WrapIfErr("failed to build delete query", err)
	}

	tag, err := s.db.Exec(ctx, deleteSQL, deleteArgs...)
	if err != nil {
		return e.WrapIfErr("failed to execute delete query", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}

func (s *Storage) List(ctx context.Context) ([]domain.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, s.dbQueryTimeout)
	defer cancel()
	subRows, err := s.listRows(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Subscription, 0, len(subRows))
	for i := range subRows {
		res = append(res, subRows[i].ToDomain())
	}
	return res, nil
}

func (s *Storage) listRows(ctx context.Context) ([]subscriptionRow, error) {
	listSQL, listArgs, err := psql.
		Select(colID, colServiceName, colPrice, colUserID, colStartDate, colEndDate).
		From(subscriptionTable).
		ToSql()
	if err != nil {
		return nil, e.WrapIfErr("failed to build select query", err)
	}

	rows, err := s.db.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, e.WrapIfErr("failed to execute select query", err)
	}
	defer rows.Close()

	subs, err := pgx.CollectRows(rows, pgx.RowToStructByName[subscriptionRow])
	if err != nil {
		return nil, e.WrapIfErr("failed to collect rows", err)
	}

	return subs, nil
}

func (s *Storage) GetFiltredSubs(ctx context.Context, filter domain.FilterTotal) ([]domain.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, s.dbQueryTimeout)
	defer cancel()
	// TODO: check from > to
	subRows, err := s.getFiltredSubsRows(ctx, filter)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Subscription, 0, len(subRows))
	for i := range subRows {
		res = append(res, subRows[i].ToDomain())
	}
	return res, nil
}

func (s *Storage) getFiltredSubsRows(ctx context.Context, filter domain.FilterTotal) ([]subscriptionRow, error) {
	query := psql.
		Select(colID, colServiceName, colPrice, colUserID, colStartDate, colEndDate).
		From(subscriptionTable)

	queryWithFilters := applyFilters(filter, query)

	totalSQL, totalArgs, err := queryWithFilters.ToSql()
	if err != nil {
		return nil, e.WrapIfErr("failed to build select query", err)
	}
	rows, err := s.db.Query(ctx, totalSQL, totalArgs...)
	if err != nil {
		return nil, e.WrapIfErr("failed to execute select query", err)
	}
	defer rows.Close()

	subs, err := pgx.CollectRows(rows, pgx.RowToStructByName[subscriptionRow])
	if err != nil {
		return nil, e.WrapIfErr("failed to collect rows", err)
	}

	return subs, nil
}

func applyFilters(filter domain.FilterTotal, query sq.SelectBuilder) sq.SelectBuilder {
	if filter.UserID != nil {
		query = query.
			Where(sq.Eq{colUserID: *filter.UserID})
	}

	if filter.ServiceName != nil {
		query = query.
			Where(sq.Eq{colServiceName: *filter.ServiceName})
	}

	switch {
	case filter.From != nil && filter.To != nil:
		// Фильтруем по полям БД startDate и endDate:
		// Если оба фильтра from и to не равны nil,
		// то нам нужно получить все строки, которые удовлетворяют условию:
		// startDate >= from || endDate <= to
		query = query.Where(
			sq.Or{
				sq.GtOrEq{colStartDate: *filter.From},
				sq.LtOrEq{colEndDate: filter.To},
			})
	case filter.To != nil:
		query = query.Where(
			sq.GtOrEq{colStartDate: *filter.From})
	case filter.From != nil:
		query = query.Where(sq.LtOrEq{colEndDate: *filter.To})
	}

	if filter.From != nil {
		query = query.
			Where(sq.Eq{colStartDate: *filter.From})
	}

	if filter.To != nil {
		query = query.
			Where(sq.Eq{colEndDate: *filter.To})
	}
	return query
}

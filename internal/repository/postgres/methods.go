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
		return uuid.UUID{}, e.WrapIfErr("create query", err)
	}

	var newID uuid.UUID
	if err := s.db.QueryRow(ctx, insertSQL, insertArgs...).Scan(&newID); err != nil {
		return uuid.UUID{}, e.WrapIfErr("query row", err)
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
		return subscriptionRow{}, e.WrapIfErr("build select query", err)
	}

	rows, err := s.db.Query(ctx, selectSQL, selectArgs...)
	if err != nil {
		return subscriptionRow{}, e.WrapIfErr("execute select query", err)
	}
	defer rows.Close()

	sub, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[subscriptionRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return subscriptionRow{}, domain.ErrSubscriptionNotFound
		}
		return subscriptionRow{}, e.WrapIfErr("collect one row", err)
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
		return e.WrapIfErr("build update query", err)
	}
	tag, err := s.db.Exec(ctx, updateSQL, updateArgs...)
	if err != nil {
		return e.WrapIfErr("execute update query", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrSubscriptionNotFound
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
		return e.WrapIfErr("build delete query", err)
	}

	tag, err := s.db.Exec(ctx, deleteSQL, deleteArgs...)
	if err != nil {
		return e.WrapIfErr("execute delete query", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrSubscriptionNotFound
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
		return nil, e.WrapIfErr("build select query", err)
	}

	rows, err := s.db.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, e.WrapIfErr("execute select query", err)
	}
	defer rows.Close()

	subs, err := pgx.CollectRows(rows, pgx.RowToStructByName[subscriptionRow])
	if err != nil {
		return nil, e.WrapIfErr("collect rows", err)
	}

	return subs, nil
}

func (s *Storage) GetFilteredSubs(ctx context.Context, filter domain.FilterTotal) ([]domain.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, s.dbQueryTimeout)
	defer cancel()
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
		return nil, e.WrapIfErr("build select query", err)
	}
	rows, err := s.db.Query(ctx, totalSQL, totalArgs...)
	if err != nil {
		return nil, e.WrapIfErr("execute select query", err)
	}
	defer rows.Close()

	subs, err := pgx.CollectRows(rows, pgx.RowToStructByName[subscriptionRow])
	if err != nil {
		return nil, e.WrapIfErr("collect rows", err)
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
		from := mapYearMonthToSql(*filter.From)
		to := mapYearMonthToSql(*filter.To)
		query = query.Where(
			sq.LtOrEq{colStartDate: to},
		).Where(
			sq.Or{
				sq.Expr(colEndDate + " IS NULL"),
				sq.GtOrEq{colEndDate: from},
			},
		)
	case filter.To != nil:
		to := mapYearMonthToSql(*filter.To)
		query = query.Where(
			sq.LtOrEq{colStartDate: to},
		)
	case filter.From != nil:
		from := mapYearMonthToSql(*filter.From)
		query = query.Where(
			sq.Or{
				sq.Expr(colEndDate + " IS NULL"),
				sq.GtOrEq{colEndDate: from},
			},
		)
	}
	return query
}

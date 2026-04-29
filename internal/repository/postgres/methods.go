package postgres

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/gladinov/e"
	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

const subscriptionDatesCheckConstraint = "subscriptions_end_date_after_start_date_check"

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
		if isConstraintViolation(err, subscriptionDatesCheckConstraint) {
			return uuid.UUID{}, domain.ErrEndDateBeforeStart
		}
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
		if isConstraintViolation(err, subscriptionDatesCheckConstraint) {
			return domain.ErrEndDateBeforeStart
		}
		return e.WrapIfErr("execute update query", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrSubscriptionNotFound
	}

	return nil
}

func (s *Storage) UpdatePartialByID(ctx context.Context, subID uuid.UUID, update domain.SubscriptionUpdate) error {
	if !update.HasChanges() {
		return domain.ErrUpdateEmpty
	}

	ctx, cancel := context.WithTimeout(ctx, s.dbQueryTimeout)
	defer cancel()

	query := psql.Update(subscriptionTable)
	query = applySubscriptionUpdate(update, query)
	query = query.Where(sq.Eq{colID: subID})

	updateSQL, updateArgs, err := query.ToSql()
	if err != nil {
		return e.WrapIfErr("build partial update query", err)
	}
	tag, err := s.db.Exec(ctx, updateSQL, updateArgs...)
	if err != nil {
		if isConstraintViolation(err, subscriptionDatesCheckConstraint) {
			return domain.ErrEndDateBeforeStart
		}
		return e.WrapIfErr("execute partial update query", err)
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

func (s *Storage) List(ctx context.Context, pagination domain.Pagination) ([]domain.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, s.dbQueryTimeout)
	defer cancel()
	subRows, err := s.listRows(ctx, pagination)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Subscription, 0, len(subRows))
	for i := range subRows {
		res = append(res, subRows[i].ToDomain())
	}
	return res, nil
}

func (s *Storage) listRows(ctx context.Context, pagination domain.Pagination) ([]subscriptionRow, error) {
	listSQL, listArgs, err := psql.
		Select(colID, colServiceName, colPrice, colUserID, colStartDate, colEndDate).
		From(subscriptionTable).
		OrderBy(colID).
		Limit(pagination.Limit).
		Offset(pagination.Offset).
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

func isConstraintViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == "23514" &&
		pgErr.ConstraintName == constraintName
}

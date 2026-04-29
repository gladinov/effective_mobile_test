package postgres

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
)

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

func applySubscriptionUpdate(update domain.SubscriptionUpdate, query sq.UpdateBuilder) sq.UpdateBuilder {
	if update.ServiceName != nil {
		query = query.Set(colServiceName, *update.ServiceName)
	}

	if update.Price != nil {
		query = query.Set(colPrice, *update.Price)
	}

	if update.UserID != nil {
		query = query.Set(colUserID, *update.UserID)
	}

	if update.StartDate != nil {
		query = query.Set(colStartDate, mapYearMonthToSql(*update.StartDate))
	}

	if update.EndDate.Value != nil {
		query = query.Set(colEndDate, mapYearMonthToSql(*update.EndDate.Value))
	}

	if update.EndDate.Clear {
		query = query.Set(colEndDate, nil)
	}

	return query
}

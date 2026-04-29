package postgres

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/gladinov/effective_mobile_test_assignment/internal/domain"
)

const (
	effectivePeriodsAlias = "effective_periods"
	colEffectiveStart     = "effective_start"
	colEffectiveEnd       = "effective_end"
)

const totalPaidQueryTemplate = `
WITH effective_periods AS (
    %s
),
paid_months AS (
    SELECT
        price,
        GREATEST(
            0,
            (
                (EXTRACT(YEAR FROM effective_end)::int -
                 EXTRACT(YEAR FROM effective_start)::int) * 12
                +
                (EXTRACT(MONTH FROM effective_end)::int -
                 EXTRACT(MONTH FROM effective_start)::int)
                + 1
            )
        ) AS months
    FROM effective_periods
)
SELECT COALESCE(SUM(price * months), 0)
FROM paid_months`

type sqlFragment struct {
	sql  string
	args []any
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

func effectiveStartExpression(filter domain.FilterTotal) sqlFragment {
	if filter.From == nil {
		return sqlFragment{sql: colStartDate}
	}

	// Начинаем считать с более поздней даты: начало подписки или filter.from.
	return sqlFragment{
		sql:  fmt.Sprintf("GREATEST(%s, ?)", colStartDate),
		args: []any{mapYearMonthToSql(*filter.From)},
	}
}

func effectiveEndExpression(filter domain.FilterTotal, currentMonth domain.YearMonth) sqlFragment {
	currentMonthSQL := mapYearMonthToSql(currentMonth)
	// Для открытой подписки end_date равен текущему месяцу расчета.
	endDateOrCurrentMonth := sqlFragment{
		sql:  fmt.Sprintf("COALESCE(%s, ?)", colEndDate),
		args: []any{currentMonthSQL},
	}

	if filter.To == nil {
		return endDateOrCurrentMonth
	}

	// Если задан filter.to, не считаем подписку дальше этого месяца.
	return sqlFragment{
		sql:  fmt.Sprintf("LEAST(%s, ?)", endDateOrCurrentMonth.sql),
		args: append(endDateOrCurrentMonth.args, mapYearMonthToSql(*filter.To)),
	}
}

func totalPaidQuery(filter domain.FilterTotal, currentMonth domain.YearMonth) (string, []any, error) {
	effectivePeriodsSQL, effectivePeriodsArgs, err := effectivePeriodsQuery(filter, currentMonth).ToSql()
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf(totalPaidQueryTemplate, effectivePeriodsSQL), effectivePeriodsArgs, nil
}

func effectivePeriodsQuery(filter domain.FilterTotal, currentMonth domain.YearMonth) sq.SelectBuilder {
	effectiveStart := effectiveStartExpression(filter)
	effectiveEnd := effectiveEndExpression(filter, currentMonth)

	query := psql.
		Select().
		Column(colPrice).
		// Эти alias используются во внешнем CTE для расчета количества месяцев.
		Column(aliasedExpr(effectiveStart, colEffectiveStart)).
		Column(aliasedExpr(effectiveEnd, colEffectiveEnd)).
		From(subscriptionTable)

	return applyFilters(filter, query)
}

func aliasedExpr(expr sqlFragment, alias string) sq.Sqlizer {
	return sq.Expr(fmt.Sprintf("%s AS %s", expr.sql, alias), expr.args...)
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

package pgrepo

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

func scanOwnerID(row pgx.Row) (uuid.UUID, error) {
	var ownerID uuid.UUID

	if err := row.Scan(&ownerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, domain.ErrNotFound
		}
		return uuid.Nil, err
	}

	return ownerID, nil
}

func buildOrderLimitClause(
	params *domain.ListParams,
	allowedSort map[string]string,
	defaultSort string,
) string {
	sortCol, ok := allowedSort[params.SortBy]
	if !ok {
		sortCol = defaultSort
	}

	order := "ASC"
	if strings.ToUpper(params.Order) == "DESC" {
		order = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s LIMIT $2 OFFSET $3", sortCol, order)
}

func buildListQuery(
	baseQuery string,
	params *domain.ListParams,
	allowedSort map[string]string,
	defaultSort string,
) string {
	return fmt.Sprintf(`
		%s
		%s
	`, baseQuery, buildOrderLimitClause(params, allowedSort, defaultSort))
}

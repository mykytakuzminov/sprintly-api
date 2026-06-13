package pgrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

type ColumnRepo struct {
	db DB
}

func NewColumnRepository(db DB) domain.ColumnRepository {
	return &ColumnRepo{db: db}
}

func (r *ColumnRepo) Create(ctx context.Context, column *domain.Column) error {
	query := `
		INSERT INTO columns (id, board_id, name, position)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query,
		column.ID,
		column.BoardID,
		column.Name,
		column.Position,
	)

	return err
}

func (r *ColumnRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Column, error) {
	query := `
		SELECT id, board_id, name, position
		FROM columns
		WHERE id = $1
	`

	return scanColumn(r.db.QueryRow(ctx, query, id))
}

func (r *ColumnRepo) GetOwnerID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	query := `
		SELECT boards.owner_id
		FROM boards
		INNER JOIN columns ON boards.id = columns.board_id
		WHERE columns.id = $1
	`

	return scanOwnerID(r.db.QueryRow(ctx, query, id))
}

func (r *ColumnRepo) GetAllByBoardID(
	ctx context.Context,
	boardID uuid.UUID,
	params *domain.ListParams,
) ([]*domain.Column, error) {
	query := getColumnListQuery(params)

	var columns []*domain.Column

	rows, err := r.db.Query(ctx, query, boardID, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		column := &domain.Column{}
		if err := rows.Scan(
			&column.ID,
			&column.BoardID,
			&column.Name,
			&column.Position,
		); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

func (r *ColumnRepo) Update(ctx context.Context, column *domain.Column) error {
	query := `
		UPDATE columns
		SET name = $2, position = $3
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, column.ID, column.Name, column.Position)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *ColumnRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM columns WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func scanColumn(row pgx.Row) (*domain.Column, error) {
	column := &domain.Column{}

	if err := row.Scan(
		&column.ID,
		&column.BoardID,
		&column.Name,
		&column.Position,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return column, nil
}

func getColumnListQuery(params *domain.ListParams) string {
	allowedSort := map[string]string{
		"name": "name",
	}

	baseQuery := `
		SELECT id, board_id, name, position
		FROM columns
		WHERE board_id = $1
	`

	return buildListQuery(baseQuery, params, allowedSort, "name")
}

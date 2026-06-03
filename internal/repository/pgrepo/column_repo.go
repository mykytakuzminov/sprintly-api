package pgrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

type ColumnRepo struct {
	pool *pgxpool.Pool
}

func NewColumnRepository(pool *pgxpool.Pool) domain.ColumnRepository {
	return &ColumnRepo{pool: pool}
}

func (r *ColumnRepo) Create(ctx context.Context, column *domain.Column) error {
	query := `
		INSERT INTO columns (id, board_id, name, position)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query,
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

	return scanColumn(r.pool.QueryRow(ctx, query, id))
}

func (r *ColumnRepo) GetOwnerID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	query := `
		SELECT boards.owner_id
		FROM boards
		INNER JOIN columns ON boards.id = columns.board_id
		WHERE columns.id = $1
	`

	return scanOwnerID(r.pool.QueryRow(ctx, query, id))
}

func (r *ColumnRepo) GetAllByBoardID(ctx context.Context, boardID uuid.UUID) ([]*domain.Column, error) {
	query := `
		SELECT id, board_id, name, position
		FROM columns
		WHERE board_id = $1
	`

	var columns []*domain.Column

	rows, err := r.pool.Query(ctx, query, boardID)
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

	tag, err := r.pool.Exec(ctx, query, column.ID, column.Name, column.Position)
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

	tag, err := r.pool.Exec(ctx, query, id)
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

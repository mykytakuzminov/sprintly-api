package pgrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

type BoardRepo struct {
	db DB
}

func NewBoardRepository(db DB) domain.BoardRepository {
	return &BoardRepo{db: db}
}

func (r *BoardRepo) Create(ctx context.Context, board *domain.Board) error {
	query := `
		INSERT INTO boards (id, owner_id, name, description)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`

	return r.db.QueryRow(ctx, query,
		board.ID,
		board.OwnerID,
		board.Name,
		board.Description,
	).Scan(&board.CreatedAt, &board.UpdatedAt)
}

func (r *BoardRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
	query := `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM boards
		WHERE id = $1
	`

	return scanBoard(r.db.QueryRow(ctx, query, id))
}

func (r *BoardRepo) GetAllByUserID(
	ctx context.Context,
	userID uuid.UUID,
	params *domain.ListParams,
) ([]*domain.Board, error) {
	query := getBoardListQuery(params)

	var boards []*domain.Board

	rows, err := r.db.Query(ctx, query, userID, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		board := &domain.Board{}
		if err := rows.Scan(
			&board.ID,
			&board.OwnerID,
			&board.Name,
			&board.Description,
			&board.CreatedAt,
			&board.UpdatedAt,
		); err != nil {
			return nil, err
		}
		boards = append(boards, board)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return boards, nil
}

func (r *BoardRepo) Update(ctx context.Context, board *domain.Board) error {
	query := `
		UPDATE boards
		SET name = $2, description = $3, updated_at = NOW()
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, board.ID, board.Name, board.Description)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *BoardRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM boards WHERE id = $1
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

func scanBoard(row pgx.Row) (*domain.Board, error) {
	board := &domain.Board{}

	if err := row.Scan(
		&board.ID,
		&board.OwnerID,
		&board.Name,
		&board.Description,
		&board.CreatedAt,
		&board.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return board, nil
}

func buildOrderLimitClause(params *domain.ListParams, allowedSort map[string]string) string {
	sortCol, ok := allowedSort[params.SortBy]
	if !ok {
		sortCol = "created_at"
	}

	order := "ASC"
	if strings.ToUpper(params.Order) == "DESC" {
		order = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s LIMIT $2 OFFSET $3", sortCol, order)
}

func getBoardListQuery(params *domain.ListParams) string {
	allowedSort := map[string]string{
		"created_at": "created_at",
		"name":       "name",
		"updated_at": "updated_at",
	}

	return fmt.Sprintf(`
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM boards
		WHERE owner_id = $1
		%s
	`, buildOrderLimitClause(params, allowedSort))
}

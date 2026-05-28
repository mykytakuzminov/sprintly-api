package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

type TaskRepo struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) domain.TaskRepository {
	return &TaskRepo{pool: pool}
}

func (r *TaskRepo) Create(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (id, owner_id, column_id, assignee_id, name, description, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	return r.pool.QueryRow(ctx, query,
		task.ID,
		task.OwnerID,
		task.ColumnID,
		task.AssigneeID,
		task.Name,
		task.Description,
		task.DueDate,
	).Scan(&task.CreatedAt, &task.UpdatedAt)
}

func (r *TaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	query := `
		SELECT id, owner_id, column_id, assignee_id, name, description, due_date, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	task := &domain.Task{}

	if err := r.pool.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.OwnerID,
		&task.ColumnID,
		&task.AssigneeID,
		&task.Name,
		&task.Description,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return task, nil
}

func (r *TaskRepo) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Task, error) {
	query := `
		SELECT id, owner_id, column_id, assignee_id, name, description, due_date, created_at, updated_at
		FROM tasks
		WHERE owner_id = $1
	`

	var tasks []*domain.Task

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		task := &domain.Task{}
		if err := rows.Scan(
			&task.ID,
			&task.OwnerID,
			&task.ColumnID,
			&task.AssigneeID,
			&task.Name,
			&task.Description,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskRepo) GetAllByColumnID(ctx context.Context, columnID uuid.UUID) ([]*domain.Task, error) {
	query := `
		SELECT id, owner_id, column_id, assignee_id, name, description, due_date, created_at, updated_at
		FROM tasks
		WHERE column_id = $1
	`

	var tasks []*domain.Task

	rows, err := r.pool.Query(ctx, query, columnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		task := &domain.Task{}
		if err := rows.Scan(
			&task.ID,
			&task.OwnerID,
			&task.ColumnID,
			&task.AssigneeID,
			&task.Name,
			&task.Description,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskRepo) GetAllByAssigneeID(ctx context.Context, assigneeID uuid.UUID) ([]*domain.Task, error) {
	query := `
		SELECT id, owner_id, column_id, assignee_id, name, description, due_date, created_at, updated_at
		FROM tasks
		WHERE assignee_id = $1
	`

	var tasks []*domain.Task

	rows, err := r.pool.Query(ctx, query, assigneeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		task := &domain.Task{}
		if err := rows.Scan(
			&task.ID,
			&task.OwnerID,
			&task.ColumnID,
			&task.AssigneeID,
			&task.Name,
			&task.Description,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskRepo) Update(ctx context.Context, task *domain.Task) error {
	query := `
		UPDATE tasks
		SET column_id = $2, assignee_id = $3, name = $4, description = $5, due_date = $6, updated_at = NOW()
		WHERE id = $1
	`

	tag, err := r.pool.Exec(ctx, query,
		task.ID,
		task.ColumnID,
		task.AssigneeID,
		task.Name,
		task.Description,
		task.DueDate,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *TaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM tasks WHERE id = $1
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

package pgrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

type TaskRepo struct {
	db DB
}

func NewTaskRepository(db DB) domain.TaskRepository {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (id, owner_id, column_id, assignee_id, name, description, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	return r.db.QueryRow(ctx, query,
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

	return scanTask(r.db.QueryRow(ctx, query, id))
}

func (r *TaskRepo) GetOwnerID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	query := `
		SELECT boards.owner_id
		FROM boards
		INNER JOIN columns ON boards.id = columns.board_id
		INNER JOIN tasks ON columns.id = tasks.column_id
		WHERE tasks.id = $1
	`

	return scanOwnerID(r.db.QueryRow(ctx, query, id))
}

func (r *TaskRepo) GetAllByUserID(
	ctx context.Context,
	userID uuid.UUID,
	params *domain.ListParams,
) ([]*domain.Task, error) {
	baseQuery := `
		SELECT id, owner_id, column_id, assignee_id, name, description, due_date, created_at, updated_at
		FROM tasks
		WHERE owner_id = $1
	`
	query := getTaskListQuery(baseQuery, params)

	rows, err := r.db.Query(ctx, query, userID, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

func (r *TaskRepo) GetAllByColumnID(
	ctx context.Context,
	columnID uuid.UUID,
	params *domain.ListParams,
) ([]*domain.Task, error) {
	baseQuery := `
		SELECT id, owner_id, column_id, assignee_id, name, description, due_date, created_at, updated_at
		FROM tasks
		WHERE column_id = $1
	`
	query := getTaskListQuery(baseQuery, params)

	rows, err := r.db.Query(ctx, query, columnID, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

func (r *TaskRepo) GetAllByAssigneeID(
	ctx context.Context,
	assigneeID uuid.UUID,
	params *domain.ListParams,
) ([]*domain.Task, error) {
	baseQuery := `
		SELECT id, owner_id, column_id, assignee_id, name, description, due_date, created_at, updated_at
		FROM tasks
		WHERE assignee_id = $1
	`
	query := getTaskListQuery(baseQuery, params)

	rows, err := r.db.Query(ctx, query, assigneeID, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTasks(rows)
}

func (r *TaskRepo) Update(ctx context.Context, task *domain.Task) error {
	query := `
		UPDATE tasks
		SET column_id = $2, assignee_id = $3, name = $4, description = $5, due_date = $6, updated_at = NOW()
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query,
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

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func scanTask(row pgx.Row) (*domain.Task, error) {
	task := &domain.Task{}

	if err := row.Scan(
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

func scanTasks(rows pgx.Rows) ([]*domain.Task, error) {
	var tasks []*domain.Task

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

func getTaskListQuery(baseQuery string, params *domain.ListParams) string {
	allowedSort := map[string]string{
		"name":       "name",
		"due_date":   "due_date",
		"created_at": "created_at",
		"updated_at": "updated_at",
	}

	return buildListQuery(baseQuery, params, allowedSort, "created_at")
}

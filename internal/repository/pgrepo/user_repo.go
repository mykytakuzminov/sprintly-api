package pgrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
)

type UserRepo struct {
	db DB
}

func NewUserRepository(db DB) domain.UserRepository {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, hash_password)
		VALUES ($1, $2, $3)
		RETURNING role, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		user.ID,
		user.Email,
		user.HashPassword,
	).Scan(&user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return err
	}

	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, hash_password, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	return scanUser(r.db.QueryRow(ctx, query, id))
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, hash_password, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	return scanUser(r.db.QueryRow(ctx, query, email))
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET hash_password = $2, updated_at = NOW()
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, user.ID, user.HashPassword)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM users WHERE id = $1
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

func scanUser(row pgx.Row) (*domain.User, error) {
	user := &domain.User{}

	if err := row.Scan(
		&user.ID,
		&user.Email,
		&user.HashPassword,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

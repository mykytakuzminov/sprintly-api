package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

type MockColumnRepository struct {
	createFn        func(ctx context.Context, column *domain.Column) error
	getByIDFn       func(ctx context.Context, id uuid.UUID) (*domain.Column, error)
	getOwnerIDFn    func(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	getAllByBoardFn func(ctx context.Context, boardID uuid.UUID, params *domain.ListParams) ([]*domain.Column, error)
	updateFn        func(ctx context.Context, column *domain.Column) error
	deleteFn        func(ctx context.Context, id uuid.UUID) error
}

func (m *MockColumnRepository) Create(ctx context.Context, column *domain.Column) error {
	if m.createFn != nil {
		return m.createFn(ctx, column)
	}
	return nil
}

func (m *MockColumnRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Column, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockColumnRepository) GetOwnerID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	if m.getOwnerIDFn != nil {
		return m.getOwnerIDFn(ctx, id)
	}
	return uuid.Nil, nil
}

func (m *MockColumnRepository) GetAllByBoardID(ctx context.Context, boardID uuid.UUID, params *domain.ListParams) ([]*domain.Column, error) {
	if m.getAllByBoardFn != nil {
		return m.getAllByBoardFn(ctx, boardID, params)
	}
	return []*domain.Column{}, nil
}

func (m *MockColumnRepository) Update(ctx context.Context, column *domain.Column) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, column)
	}
	return nil
}

func (m *MockColumnRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestColumnService_Create(t *testing.T) {
	userID := uuid.New()
	boardID := uuid.New()

	board := &domain.Board{
		ID:      boardID,
		OwnerID: userID,
	}

	boardRepo := &MockBoardRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Board, error) {
			return board, nil
		},
	}
	columnRepo := &MockColumnRepository{
		createFn: func(_ context.Context, _ *domain.Column) error {
			return nil
		},
	}

	svc := NewColumnService(columnRepo, boardRepo)

	column, err := svc.Create(context.Background(), userID, boardID, &domain.CreateColumnInput{
		Name:     "In Progress",
		Position: 1,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if column.Name != "In Progress" {
		t.Errorf("expected column name %v, got %v", "In Progress", column.Name)
	}
	if column.BoardID != boardID {
		t.Errorf("expected BoardID %v, got %v", boardID, column.BoardID)
	}
}

func TestColumnService_Create_InvalidBodyRequest(t *testing.T) {
	svc := NewColumnService(&MockColumnRepository{}, &MockBoardRepository{})

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), &domain.CreateColumnInput{
		Name:     strings.Repeat("a", 101),
		Position: 1,
	})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestColumnService_Create_NotFound(t *testing.T) {
	boardRepo := &MockBoardRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Board, error) {
			return nil, domain.ErrNotFound
		},
	}

	svc := NewColumnService(&MockColumnRepository{}, boardRepo)

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), &domain.CreateColumnInput{
		Name:     "In Progress",
		Position: 1,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestColumnService_Create_AccessDenied(t *testing.T) {
	boardRepo := &MockBoardRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Board, error) {
			return &domain.Board{
				ID:      uuid.New(),
				OwnerID: uuid.New(),
			}, nil
		},
	}

	svc := NewColumnService(&MockColumnRepository{}, boardRepo)

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New(), &domain.CreateColumnInput{
		Name:     "In Progress",
		Position: 1,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestColumnService_Update(t *testing.T) {
	userID := uuid.New()
	columnID := uuid.New()

	columnRepo := &MockColumnRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return userID, nil
		},
		updateFn: func(_ context.Context, _ *domain.Column) error {
			return nil
		},
	}

	svc := NewColumnService(columnRepo, &MockBoardRepository{})

	err := svc.Update(context.Background(), columnID, userID, &domain.UpdateColumnInput{
		Name:     "Done",
		Position: 2,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestColumnService_Update_InvalidBodyRequest(t *testing.T) {
	svc := NewColumnService(&MockColumnRepository{}, &MockBoardRepository{})

	err := svc.Update(context.Background(), uuid.New(), uuid.New(), &domain.UpdateColumnInput{
		Name:     strings.Repeat("a", 101),
		Position: 1,
	})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestColumnService_Update_NotFound(t *testing.T) {
	columnRepo := &MockColumnRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return uuid.Nil, domain.ErrNotFound
		},
	}

	svc := NewColumnService(columnRepo, &MockBoardRepository{})

	err := svc.Update(context.Background(), uuid.New(), uuid.New(), &domain.UpdateColumnInput{
		Name:     "Done",
		Position: 1,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestColumnService_Update_AccessDenied(t *testing.T) {
	columnRepo := &MockColumnRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return uuid.New(), nil
		},
	}

	svc := NewColumnService(columnRepo, &MockBoardRepository{})

	err := svc.Update(context.Background(), uuid.New(), uuid.New(), &domain.UpdateColumnInput{
		Name:     "Done",
		Position: 1,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestColumnService_Delete(t *testing.T) {
	userID := uuid.New()

	columnRepo := &MockColumnRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return userID, nil
		},
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return nil
		},
	}

	svc := NewColumnService(columnRepo, &MockBoardRepository{})

	err := svc.Delete(context.Background(), uuid.New(), userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestColumnService_Delete_NotFound(t *testing.T) {
	columnRepo := &MockColumnRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return uuid.Nil, domain.ErrNotFound
		},
	}

	svc := NewColumnService(columnRepo, &MockBoardRepository{})

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestColumnService_Delete_AccessDenied(t *testing.T) {
	columnRepo := &MockColumnRepository{
		getOwnerIDFn: func(_ context.Context, _ uuid.UUID) (uuid.UUID, error) {
			return uuid.New(), nil
		},
	}

	svc := NewColumnService(columnRepo, &MockBoardRepository{})

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

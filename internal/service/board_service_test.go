package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

type MockBoardRepository struct {
	createFn  func(ctx context.Context, board *domain.Board) error
	getByIDFn func(ctx context.Context, id uuid.UUID) (*domain.Board, error)
	updateFn  func(ctx context.Context, board *domain.Board) error
	deleteFn  func(ctx context.Context, id uuid.UUID) error
}

func (m *MockBoardRepository) Create(ctx context.Context, board *domain.Board) error {
	return m.createFn(ctx, board)
}

func (m *MockBoardRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Board, error) {
	return m.getByIDFn(ctx, id)
}

func (m *MockBoardRepository) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Board, error) {
	return nil, nil
}

func (m *MockBoardRepository) Update(ctx context.Context, board *domain.Board) error {
	return m.updateFn(ctx, board)
}

func (m *MockBoardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestBoardService_Create(t *testing.T) {
	repo := &MockBoardRepository{
		createFn: func(_ context.Context, _ *domain.Board) error {
			return nil
		},
	}

	svc := NewBoardService(repo)

	board, err := svc.Create(context.Background(), uuid.New(), &domain.CreateBoardInput{
		Name:        "name",
		Description: nil,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if board.Name != "name" {
		t.Errorf("expected board name %v, got %v", "name", board.Name)
	}
}

func TestBoardService_Create_InvalidBodyRequest(t *testing.T) {
	svc := NewBoardService(&MockBoardRepository{})

	_, err := svc.Create(context.Background(), uuid.New(), &domain.CreateBoardInput{
		Name:        strings.Repeat("a", 101),
		Description: nil,
	})
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestBoardService_Update(t *testing.T) {
	boardID := uuid.New()
	userID := uuid.New()

	board := &domain.Board{
		ID:      boardID,
		OwnerID: userID,
		Name:    "name",
	}

	repo := &MockBoardRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Board, error) {
			return board, nil
		},
		updateFn: func(_ context.Context, _ *domain.Board) error {
			return nil
		},
	}

	svc := NewBoardService(repo)

	err := svc.Update(context.Background(), boardID, userID, &domain.UpdateBoardInput{
		Name:        "newname",
		Description: nil,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if board.Name != "newname" {
		t.Errorf("expected board name %v, got %v", "newname", board.Name)
	}
}

func TestBoardService_Update_NotFound(t *testing.T) {
	repo := &MockBoardRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Board, error) {
			return nil, domain.ErrNotFound
		},
	}

	svc := NewBoardService(repo)

	err := svc.Update(context.Background(), uuid.New(), uuid.New(), &domain.UpdateBoardInput{
		Name:        "newname",
		Description: nil,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestBoardService_Update_AccessDenied(t *testing.T) {
	boardID := uuid.New()

	board := &domain.Board{
		ID:      boardID,
		OwnerID: uuid.New(),
		Name:    "name",
	}

	repo := &MockBoardRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Board, error) {
			return board, nil
		},
	}

	svc := NewBoardService(repo)

	err := svc.Update(context.Background(), boardID, uuid.New(), &domain.UpdateBoardInput{
		Name:        "newname",
		Description: nil,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestBoardService_Delete(t *testing.T) {
	boardID := uuid.New()
	userID := uuid.New()

	board := &domain.Board{
		ID:      boardID,
		OwnerID: userID,
		Name:    "name",
	}

	repo := &MockBoardRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Board, error) {
			return board, nil
		},
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return nil
		},
	}

	svc := NewBoardService(repo)

	err := svc.Delete(context.Background(), boardID, userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBoardService_Delete_NotFound(t *testing.T) {
	repo := &MockBoardRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Board, error) {
			return nil, domain.ErrNotFound
		},
	}

	svc := NewBoardService(repo)

	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestBoardService_Delete_AccessDenied(t *testing.T) {
	boardID := uuid.New()

	board := &domain.Board{
		ID:      boardID,
		OwnerID: uuid.New(),
		Name:    "name",
	}

	repo := &MockBoardRepository{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Board, error) {
			return board, nil
		},
	}

	svc := NewBoardService(repo)

	err := svc.Delete(context.Background(), boardID, uuid.New())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

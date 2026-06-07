package pgrepo

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

func createTestBoard(ctx context.Context, db DB, ownerID uuid.UUID) *domain.Board {
	board := &domain.Board{
		ID:      uuid.New(),
		OwnerID: ownerID,
		Name:    "board",
	}
	_ = NewBoardRepository(db).Create(ctx, board)
	return board
}

func TestBoardRepo_Create(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		boardRepo := NewBoardRepository(db)

		board := &domain.Board{
			ID:      uuid.New(),
			OwnerID: user.ID,
			Name:    "board",
		}

		err := boardRepo.Create(ctx, board)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if board.CreatedAt.IsZero() {
			t.Errorf("expected CreatedAt to be set")
		}
		if board.UpdatedAt.IsZero() {
			t.Errorf("expected UpdatedAt to be set")
		}
	})
}

func TestBoardRepo_GetByID(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		boardRepo := NewBoardRepository(db)

		board := &domain.Board{
			ID:      uuid.New(),
			OwnerID: user.ID,
			Name:    "board",
		}
		_ = boardRepo.Create(ctx, board)

		found, err := boardRepo.GetByID(ctx, board.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if found.ID != board.ID {
			t.Errorf("expected ID %v, got %v", board.ID, found.ID)
		}
		if found.OwnerID != board.OwnerID {
			t.Errorf("expected OwnerID %v, got %v", board.OwnerID, found.OwnerID)
		}
		if found.Name != board.Name {
			t.Errorf("expected Name %v, got %v", board.Name, found.Name)
		}
	})
}

func TestBoardRepo_GetByID_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewBoardRepository(db)

		_, err := repo.GetByID(ctx, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestBoardRepo_GetAllByUserID(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		boardRepo := NewBoardRepository(db)

		_ = boardRepo.Create(ctx, &domain.Board{
			ID:      uuid.New(),
			OwnerID: user.ID,
			Name:    "board 1",
		})
		_ = boardRepo.Create(ctx, &domain.Board{
			ID:      uuid.New(),
			OwnerID: user.ID,
			Name:    "board 2",
		})

		boards, err := boardRepo.GetAllByUserID(ctx, user.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(boards) != 2 {
			t.Fatalf("expected 2 boards, got %v", len(boards))
		}
	})
}

func TestBoardRepo_GetAllByUserID_Empty(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewBoardRepository(db)

		boards, err := repo.GetAllByUserID(ctx, uuid.New())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(boards) != 0 {
			t.Fatalf("expected 0 boards, got %v", len(boards))
		}
	})
}

func TestBoardRepo_GetAllByUserID_OnlyOwnBoards(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user1 := createTestUser(ctx, db)
		user2 := createTestUser(ctx, db)
		boardRepo := NewBoardRepository(db)

		_ = boardRepo.Create(ctx, &domain.Board{
			ID:      uuid.New(),
			OwnerID: user1.ID,
			Name:    "user1 board",
		})
		_ = boardRepo.Create(ctx, &domain.Board{
			ID:      uuid.New(),
			OwnerID: user2.ID,
			Name:    "user2 board",
		})

		boards, err := boardRepo.GetAllByUserID(ctx, user1.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(boards) != 1 {
			t.Fatalf("expected 1 board, got %v", len(boards))
		}
		if boards[0].OwnerID != user1.ID {
			t.Errorf("expected OwnerID %v, got %v", user1.ID, boards[0].OwnerID)
		}
	})
}

func TestBoardRepo_Update(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		boardRepo := NewBoardRepository(db)

		desc := "new description"
		board.Name = "newboard"
		board.Description = &desc

		err := boardRepo.Update(ctx, board)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		found, _ := boardRepo.GetByID(ctx, board.ID)
		if found.Name != "newboard" {
			t.Errorf("expected Name %v, got %v", "newboard", found.Name)
		}
		if found.Description == nil || *found.Description != desc {
			t.Errorf("expected Description %v, got %v", desc, found.Description)
		}
	})
}

func TestBoardRepo_Update_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewBoardRepository(db)

		err := repo.Update(ctx, &domain.Board{
			ID:   uuid.New(),
			Name: "board",
		})
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestBoardRepo_Delete(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		boardRepo := NewBoardRepository(db)

		err := boardRepo.Delete(ctx, board.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = boardRepo.GetByID(ctx, board.ID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected board to be deleted, got %v", err)
		}
	})
}

func TestBoardRepo_Delete_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewBoardRepository(db)

		err := repo.Delete(ctx, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

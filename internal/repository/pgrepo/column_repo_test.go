package pgrepo

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
)

func TestColumnRepo_Create(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		repo := NewColumnRepository(db)

		column := &domain.Column{
			ID:       uuid.New(),
			BoardID:  board.ID,
			Name:     "column",
			Position: 1,
		}

		err := repo.Create(ctx, column)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestColumnRepo_GetByID(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column := createTestColumn(ctx, db, board.ID)
		repo := NewColumnRepository(db)

		found, err := repo.GetByID(ctx, column.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if found.ID != column.ID {
			t.Errorf("expected ID %v, got %v", column.ID, found.ID)
		}
		if found.BoardID != column.BoardID {
			t.Errorf("expected BoardID %v, got %v", column.BoardID, found.BoardID)
		}
		if found.Name != column.Name {
			t.Errorf("expected Name %v, got %v", column.Name, found.Name)
		}
		if found.Position != column.Position {
			t.Errorf("expected Position %v, got %v", column.Position, found.Position)
		}
	})
}

func TestColumnRepo_GetByID_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewColumnRepository(db)

		_, err := repo.GetByID(ctx, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestColumnRepo_GetOwnerID(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column := createTestColumn(ctx, db, board.ID)
		repo := NewColumnRepository(db)

		ownerID, err := repo.GetOwnerID(ctx, column.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if ownerID != user.ID {
			t.Errorf("expected OwnerID %v, got %v", user.ID, ownerID)
		}
	})
}

func TestColumnRepo_GetOwnerID_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewColumnRepository(db)

		_, err := repo.GetOwnerID(ctx, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestColumnRepo_GetAllByBoardID(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		params := createTestParams(5, 0, "created_at", "ASC")
		repo := NewColumnRepository(db)

		_ = repo.Create(ctx, &domain.Column{
			ID:       uuid.New(),
			BoardID:  board.ID,
			Name:     "column 1",
			Position: 1,
		})
		_ = repo.Create(ctx, &domain.Column{
			ID:       uuid.New(),
			BoardID:  board.ID,
			Name:     "column 2",
			Position: 2,
		})

		columns, err := repo.GetAllByBoardID(ctx, board.ID, params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(columns) != 2 {
			t.Fatalf("expected 2 columns, got %v", len(columns))
		}
	})
}

func TestColumnRepo_GetAllByBoardID_Empty(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		params := createTestParams(5, 0, "created_at", "ASC")
		repo := NewColumnRepository(db)

		columns, err := repo.GetAllByBoardID(ctx, uuid.New(), params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(columns) != 0 {
			t.Fatalf("expected 0 columns, got %v", len(columns))
		}
	})
}

func TestColumnRepo_GetAllByBoardID_OnlyOwnColumns(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board1 := createTestBoard(ctx, db, user.ID)
		board2 := createTestBoard(ctx, db, user.ID)
		params := createTestParams(5, 0, "created_at", "ASC")
		repo := NewColumnRepository(db)

		_ = repo.Create(ctx, &domain.Column{
			ID:       uuid.New(),
			BoardID:  board1.ID,
			Name:     "board1 column",
			Position: 1,
		})
		_ = repo.Create(ctx, &domain.Column{
			ID:       uuid.New(),
			BoardID:  board2.ID,
			Name:     "board2 column",
			Position: 1,
		})

		columns, err := repo.GetAllByBoardID(ctx, board1.ID, params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(columns) != 1 {
			t.Fatalf("expected 1 column, got %v", len(columns))
		}
		if columns[0].BoardID != board1.ID {
			t.Errorf("expected BoardID %v, got %v", board1.ID, columns[0].BoardID)
		}
	})
}

func TestColumnRepo_GetAllByBoardID_WithLimit(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		params := createTestParams(3, 0, "created_at", "ASC")
		repo := NewColumnRepository(db)

		for i := 0; i < 5; i++ {
			_ = repo.Create(ctx, &domain.Column{
				ID:       uuid.New(),
				BoardID:  board.ID,
				Name:     "column",
				Position: 1,
			})
		}

		columns, err := repo.GetAllByBoardID(ctx, board.ID, params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(columns) != 3 {
			t.Fatalf("expected 3 columns, got %v", len(columns))
		}
	})
}

func TestColumnRepo_GetAllByBoardID_WithSortByAndOrder(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		params := createTestParams(3, 0, "name", "DESC")
		repo := NewColumnRepository(db)

		for i := 0; i < 5; i++ {
			_ = repo.Create(ctx, &domain.Column{
				ID:       uuid.New(),
				BoardID:  board.ID,
				Name:     fmt.Sprintf("column%d", i),
				Position: 1,
			})
		}

		columns, err := repo.GetAllByBoardID(ctx, board.ID, params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if columns[0].Name != "column4" {
			t.Fatalf("expected column4, got %v", columns[0].Name)
		}
	})
}

func TestColumnRepo_Update(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column := createTestColumn(ctx, db, board.ID)
		repo := NewColumnRepository(db)

		column.Name = "newcolumn"
		column.Position = 2

		err := repo.Update(ctx, column)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		found, _ := repo.GetByID(ctx, column.ID)
		if found.Name != "newcolumn" {
			t.Errorf("expected Name %v, got %v", "newcolumn", found.Name)
		}
		if found.Position != 2 {
			t.Errorf("expected Position %v, got %v", 2, found.Position)
		}
	})
}

func TestColumnRepo_Update_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewColumnRepository(db)

		err := repo.Update(ctx, &domain.Column{
			ID:       uuid.New(),
			Name:     "column",
			Position: 1,
		})
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestColumnRepo_Delete(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column := createTestColumn(ctx, db, board.ID)
		repo := NewColumnRepository(db)

		err := repo.Delete(ctx, column.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = repo.GetByID(ctx, column.ID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected column to be deleted, got %v", err)
		}
	})
}

func TestColumnRepo_Delete_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewColumnRepository(db)

		err := repo.Delete(ctx, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

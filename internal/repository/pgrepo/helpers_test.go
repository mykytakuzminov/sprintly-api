package pgrepo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/domain"
)

func withTx(t *testing.T, fn func(ctx context.Context, db DB)) {
	t.Helper()

	ctx := context.Background()

	tx, err := testPool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback(ctx)

	fn(ctx, tx)
}

func createTestBoard(ctx context.Context, db DB, ownerID uuid.UUID) *domain.Board {
	board := &domain.Board{
		ID:      uuid.New(),
		OwnerID: ownerID,
		Name:    "board",
	}
	_ = NewBoardRepository(db).Create(ctx, board)
	return board
}

func createTestColumn(ctx context.Context, db DB, boardID uuid.UUID) *domain.Column {
	column := &domain.Column{
		ID:       uuid.New(),
		BoardID:  boardID,
		Name:     "column",
		Position: 1,
	}

	_ = NewColumnRepository(db).Create(ctx, column)
	return column
}

func createTestTask(
	ctx context.Context,
	db DB,
	ownerID, columnID uuid.UUID,
) *domain.Task {
	task := &domain.Task{
		ID:       uuid.New(),
		OwnerID:  ownerID,
		ColumnID: columnID,
		Name:     "task",
	}
	_ = NewTaskRepository(db).Create(ctx, task)
	return task
}

func createTestParams(
	limit int,
	offset int,
	sortBy string,
	order string,
) *domain.ListParams {
	return &domain.ListParams{
		Limit:  limit,
		Offset: offset,
		SortBy: sortBy,
		Order:  order,
	}
}

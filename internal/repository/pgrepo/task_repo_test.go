package pgrepo

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mykytakuzminov/sprintly-api/internal/domain"
)

func TestTaskRepo_Create(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column := createTestColumn(ctx, db, board.ID)
		repo := NewTaskRepository(db)

		task := &domain.Task{
			ID:       uuid.New(),
			OwnerID:  user.ID,
			ColumnID: column.ID,
			Name:     "task",
		}

		err := repo.Create(ctx, task)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if task.CreatedAt.IsZero() {
			t.Errorf("expected CreatedAt to be set")
		}
		if task.UpdatedAt.IsZero() {
			t.Errorf("expected UpdatedAt to be set")
		}
	})
}

func TestTaskRepo_Create_WithOptionalFields(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		assignee := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column := createTestColumn(ctx, db, board.ID)
		repo := NewTaskRepository(db)

		desc := "description"
		dueDate := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Microsecond)

		task := &domain.Task{
			ID:          uuid.New(),
			OwnerID:     user.ID,
			ColumnID:    column.ID,
			AssigneeID:  &assignee.ID,
			Name:        "task",
			Description: &desc,
			DueDate:     &dueDate,
		}

		err := repo.Create(ctx, task)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		found, _ := repo.GetByID(ctx, task.ID)
		if found.AssigneeID == nil || *found.AssigneeID != assignee.ID {
			t.Errorf("expected AssigneeID %v, got %v", assignee.ID, found.AssigneeID)
		}
		if found.Description == nil || *found.Description != desc {
			t.Errorf("expected Description %v, got %v", desc, found.Description)
		}
		if found.DueDate == nil || !found.DueDate.Equal(dueDate) {
			t.Errorf("expected DueDate %v, got %v", dueDate, found.DueDate)
		}
	})
}

func TestTaskRepo_GetByID(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column := createTestColumn(ctx, db, board.ID)
		task := createTestTask(ctx, db, user.ID, column.ID)
		repo := NewTaskRepository(db)

		found, err := repo.GetByID(ctx, task.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if found.ID != task.ID {
			t.Errorf("expected ID %v, got %v", task.ID, found.ID)
		}
		if found.OwnerID != task.OwnerID {
			t.Errorf("expected OwnerID %v, got %v", task.OwnerID, found.OwnerID)
		}
		if found.ColumnID != task.ColumnID {
			t.Errorf("expected ColumnID %v, got %v", task.ColumnID, found.ColumnID)
		}
		if found.Name != task.Name {
			t.Errorf("expected Name %v, got %v", task.Name, found.Name)
		}
	})
}

func TestTaskRepo_GetByID_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewTaskRepository(db)

		_, err := repo.GetByID(ctx, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestTaskRepo_GetOwnerID(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column := createTestColumn(ctx, db, board.ID)
		task := createTestTask(ctx, db, user.ID, column.ID)
		repo := NewTaskRepository(db)

		ownerID, err := repo.GetOwnerID(ctx, task.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if ownerID != user.ID {
			t.Errorf("expected OwnerID %v, got %v", user.ID, ownerID)
		}
	})
}

func TestTaskRepo_GetOwnerID_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewTaskRepository(db)

		_, err := repo.GetOwnerID(ctx, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestTaskRepo_GetAllByUserID(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column := createTestColumn(ctx, db, board.ID)
		params := createTestParams(5, 0, "created_at", "ASC")
		repo := NewTaskRepository(db)

		_ = repo.Create(ctx, &domain.Task{
			ID:       uuid.New(),
			OwnerID:  user.ID,
			ColumnID: column.ID,
			Name:     "task 1",
		})
		_ = repo.Create(ctx, &domain.Task{
			ID:       uuid.New(),
			OwnerID:  user.ID,
			ColumnID: column.ID,
			Name:     "task 2",
		})

		tasks, err := repo.GetAllByUserID(ctx, user.ID, params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(tasks) != 2 {
			t.Fatalf("expected 2 tasks, got %v", len(tasks))
		}
	})
}

func TestTaskRepo_GetAllByUserID_Empty(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		params := createTestParams(5, 0, "created_at", "ASC")
		repo := NewTaskRepository(db)

		tasks, err := repo.GetAllByUserID(ctx, uuid.New(), params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(tasks) != 0 {
			t.Fatalf("expected 0 tasks, got %v", len(tasks))
		}
	})
}

func TestTaskRepo_GetAllByColumnID(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column1 := createTestColumn(ctx, db, board.ID)
		column2 := createTestColumn(ctx, db, board.ID)
		params := createTestParams(5, 0, "created_at", "ASC")
		repo := NewTaskRepository(db)

		_ = repo.Create(ctx, &domain.Task{
			ID:       uuid.New(),
			OwnerID:  user.ID,
			ColumnID: column1.ID,
			Name:     "task in column1",
		})
		_ = repo.Create(ctx, &domain.Task{
			ID:       uuid.New(),
			OwnerID:  user.ID,
			ColumnID: column2.ID,
			Name:     "task in column2",
		})

		tasks, err := repo.GetAllByColumnID(ctx, column1.ID, params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(tasks) != 1 {
			t.Fatalf("expected 1 task, got %v", len(tasks))
		}
		if tasks[0].ColumnID != column1.ID {
			t.Errorf("expected ColumnID %v, got %v", column1.ID, tasks[0].ColumnID)
		}
	})
}

func TestTaskRepo_GetAllByColumnID_Empty(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		params := createTestParams(5, 0, "created_at", "ASC")
		repo := NewTaskRepository(db)

		tasks, err := repo.GetAllByColumnID(ctx, uuid.New(), params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(tasks) != 0 {
			t.Fatalf("expected 0 tasks, got %v", len(tasks))
		}
	})
}

func TestTaskRepo_GetAllByAssigneeID(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		assignee := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column := createTestColumn(ctx, db, board.ID)
		params := createTestParams(5, 0, "created_at", "ASC")
		repo := NewTaskRepository(db)

		_ = repo.Create(ctx, &domain.Task{
			ID:         uuid.New(),
			OwnerID:    user.ID,
			ColumnID:   column.ID,
			AssigneeID: &assignee.ID,
			Name:       "assigned task",
		})
		_ = repo.Create(ctx, &domain.Task{
			ID:       uuid.New(),
			OwnerID:  user.ID,
			ColumnID: column.ID,
			Name:     "unassigned task",
		})

		tasks, err := repo.GetAllByAssigneeID(ctx, assignee.ID, params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(tasks) != 1 {
			t.Fatalf("expected 1 task, got %v", len(tasks))
		}
		if *tasks[0].AssigneeID != assignee.ID {
			t.Errorf("expected AssigneeID %v, got %v", assignee.ID, tasks[0].AssigneeID)
		}
	})
}

func TestTaskRepo_GetAllByAssigneeID_Empty(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		params := createTestParams(5, 0, "created_at", "ASC")
		repo := NewTaskRepository(db)

		tasks, err := repo.GetAllByAssigneeID(ctx, uuid.New(), params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(tasks) != 0 {
			t.Fatalf("expected 0 tasks, got %v", len(tasks))
		}
	})
}

func TestTaskRepo_GetAllByAssigneeID_WithLimit(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		assignee := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column := createTestColumn(ctx, db, board.ID)
		params := createTestParams(3, 0, "created_at", "ASC")
		repo := NewTaskRepository(db)

		for i := 0; i < 5; i++ {
			_ = repo.Create(ctx, &domain.Task{
				ID:         uuid.New(),
				OwnerID:    user.ID,
				ColumnID:   column.ID,
				AssigneeID: &assignee.ID,
				Name:       "assigned task",
			})
		}

		tasks, err := repo.GetAllByAssigneeID(ctx, assignee.ID, params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(tasks) != 3 {
			t.Fatalf("expected 3 tasks, got %v", len(tasks))
		}
	})
}

func TestTaskRepo_GetAllByAssigneeID_WithSortByAndOrder(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		assignee := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column := createTestColumn(ctx, db, board.ID)
		params := createTestParams(3, 0, "name", "DESC")
		repo := NewTaskRepository(db)

		for i := 0; i < 5; i++ {
			_ = repo.Create(ctx, &domain.Task{
				ID:         uuid.New(),
				OwnerID:    user.ID,
				ColumnID:   column.ID,
				AssigneeID: &assignee.ID,
				Name:       fmt.Sprintf("assigned task%d", i),
			})
		}

		tasks, err := repo.GetAllByAssigneeID(ctx, assignee.ID, params)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if tasks[0].Name != "assigned task4" {
			t.Fatalf("expected assigned task4, got %v", tasks[0].Name)
		}
	})
}

func TestTaskRepo_Update(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column1 := createTestColumn(ctx, db, board.ID)
		column2 := createTestColumn(ctx, db, board.ID)
		task := createTestTask(ctx, db, user.ID, column1.ID)
		assignee := createTestUser(ctx, db)
		repo := NewTaskRepository(db)

		desc := "new description"
		dueDate := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Microsecond)

		task.ColumnID = column2.ID
		task.AssigneeID = &assignee.ID
		task.Name = "newtask"
		task.Description = &desc
		task.DueDate = &dueDate

		err := repo.Update(ctx, task)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		found, _ := repo.GetByID(ctx, task.ID)
		if found.ColumnID != column2.ID {
			t.Errorf("expected ColumnID %v, got %v", column2.ID, found.ColumnID)
		}
		if found.AssigneeID == nil || *found.AssigneeID != assignee.ID {
			t.Errorf("expected AssigneeID %v, got %v", assignee.ID, found.AssigneeID)
		}
		if found.Name != "newtask" {
			t.Errorf("expected Name %v, got %v", "newtask", found.Name)
		}
		if found.Description == nil || *found.Description != desc {
			t.Errorf("expected Description %v, got %v", desc, found.Description)
		}
		if found.DueDate == nil || !found.DueDate.Equal(dueDate) {
			t.Errorf("expected DueDate %v, got %v", dueDate, found.DueDate)
		}
	})
}

func TestTaskRepo_Update_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewTaskRepository(db)

		err := repo.Update(ctx, &domain.Task{
			ID:       uuid.New(),
			ColumnID: uuid.New(),
			Name:     "task",
		})
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestTaskRepo_Delete(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		user := createTestUser(ctx, db)
		board := createTestBoard(ctx, db, user.ID)
		column := createTestColumn(ctx, db, board.ID)
		task := createTestTask(ctx, db, user.ID, column.ID)
		repo := NewTaskRepository(db)

		err := repo.Delete(ctx, task.ID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = repo.GetByID(ctx, task.ID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected task to be deleted, got %v", err)
		}
	})
}

func TestTaskRepo_Delete_NotFound(t *testing.T) {
	withTx(t, func(ctx context.Context, db DB) {
		repo := NewTaskRepository(db)

		err := repo.Delete(ctx, uuid.New())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})
}

package pgrepo

import (
	"context"
	"testing"
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

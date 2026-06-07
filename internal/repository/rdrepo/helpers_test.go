package rdrepo

import (
	"context"
	"testing"
)

func teardown(t *testing.T) {
	t.Helper()
	testClient.FlushDB(context.Background())
}

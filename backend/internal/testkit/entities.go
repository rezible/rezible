package testkit

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/rezible/rezible/ent"
)

var seq atomic.Int64

func next(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, seq.Add(1))
}

func CreateUser(t *testing.T, db *ent.Client, ctx context.Context) *ent.User {
	t.Helper()
	u, err := db.User.Create().SetEmail(next("user") + "@example.com").SetName("Test User").Save(ctx)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

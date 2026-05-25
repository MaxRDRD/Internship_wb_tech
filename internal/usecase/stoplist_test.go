package usecase

import (
	"context"
	"testing"

	"topq/internal/adapters/memory"
)

func TestStopListAddRemove(t *testing.T) {
	repo := memory.NewStopListRepo()
	stop := NewStopList(repo)
	ctx := context.Background()

	if err := stop.Add(ctx, "  Iphone 15 "); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	ok, err := stop.Contains(ctx, "iphone 15")
	if err != nil {
		t.Fatalf("Contains failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected query to be blocked")
	}

	if err := stop.Remove(ctx, "iphone 15"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	ok, err = stop.Contains(ctx, "iphone 15")
	if err != nil {
		t.Fatalf("Contains failed: %v", err)
	}
	if ok {
		t.Fatalf("expected query to be unblocked")
	}
}

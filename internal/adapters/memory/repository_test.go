package memory

import (
	"context"
	"testing"
	"time"

	"topq/internal/domain"
)

func TestSlidingWindowRepoCounts(t *testing.T) {
	repo := NewSlidingWindowRepo(300)
	ctx := context.Background()
	now := time.Now().UTC()

	events := []domain.SearchEvent{
		{Query: "iphone 15", OccurredAt: now},
		{Query: "iphone 15", OccurredAt: now},
		{Query: "airpods", OccurredAt: now},
	}

	for _, event := range events {
		if err := repo.AddEvent(ctx, event); err != nil {
			t.Fatalf("AddEvent failed: %v", err)
		}
	}

	items, err := repo.GetTopN(ctx, 2)
	if err != nil {
		t.Fatalf("GetTopN failed: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if items[0].Query != "iphone 15" || items[0].Count != 2 {
		t.Fatalf("unexpected top item: %+v", items[0])
	}
}

func TestSlidingWindowRepoExpiresOldEvents(t *testing.T) {
	repo := NewSlidingWindowRepo(2)
	ctx := context.Background()
	now := time.Now().UTC()

	oldEvent := domain.SearchEvent{Query: "iphone 15", OccurredAt: now.Add(-10 * time.Second)}
	if err := repo.AddEvent(ctx, oldEvent); err != nil {
		t.Fatalf("AddEvent failed: %v", err)
	}

	items, err := repo.GetTopN(ctx, 10)
	if err != nil {
		t.Fatalf("GetTopN failed: %v", err)
	}

	if len(items) != 0 {
		t.Fatalf("expected empty items, got %v", items)
	}
}

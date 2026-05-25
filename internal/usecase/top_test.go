package usecase

import (
	"context"
	"testing"
	"time"

	"topq/internal/adapters/memory"
	"topq/internal/domain"
)

func TestTopFiltersStopList(t *testing.T) {
	topRepo := memory.NewSlidingWindowRepo(300)
	stopRepo := memory.NewStopListRepo()
	top := NewTop(topRepo, stopRepo)
	stop := NewStopList(stopRepo)

	ctx := context.Background()

	events := []domain.SearchEvent{
		{Query: "iphone 15", OccurredAt: time.Now().UTC()},
		{Query: "airpods", OccurredAt: time.Now().UTC()},
	}

	for _, event := range events {
		if err := topRepo.AddEvent(ctx, event); err != nil {
			t.Fatalf("AddEvent failed: %v", err)
		}
	}

	if err := stop.Add(ctx, "airpods"); err != nil {
		t.Fatalf("Add stop-list failed: %v", err)
	}

	items, err := top.Get(ctx, 10)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(items) != 1 || items[0].Query != "iphone 15" {
		t.Fatalf("unexpected items: %v", items)
	}
}

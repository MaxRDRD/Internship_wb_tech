package usecase

import (
	"context"
	"testing"
	"time"

	"topq/internal/adapters/memory"
	"topq/internal/domain"
)

func TestIngestRespectsStopList(t *testing.T) {
	topRepo := memory.NewSlidingWindowRepo(300)
	stopRepo := memory.NewStopListRepo()
	stop := NewStopList(stopRepo)
	ingest := NewIngest(topRepo, stopRepo, nil)

	ctx := context.Background()

	if err := stop.Add(ctx, "iphone 15"); err != nil {
		t.Fatalf("Add stop-list failed: %v", err)
	}

	event := domain.SearchEvent{Query: "iPhone 15", OccurredAt: time.Now().UTC()}
	if err := ingest.Handle(ctx, event); err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	items, err := topRepo.GetTopN(ctx, 10)
	if err != nil {
		t.Fatalf("GetTopN failed: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty top list, got %v", items)
	}
}

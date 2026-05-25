package usecase

import (
	"context"

	"topq/internal/domain"
	"topq/internal/ports"
)

type Ingest struct {
	repo     ports.TopRepository
	stopRepo ports.StopListRepository
}

func NewIngest(repo ports.TopRepository, stopRepo ports.StopListRepository) *Ingest {
	return &Ingest{repo: repo, stopRepo: stopRepo}
}

func (i *Ingest) Handle(ctx context.Context, event domain.SearchEvent) error {
	normalized := normalizeQuery(event.Query)
	if normalized == "" {
		return nil
	}

	if i.stopRepo != nil {
		blocked, err := i.stopRepo.Contains(ctx, normalized)
		if err != nil {
			return err
		}
		if blocked {
			return nil
		}
	}

	event.Query = normalized
	return i.repo.AddEvent(ctx, event)
}

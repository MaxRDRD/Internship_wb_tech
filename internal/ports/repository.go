package ports

import (
	"context"

	"topq/internal/domain"
)

type TopRepository interface {
	AddEvent(ctx context.Context, event domain.SearchEvent) error
	GetTopN(ctx context.Context, n int) ([]domain.TopItem, error)
}

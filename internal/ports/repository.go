package ports

import (
	"context"

	"topq/internal/domain"
)

type TopRepository interface {
	AddEvent(ctx context.Context, event domain.SearchEvent) error // Добавление события
	GetTopN(ctx context.Context, n int) ([]domain.TopItem, error) // Получение топа
}

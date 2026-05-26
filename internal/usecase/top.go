package usecase

import (
	"context"

	"topq/internal/domain"
	"topq/internal/ports"
)

// Use case для получения топа запросов
type Top struct {
	repo     ports.TopRepository
	stopRepo ports.StopListRepository
}

// Создание нового use case для получения топа запросов
func NewTop(repo ports.TopRepository, stopRepo ports.StopListRepository) *Top {
	return &Top{repo: repo, stopRepo: stopRepo}
}

// Получение топа запросов
func (t *Top) Get(ctx context.Context, n int) ([]domain.TopItem, error) {
	items, err := t.repo.GetTopN(ctx, n)
	if err != nil {
		return nil, err
	}
	if t.stopRepo == nil || len(items) == 0 {
		return items, nil
	}

	stopList, err := t.stopRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	if len(stopList) == 0 {
		return items, nil
	}

	// Создаем множество для быстрого поиска стоп-слов
	blocked := make(map[string]struct{}, len(stopList))
	for _, query := range stopList {
		blocked[query] = struct{}{}
	}
	// Фильтруем топ, исключая запросы из стоп-листа
	filtered := make([]domain.TopItem, 0, len(items))
	for _, item := range items {
		if _, ok := blocked[normalizeQuery(item.Query)]; ok {
			continue
		}
		filtered = append(filtered, item)
	}

	if n > 0 && len(filtered) > n {
		filtered = filtered[:n]
	}
	return filtered, nil
}

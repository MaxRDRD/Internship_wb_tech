package usecase

import (
	"context"
	"errors"

	"topq/internal/ports"
)

// Use case для работы со стоп-листом
type StopList struct {
	repo ports.StopListRepository
}

// Создание нового use case для работы со стоп-листом
func NewStopList(repo ports.StopListRepository) *StopList {
	return &StopList{repo: repo}
}

// Добавление стоп-слова в стоп-лист
func (s *StopList) Add(ctx context.Context, query string) error {
	normalized := normalizeQuery(query)
	if normalized == "" {
		return errors.New("query is required")
	}
	return s.repo.Add(ctx, normalized)
}

// Удаление стоп-слова из стоп-листа
func (s *StopList) Remove(ctx context.Context, query string) error {
	normalized := normalizeQuery(query)
	if normalized == "" {
		return errors.New("query is required")
	}
	return s.repo.Remove(ctx, normalized)
}

// Получение стоп-листа
func (s *StopList) List(ctx context.Context) ([]string, error) {
	return s.repo.List(ctx)
}

// Проверка, есть ли запрос в стоп-листе
func (s *StopList) Contains(ctx context.Context, query string) (bool, error) {
	normalized := normalizeQuery(query)
	if normalized == "" {
		return false, nil
	}
	return s.repo.Contains(ctx, normalized)
}

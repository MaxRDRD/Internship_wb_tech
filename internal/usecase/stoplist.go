package usecase

import (
	"context"
	"errors"

	"topq/internal/ports"
)

type StopList struct {
	repo ports.StopListRepository
}

func NewStopList(repo ports.StopListRepository) *StopList {
	return &StopList{repo: repo}
}

func (s *StopList) Add(ctx context.Context, query string) error {
	normalized := normalizeQuery(query)
	if normalized == "" {
		return errors.New("query is required")
	}
	return s.repo.Add(ctx, normalized)
}

func (s *StopList) Remove(ctx context.Context, query string) error {
	normalized := normalizeQuery(query)
	if normalized == "" {
		return errors.New("query is required")
	}
	return s.repo.Remove(ctx, normalized)
}

func (s *StopList) List(ctx context.Context) ([]string, error) {
	return s.repo.List(ctx)
}

func (s *StopList) Contains(ctx context.Context, query string) (bool, error) {
	normalized := normalizeQuery(query)
	if normalized == "" {
		return false, nil
	}
	return s.repo.Contains(ctx, normalized)
}

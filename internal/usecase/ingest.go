package usecase

import (
	"context"

	"topq/internal/domain"
	"topq/internal/ports"
)

// Use case для обработки входящих событий поиска
type Ingest struct {
	repo     ports.TopRepository      // репозиторий для сохранения событий и получения топа
	stopRepo ports.StopListRepository // репозиторий для работы со стоп-листом
	antiSpam *AntiSpam                // антиспам для ограничения количества запросов от одной сессии/пользователя
}

// Создание нового use case для обработки событий поиска
func NewIngest(repo ports.TopRepository, stopRepo ports.StopListRepository, antiSpam *AntiSpam) *Ingest {
	return &Ingest{repo: repo, stopRepo: stopRepo, antiSpam: antiSpam}
}

// Ручка для нормализация входящего запроса
func (i *Ingest) Handle(ctx context.Context, event domain.SearchEvent) error {
	normalized := normalizeQuery(event.Query)
	if normalized == "" {
		return nil
	}
	// Проверяем, есть ли запрос в стоп-листе. Если да, то игнорируем его
	if i.stopRepo != nil {
		blocked, err := i.stopRepo.Contains(ctx, normalized)
		if err != nil {
			return err
		}
		if blocked {
			return nil
		}
	}
	// Проверяем антиспам. Если запрос от одной сессии/пользователя превышает лимит, то игнорируем его
	if i.antiSpam != nil {
		sessionKey := selectSessionKey(event)
		if sessionKey != "" && i.antiSpam.Allow(sessionKey, event.OccurredAt) == false {
			return nil
		}
	}

	event.Query = normalized
	return i.repo.AddEvent(ctx, event)
}

// Выбираем ключ для антиспама: сначала сессия, если есть, иначе пользователь, иначе пустой ключ
func selectSessionKey(event domain.SearchEvent) string {
	if event.SessionID != "" {
		return "s:" + event.SessionID
	}
	if event.UserID != "" {
		return "u:" + event.UserID
	}
	return ""
}

package memory

import (
	"context"
	"sort"
	"sync"
)

type StopListRepo struct {
	mu    sync.RWMutex        // мьютекс для защиты доступа к данным
	items map[string]struct{} // множество стоп-слов
}

// Создание нового репозитория для стоп-листа
func NewStopListRepo() *StopListRepo {
	return &StopListRepo{items: make(map[string]struct{})}
}

// Добавление запроса в стоп-лист
func (s *StopListRepo) Add(_ context.Context, query string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[query] = struct{}{}
	return nil
}

// Удаление запроса из стоп-листа
func (s *StopListRepo) Remove(_ context.Context, query string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, query)
	return nil
}

// Проверка состоит слово с стоп-листе
func (s *StopListRepo) Contains(_ context.Context, query string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.items[query]
	return ok, nil
}

// Получение стоп-листа
func (s *StopListRepo) List(_ context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]string, 0, len(s.items))
	for query := range s.items {
		items = append(items, query)
	}

	sort.Strings(items)
	return items, nil
}

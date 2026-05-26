package usecase

import (
	"sync"
	"time"
)

type AntiSpam struct {
	limit  int           // максимальное количество запросов в окне
	window time.Duration // размер окна для подсчета запросов

	mu       sync.Mutex              // мьютекс для защиты доступа к byUser
	byUser   map[string]*spamCounter //map для хранения счетчика запросов и времени сброса для каждой сессии/пользователя
	lastTrim time.Time               // время последнего удаления устаревших записей из byUser
}

// структура для хранения счетчика запросов и времени сброса для каждой сессии/пользователя
type spamCounter struct {
	count int
	reset time.Time
}

// Создание нового антиспамера
func NewAntiSpam(limit int, window time.Duration) *AntiSpam {
	if limit <= 0 {
		limit = 10
	}
	if window <= 0 {
		window = time.Minute
	}

	return &AntiSpam{
		limit:  limit,
		window: window,
		byUser: make(map[string]*spamCounter),
	}
}

// Возвращаем true, если сессия/пользователь не превышает лимит в текущем окне
func (a *AntiSpam) Allow(key string, now time.Time) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry, ok := a.byUser[key]
	if !ok || now.After(entry.reset) {
		a.byUser[key] = &spamCounter{count: 1, reset: now.Add(a.window)}
		a.trim(now)
		return true
	}

	if entry.count >= a.limit {
		return false
	}

	entry.count++
	a.trim(now)
	return true
}

// Удаляем устаревшие записи из byUser, если прошло достаточно времени с последнего удаления
func (a *AntiSpam) trim(now time.Time) {
	if !a.lastTrim.IsZero() && now.Sub(a.lastTrim) < a.window {
		return
	}

	for key, entry := range a.byUser {
		if now.After(entry.reset) {
			delete(a.byUser, key)
		}
	}

	a.lastTrim = now
}

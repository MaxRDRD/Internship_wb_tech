package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"topq/internal/domain"
)

type SlidingWindowRepo struct {
	windowSec int // размер окна в секундах, для которого считается топ

	mu            sync.Mutex // мьютекс для защиты доступа к данным
	currentSecond int64      // текущее время в секундах
	currentBucket int        // индекс текущего bucket

	buckets []map[string]int // слайс из windowSec bucket-ов
	// каждый bucket - это map query->count для событий, произошедших в эту секунду
	totals map[string]int // общая map query->count для всех событий в окне, для быстрого получения топа
}

// Создание нового репозитория с скользящим окном
func NewSlidingWindowRepo(windowSec int) *SlidingWindowRepo {
	if windowSec <= 0 {
		windowSec = 300
	}
	// Инициализируем buckets и totals
	buckets := make([]map[string]int, windowSec)
	for i := range buckets {
		buckets[i] = make(map[string]int)
	}

	return &SlidingWindowRepo{
		windowSec: windowSec,
		buckets:   buckets,
		totals:    make(map[string]int),
	}
}

// Добавление события в репозиторий
func (r *SlidingWindowRepo) AddEvent(_ context.Context, event domain.SearchEvent) error {
	// Блокируем доступ к map
	r.mu.Lock()
	defer r.mu.Unlock()
	// Вычисляем секунду, в которую произошло событие, и продвигаем окно вперед, если нужно
	eventSec := event.OccurredAt.Unix()
	r.advance(eventSec)
	// Добавляем событие в текущий bucket и обновляем общую map totals
	bucket := r.buckets[r.currentBucket]
	bucket[event.Query]++
	r.totals[event.Query]++
	return nil
}

// Получение топ N запросов
func (r *SlidingWindowRepo) GetTopN(_ context.Context, n int) ([]domain.TopItem, error) {
	if n <= 0 {
		return []domain.TopItem{}, nil
	}
	// Блокируем доступ к map
	r.mu.Lock()
	defer r.mu.Unlock()
	// Продвигаем окно вперед, чтобы удалить старые данные и обновить текущий bucket
	nowSec := time.Now().UTC().Unix()
	r.advance(nowSec)
	// Создаем слайс топ-элементов на основе общей map totals
	items := make([]domain.TopItem, 0, len(r.totals))
	for query, count := range r.totals {
		items = append(items, domain.TopItem{Query: query, Count: count})
	}
	// Сортируем топ-элементы по убыванию количества, а при равенстве - по алфавиту
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Query < items[j].Query
		}
		return items[i].Count > items[j].Count
	})

	if n > len(items) {
		n = len(items)
	}
	return items[:n], nil
}

// Продвижение окна вперед, удаление старых данных и обновление текущего bucket-а
func (r *SlidingWindowRepo) advance(nowSec int64) {
	if r.currentSecond == 0 {
		r.currentSecond = nowSec
		// Расчет текущего bucket-а на основе текущего времени
		r.currentBucket = int(nowSec % int64(r.windowSec))
		r.buckets[r.currentBucket] = make(map[string]int)
		return
	}

	if nowSec <= r.currentSecond {
		return
	}
	// Вычисляем, сколько секунд прошло с момента последнего обновления
	delta := nowSec - r.currentSecond
	// Если прошло больше времени, чем размер окна, очищаем все данные
	if delta >= int64(r.windowSec) {
		r.totals = make(map[string]int)
		for i := range r.buckets {
			r.buckets[i] = make(map[string]int)
		}
		r.currentSecond = nowSec
		r.currentBucket = int(nowSec % int64(r.windowSec))
		return
	}

	// Удаляем старые данные и обновляем текущий bucket
	for i := int64(1); i <= delta; i++ {
		// Вычисляем индекс bucket-а, который нужно очистить
		idx := (r.currentBucket + int(i)) % r.windowSec
		for query, count := range r.buckets[idx] {
			// Обновляем общую map totals, вычитая количество из удаляемого bucket-а
			total := r.totals[query] - count
			// Если после вычитания количество стало меньше или равно нуля, удаляем запрос из totals
			if total <= 0 {
				delete(r.totals, query)
				// Иначе обновляем значение в totals
			} else {
				r.totals[query] = total
			}
		}
		r.buckets[idx] = make(map[string]int)
	}
	// Обновляем текущий bucket и время последнего обновления
	r.currentBucket = (r.currentBucket + int(delta)) % r.windowSec
	r.currentSecond = nowSec
}

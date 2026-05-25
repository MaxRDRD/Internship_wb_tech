package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"topq/internal/domain"
)

type SlidingWindowRepo struct {
	windowSec int

	mu            sync.Mutex
	currentSecond int64
	currentBucket int

	buckets []map[string]int
	totals  map[string]int
}

func NewSlidingWindowRepo(windowSec int) *SlidingWindowRepo {
	if windowSec <= 0 {
		windowSec = 300
	}

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

func (r *SlidingWindowRepo) AddEvent(_ context.Context, event domain.SearchEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	eventSec := event.OccurredAt.Unix()
	r.advance(eventSec)

	bucket := r.buckets[r.currentBucket]
	bucket[event.Query]++
	r.totals[event.Query]++
	return nil
}

func (r *SlidingWindowRepo) GetTopN(_ context.Context, n int) ([]domain.TopItem, error) {
	if n <= 0 {
		return []domain.TopItem{}, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	nowSec := time.Now().UTC().Unix()
	r.advance(nowSec)

	items := make([]domain.TopItem, 0, len(r.totals))
	for query, count := range r.totals {
		items = append(items, domain.TopItem{Query: query, Count: count})
	}

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

func (r *SlidingWindowRepo) advance(nowSec int64) {
	if r.currentSecond == 0 {
		r.currentSecond = nowSec
		r.currentBucket = int(nowSec % int64(r.windowSec))
		r.buckets[r.currentBucket] = make(map[string]int)
		return
	}

	if nowSec <= r.currentSecond {
		return
	}

	delta := nowSec - r.currentSecond
	if delta >= int64(r.windowSec) {
		r.totals = make(map[string]int)
		for i := range r.buckets {
			r.buckets[i] = make(map[string]int)
		}
		r.currentSecond = nowSec
		r.currentBucket = int(nowSec % int64(r.windowSec))
		return
	}

	for i := int64(1); i <= delta; i++ {
		idx := (r.currentBucket + int(i)) % r.windowSec
		for query, count := range r.buckets[idx] {
			total := r.totals[query] - count
			if total <= 0 {
				delete(r.totals, query)
			} else {
				r.totals[query] = total
			}
		}
		r.buckets[idx] = make(map[string]int)
	}

	r.currentBucket = (r.currentBucket + int(delta)) % r.windowSec
	r.currentSecond = nowSec
}

package domain

import "time"

type SearchEvent struct {
	Query      string
	OccurredAt time.Time
}

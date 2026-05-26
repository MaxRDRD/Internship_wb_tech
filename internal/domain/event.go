package domain

import "time"

type SearchEvent struct {
	Query      string
	SessionID  string
	UserID     string
	OccurredAt time.Time
}

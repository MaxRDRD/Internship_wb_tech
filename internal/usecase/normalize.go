package usecase

import "strings"

func normalizeQuery(query string) string {
	return strings.ToLower(strings.TrimSpace(query))
}

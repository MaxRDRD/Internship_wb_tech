package usecase

import "strings"

// хелпер функция для нормализации запроса
func normalizeQuery(query string) string {
	return strings.ToLower(strings.TrimSpace(query))
}

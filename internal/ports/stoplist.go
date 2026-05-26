package ports

import "context"

type StopListRepository interface {
	Add(ctx context.Context, query string) error              // добавление стоп-слова
	Remove(ctx context.Context, query string) error           //удаление стоп-слова
	Contains(ctx context.Context, query string) (bool, error) // проверка нахождение стоп-слова в стоп-листе
	List(ctx context.Context) ([]string, error)               //получение стоп-листа
}

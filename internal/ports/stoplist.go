package ports

import "context"

type StopListRepository interface {
	Add(ctx context.Context, query string) error
	Remove(ctx context.Context, query string) error
	Contains(ctx context.Context, query string) (bool, error)
	List(ctx context.Context) ([]string, error)
}

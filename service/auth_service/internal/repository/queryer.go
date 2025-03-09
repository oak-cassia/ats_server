package repository

import "context"

type Queryer interface {
	GetContext(ctx context.Context, dst interface{}, query string, args ...any) error
	SelectContext(ctx context.Context, dst interface{}, query string, args ...any) error
}

package repository

import (
	"context"
	"database/sql"
)

type Queryer interface {
	GetContext(ctx context.Context, dst interface{}, query string, args ...any) error
	SelectContext(ctx context.Context, dst interface{}, query string, args ...any) error
}

type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

package repository

import (
	"context"
	"database/sql"
)

type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	GetContext(ctx context.Context, dst interface{}, query string, args ...any) error
}

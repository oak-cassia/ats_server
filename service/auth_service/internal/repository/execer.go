package repository

import (
	"context"
	"database/sql"
)

type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

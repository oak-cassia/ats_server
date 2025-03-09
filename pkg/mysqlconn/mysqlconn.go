package mysqlconn

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type MySQLConn struct {
	conn *sqlx.DB
}

func New(user, password, host, port, name string) (*MySQLConn, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	xdb := sqlx.NewDb(db, "mysql")
	return &MySQLConn{conn: xdb}, nil
}

func (mc *MySQLConn) Close() error {
	if mc.conn == nil {
		return nil
	}
	return mc.conn.Close()
}

func (mc *MySQLConn) Conn() *sqlx.DB {
	return mc.conn
}

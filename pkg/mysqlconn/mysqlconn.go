package mysqlconn

import (
	"database/sql"
	"fmt"
	"time"
)

type MySQLConn struct {
	conn *sql.DB
}

func New(user, password, host, name string) (*MySQLConn, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, password, host, name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &MySQLConn{conn: db}, nil
}

func (mc *MySQLConn) Close() error {
	if mc.conn == nil {
		return nil
	}
	return mc.conn.Close()
}

func (mc *MySQLConn) Conn() *sql.DB {
	return mc.conn
}

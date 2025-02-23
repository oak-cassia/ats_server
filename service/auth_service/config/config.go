package config

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"redisclient"
	_ "redisclient"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type AppConfig struct {
	DB    *sql.DB
	Redis *redisclient.RedisClient
}

func (ac *AppConfig) Close() {
	_ = ac.DB.Close()
}

func LoadConfig() (*AppConfig, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := mysqlConfig()
	redis := redisConfig()

	return &AppConfig{
		DB:    db,
		Redis: redis,
	}, nil
}

func mysqlConfig() (*sql.DB, error) {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	name := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, host, name)
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

	return db, nil
}

func redisConfig() *redisclient.RedisClient {
	host := os.Getenv("REDIS_HOST")
	password := os.Getenv("REDIS_PASSWORD")
	db := 0
	return redisclient.NewRedisClient(host, password, db)
}

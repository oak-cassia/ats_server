package config

import (
	"github.com/caarlos0/env/v11"
	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	Env  string `env:"ENV" envDefault:"dev"`
	Port int    `env:"PORT" envDefault:"80"`

	DbUser    string `env:"DB_USER"`
	DbPw      string `env:"DB_PASSWORD"`
	DbHost    string `env:"DB_HOST"`
	DbPort    string `env:"DB_PORT"`
	DbName    string `env:"DB_NAME"`
	RedisHost string `env:"REDIS_HOST"`
	RedisPw   string `env:"REDIS_PASSWORD"`
}

func New() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

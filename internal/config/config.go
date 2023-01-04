package config

import (
	"flag"
	"github.com/caarlos0/env"
)

type Config struct {
	Addr string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	//Database string `env:"DATABASE_URI"`
	Database string `env:"DATABASE_URI" envDefault:"host=localhost port=5432 user=postgres password=031995 TimeZone=UTC"`
	Key      string `env:"COOKIES_KEY" envDefault:"V3ry$trongK3y"`
	Accrual  string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"localhost:8081"`
}

var cfg Config

func init() {
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "Server Address")
	flag.StringVar(&cfg.Database, "d", cfg.Database, "DSN for PGSQL")
	flag.StringVar(&cfg.Accrual, "r", cfg.Accrual, "ACCRUAL_SYSTEM_ADDRESS")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "Key string for sign cookies")
}

func New() (Config, error) {
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	flag.Parse()
	return cfg, nil
}

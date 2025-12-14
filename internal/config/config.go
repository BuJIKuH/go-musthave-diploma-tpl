// Package config загружает конфигурацию из .env, переменных окружения и флагов.
package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	AuthSecret           string `env:"AUTH_SECRET"`
}

func (c *Config) String() string {
	return fmt.Sprintf(
		"--a %s --d %s --r %s",
		c.RunAddress,
		c.DatabaseURI,
		c.AccrualSystemAddress,
	)
}

func InitConfig() *Config {
	if err := godotenv.Load(); err != nil {
		fmt.Println("ℹ️ .env not found or not loaded")
	}

	var cfg Config

	flag.StringVar(&cfg.RunAddress, "a", "", "Server address (e.g. :8080)")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "PostgreSQL DSN")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Accrual system base URL")
	flag.Parse()

	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		cfg.RunAddress = envRunAddress
	} else {
		cfg.RunAddress = "localhost:8080"
	}

	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		cfg.DatabaseURI = envDatabaseURI
	} else {
		panic("❌ CONFIG ERROR: DATABASE_URI is required (set via -d flag or DATABASE_URI env/.env)")
	}

	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
		cfg.AccrualSystemAddress = envAccrualSystemAddress
	} else {
		panic("❌ CONFIG ERROR: ACCRUAL_SYSTEM_ADDRESS is required (set via -r flag or ACCRUAL_SYSTEM_ADDRESS env/.env)")
	}

	if envAuthSecret := os.Getenv("AUTH_SECRET"); envAuthSecret != "" {
		cfg.AuthSecret = envAuthSecret
	} else {
		cfg.AuthSecret = "very-hard-secrets"
	}

	return &cfg
}

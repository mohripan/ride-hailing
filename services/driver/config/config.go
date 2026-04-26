package config

import (
	"os"
	"strings"
)

type Config struct {
	Env      string
	HTTPPort string
	Postgres PostgresConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
}

type PostgresConfig struct{ DSN string }
type RedisConfig struct {
	Addr     string
	Password string
}
type KafkaConfig struct{ Brokers []string }

// Load reads configuration from environment variables with sensible defaults
// for local development (matching the docker-compose.yml services).
func Load() *Config {
	return &Config{
		Env:      getEnv("ENV", "development"),
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		Postgres: PostgresConfig{
			DSN: getEnv("DRIVER_POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/ride_hailing_driver?sslmode=disable"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

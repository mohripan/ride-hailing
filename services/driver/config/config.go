package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"ride-hailing/shared/pkg/messaging"
)

type Config struct {
	Env      string
	HTTPPort string
	Postgres PostgresConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	Dapr     DaprConfig
}

type PostgresConfig struct{ DSN string }
type RedisConfig struct {
	Addr     string
	Password string
}
type KafkaConfig struct{ Brokers []string }
type DaprConfig struct {
	HTTPPort        string
	PubSubName      string
	OutboxBatchSize int
	PollInterval    time.Duration
	RetryDelay      time.Duration
	ClaimTimeout    time.Duration
}

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
		Dapr: DaprConfig{
			HTTPPort:        getEnv("DRIVER_DAPR_HTTP_PORT", "3500"),
			PubSubName:      getEnv("DAPR_PUBSUB_NAME", messaging.DefaultPubSubName),
			OutboxBatchSize: mustGetEnvInt("OUTBOX_BATCH_SIZE", 25),
			PollInterval:    mustGetEnvDuration("OUTBOX_POLL_INTERVAL", 2*time.Second),
			RetryDelay:      mustGetEnvDuration("OUTBOX_BASE_RETRY_DELAY", 5*time.Second),
			ClaimTimeout:    mustGetEnvDuration("OUTBOX_CLAIM_TIMEOUT", 30*time.Second),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		panic(fmt.Sprintf("invalid %s value %q: %v", key, raw, err))
	}
	return value
}

func mustGetEnvDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		panic(fmt.Sprintf("invalid %s value %q: %v", key, raw, err))
	}
	return value
}

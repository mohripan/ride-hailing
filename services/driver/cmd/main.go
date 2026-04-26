package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ride-hailing/services/driver/config"
	"ride-hailing/services/driver/internal/application"
	"ride-hailing/services/driver/internal/infrastructure/postgres"
	infraredis "ride-hailing/services/driver/internal/infrastructure/redis"
	handler "ride-hailing/services/driver/internal/interfaces/http"
	"ride-hailing/shared/pkg/daprpubsub"
	sharedlogger "ride-hailing/shared/pkg/logger"
	"ride-hailing/shared/pkg/outbox"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// @title           Driver Service API
// @version         1.0
// @description     Ride-hailing driver service
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	cfg := config.Load()

	logger, err := sharedlogger.New(cfg.Env)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync() //nolint:errcheck

	// ── Postgres ──────────────────────────────────────────────────────────────
	db, err := sqlx.Connect("postgres", cfg.Postgres.DSN)
	if err != nil {
		logger.Fatal("postgres connect failed", zap.Error(err))
	}
	defer db.Close()
	logger.Info("connected to postgres")

	// ── Redis ─────────────────────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("redis connect failed", zap.Error(err))
	}
	defer rdb.Close()
	logger.Info("connected to redis")

	// ── Wire up infrastructure → application → HTTP ───────────────────────────
	driverRepo := postgres.NewRepository(db)
	driverCache := infraredis.NewCache(rdb)
	outboxStore := outbox.NewStore(db)
	publisher := daprpubsub.NewPublisher(cfg.Dapr.HTTPPort, cfg.Dapr.PubSubName)
	relay := outbox.NewRelay("driver-service", outboxStore, publisher, logger, outbox.RelayConfig{
		BatchSize:      cfg.Dapr.OutboxBatchSize,
		PollInterval:   cfg.Dapr.PollInterval,
		BaseRetryDelay: cfg.Dapr.RetryDelay,
		ClaimTimeout:   cfg.Dapr.ClaimTimeout,
	})

	driverSvc := application.NewDriverService(driverRepo, driverCache, logger)

	h := handler.NewHandler(driverSvc, logger)
	r := handler.NewRouter(h, logger)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := relay.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("outbox relay stopped", zap.Error(err))
		}
	}()

	go func() {
		logger.Info("driver-service started", zap.String("port", cfg.HTTPPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", zap.Error(err))
	}
	logger.Info("driver-service stopped")
}

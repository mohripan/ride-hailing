package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ride-hailing/services/rider/config"
	"ride-hailing/services/rider/internal/application"
	"ride-hailing/services/rider/internal/infrastructure/kafka"
	"ride-hailing/services/rider/internal/infrastructure/postgres"
	infraredis "ride-hailing/services/rider/internal/infrastructure/redis"
	handler "ride-hailing/services/rider/internal/interfaces/http"
	sharedlogger "ride-hailing/shared/pkg/logger"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// @title           Rider Service API
// @version         1.0
// @description     Ride-hailing rider service
// @host            localhost:8081
// @BasePath        /api/v1
func main() {
	cfg := config.Load()

	logger, err := sharedlogger.New(cfg.Env)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()

	db, err := sqlx.Connect("postgres", cfg.Postgres.DSN)
	if err != nil {
		logger.Fatal("postgres connect failed", zap.Error(err))
	}
	defer db.Close()
	logger.Info("connected to postgres")

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("redis connect failed", zap.Error(err))
	}
	defer rdb.Close()
	logger.Info("connected to redis")

	riderRepo := postgres.NewRepository(db)
	riderCache := infraredis.NewCache(rdb)
	producer := kafka.NewProducer(cfg.Kafka.Brokers, logger)
	defer producer.Close()

	riderSvc := application.NewRiderService(riderRepo, riderCache, producer, logger)
	h := handler.NewHandler(riderSvc, logger)
	r := handler.NewRouter(h, logger)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("rider-service started", zap.String("port", cfg.HTTPPort))
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
	logger.Info("rider-service stopped")
}

package outbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Publisher interface {
	Publish(ctx context.Context, msg Message) error
}

type RelayConfig struct {
	BatchSize      int
	PollInterval   time.Duration
	BaseRetryDelay time.Duration
	ClaimTimeout   time.Duration
}

type Relay struct {
	serviceName string
	store       *Store
	publisher   Publisher
	logger      *zap.Logger
	cfg         RelayConfig
}

func NewRelay(serviceName string, store *Store, publisher Publisher, logger *zap.Logger, cfg RelayConfig) *Relay {
	return &Relay{
		serviceName: serviceName,
		store:       store,
		publisher:   publisher,
		logger:      logger,
		cfg:         cfg,
	}
}

func (r *Relay) Run(ctx context.Context) error {
	if err := r.flush(ctx); err != nil && !errors.Is(err, context.Canceled) {
		r.logger.Error("initial outbox flush failed", zap.String("service", r.serviceName), zap.Error(err))
	}

	ticker := time.NewTicker(r.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := r.flush(ctx); err != nil && !errors.Is(err, context.Canceled) {
				r.logger.Error("outbox flush failed", zap.String("service", r.serviceName), zap.Error(err))
			}
		}
	}
}

func (r *Relay) flush(ctx context.Context) error {
	messages, err := r.store.ClaimBatch(ctx, r.cfg.BatchSize, time.Now().Add(-r.cfg.ClaimTimeout))
	if err != nil || len(messages) == 0 {
		return err
	}

	for _, msg := range messages {
		if err := r.publisher.Publish(ctx, msg); err != nil {
			nextAttemptAt := time.Now().Add(retryDelay(msg.Attempts, r.cfg.BaseRetryDelay))
			if markErr := r.store.MarkFailed(ctx, msg.ID, trimError(err), nextAttemptAt); markErr != nil {
				return fmt.Errorf("mark failed outbox message %s: %w", msg.ID, markErr)
			}

			r.logger.Warn(
				"outbox publish failed",
				zap.String("service", r.serviceName),
				zap.String("message_id", msg.ID),
				zap.String("topic", msg.Topic),
				zap.Int("attempts", msg.Attempts),
				zap.Error(err),
			)
			continue
		}

		if err := r.store.MarkPublished(ctx, msg.ID, time.Now().UTC()); err != nil {
			return fmt.Errorf("mark published outbox message %s: %w", msg.ID, err)
		}
	}

	return nil
}

func retryDelay(attempts int, base time.Duration) time.Duration {
	if attempts < 1 {
		return base
	}
	if attempts > 6 {
		attempts = 6
	}

	delay := base
	for i := 1; i < attempts; i++ {
		delay *= 2
	}
	return delay
}

func trimError(err error) string {
	const maxLength = 1024

	if err == nil {
		return ""
	}

	msg := err.Error()
	if len(msg) <= maxLength {
		return msg
	}
	return msg[:maxLength]
}

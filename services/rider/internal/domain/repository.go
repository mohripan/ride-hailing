package domain

import (
	"context"

	"ride-hailing/shared/pkg/messaging"
)

type Repository interface {
	Save(ctx context.Context, r *Rider, messages []messaging.OutboxMessage) error
	Update(ctx context.Context, r *Rider, messages []messaging.OutboxMessage) error
	AppendOutbox(ctx context.Context, messages []messaging.OutboxMessage) error
	FindByID(ctx context.Context, id string) (*Rider, error)
	FindByUserID(ctx context.Context, userID string) (*Rider, error)
}

type Cache interface {
	Set(ctx context.Context, r *Rider) error
	Get(ctx context.Context, id string) (*Rider, error)
	Delete(ctx context.Context, id string) error
}

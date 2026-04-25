package domain

import "context"

type Repository interface {
	Save(ctx context.Context, r *Rider) error
	Update(ctx context.Context, r *Rider) error
	FindByID(ctx context.Context, id string) (*Rider, error)
	FindByUserID(ctx context.Context, userID string) (*Rider, error)
}

type Cache interface {
	Set(ctx context.Context, r *Rider) error
	Get(ctx context.Context, id string) (*Rider, error)
	Delete(ctx context.Context, id string) error
}

type EventPublisher interface {
	PublishRiderRegistered(ctx context.Context, riderID, userID, name string) error
	PublishWalletToppedUp(ctx context.Context, riderID string, amount, balance float64) error
}

package domain

import "context"

// Repository is a PORT — an interface the domain defines.
// The ADAPTER (postgres/driver_repository.go) implements it.
// The domain never imports infrastructure.
type Repository interface {
	Save(ctx context.Context, d *Driver) error
	Update(ctx context.Context, d *Driver) error
	FindByID(ctx context.Context, id string) (*Driver, error)
	FindByUserID(ctx context.Context, userID string) (*Driver, error)
}

// Cache is another port — swappable independently of the DB.
type Cache interface {
	Set(ctx context.Context, d *Driver) error
	Get(ctx context.Context, id string) (*Driver, error)
	Delete(ctx context.Context, id string) error
}

// EventPublisher is the port for outbound async events.
type EventPublisher interface {
	PublishStatusChanged(ctx context.Context, driverID, oldStatus, newStatus string) error
	PublishLocationUpdated(ctx context.Context, driverID string, lat, lng float64) error
}

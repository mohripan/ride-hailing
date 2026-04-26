package application

import (
	"context"

	"ride-hailing/services/driver/internal/application/commands"
	"ride-hailing/services/driver/internal/application/queries"
	"ride-hailing/services/driver/internal/domain"
	"ride-hailing/shared/events"
	"ride-hailing/shared/pkg/messaging"

	"go.uber.org/zap"
)

// DriverService is the application layer — it orchestrates domain objects
// and calls ports (repository, cache). It has NO business logic;
// business logic lives in the domain.
type DriverService struct {
	repo   domain.Repository
	cache  domain.Cache
	logger *zap.Logger
}

func NewDriverService(
	repo domain.Repository,
	cache domain.Cache,
	logger *zap.Logger,
) *DriverService {
	return &DriverService{repo: repo, cache: cache, logger: logger}
}

func (s *DriverService) RegisterDriver(ctx context.Context, cmd commands.RegisterDriverCommand) (*domain.Driver, error) {
	vehicle := domain.Vehicle{
		Make:        cmd.VehicleMake,
		Model:       cmd.VehicleModel,
		Year:        cmd.VehicleYear,
		PlateNumber: cmd.PlateNumber,
		Type:        domain.VehicleType(cmd.VehicleType),
	}

	driver, err := domain.NewDriver(cmd.ID, cmd.UserID, cmd.Name, cmd.Phone, vehicle)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, driver, nil); err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, driver); err != nil {
		s.logger.Warn("cache warm-up failed after register", zap.String("driver_id", driver.ID), zap.Error(err))
	}

	return driver, nil
}

func (s *DriverService) ChangeStatus(ctx context.Context, cmd commands.ChangeStatusCommand) (*domain.Driver, error) {
	driver, err := s.getDriver(ctx, cmd.DriverID)
	if err != nil {
		return nil, err
	}

	oldStatus := string(driver.Status)

	// Business rule lives in the domain, not here
	if err := driver.ChangeStatus(domain.Status(cmd.NewStatus)); err != nil {
		return nil, err
	}

	message, err := messaging.NewOutboxMessage(
		events.TopicDriverStatusChanged,
		driver.ID,
		events.TopicDriverStatusChanged,
		"driver-service",
		events.DriverStatusChanged{
			DriverID:  driver.ID,
			OldStatus: oldStatus,
			NewStatus: string(driver.Status),
			Timestamp: driver.UpdatedAt,
		},
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, driver, []messaging.OutboxMessage{message}); err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, driver); err != nil {
		s.logger.Warn("cache update failed after status change", zap.String("driver_id", driver.ID), zap.Error(err))
	}

	return driver, nil
}

func (s *DriverService) UpdateLocation(ctx context.Context, cmd commands.UpdateLocationCommand) error {
	driver, err := s.getDriver(ctx, cmd.DriverID)
	if err != nil {
		return err
	}

	if err := driver.UpdateLocation(cmd.Latitude, cmd.Longitude); err != nil {
		return err
	}

	message, err := messaging.NewOutboxMessage(
		events.TopicDriverLocationUpdated,
		driver.ID,
		events.TopicDriverLocationUpdated,
		"driver-service",
		events.DriverLocationUpdated{
			DriverID:  driver.ID,
			Latitude:  cmd.Latitude,
			Longitude: cmd.Longitude,
			Timestamp: driver.UpdatedAt,
		},
	)
	if err != nil {
		return err
	}

	// Location is high-frequency (every few seconds per driver), so the durable
	// write goes to the outbox instead of updating the driver table on every ping.
	if err := s.repo.AppendOutbox(ctx, []messaging.OutboxMessage{message}); err != nil {
		return err
	}

	if err := s.cache.Set(ctx, driver); err != nil {
		s.logger.Warn("cache update failed after location update", zap.String("driver_id", driver.ID), zap.Error(err))
	}

	return nil
}

func (s *DriverService) GetDriver(ctx context.Context, qry queries.GetDriverQuery) (*domain.Driver, error) {
	return s.getDriver(ctx, qry.DriverID)
}

// getDriver implements cache-aside: try Redis first, fall back to Postgres.
// Cache errors are transparent — we always return data if the DB has it.
func (s *DriverService) getDriver(ctx context.Context, id string) (*domain.Driver, error) {
	if driver, err := s.cache.Get(ctx, id); err == nil {
		return driver, nil
	}

	driver, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, driver); err != nil {
		s.logger.Warn("cache warm-up failed on read", zap.String("driver_id", id), zap.Error(err))
	}

	return driver, nil
}

package application

import (
	"context"

	"ride-hailing/services/rider/internal/application/commands"
	"ride-hailing/services/rider/internal/application/queries"
	"ride-hailing/services/rider/internal/domain"
	"ride-hailing/shared/events"
	"ride-hailing/shared/pkg/messaging"

	"go.uber.org/zap"
)

type RiderService struct {
	repo   domain.Repository
	cache  domain.Cache
	logger *zap.Logger
}

func NewRiderService(
	repo domain.Repository,
	cache domain.Cache,
	logger *zap.Logger,
) *RiderService {
	return &RiderService{repo: repo, cache: cache, logger: logger}
}

func (s *RiderService) RegisterRider(ctx context.Context, cmd commands.RegisterRiderCommand) (*domain.Rider, error) {
	rider, err := domain.NewRider(cmd.ID, cmd.UserID, cmd.Name, cmd.Phone, cmd.Email)
	if err != nil {
		return nil, err
	}

	message, err := messaging.NewOutboxMessage(
		events.TopicRiderRegistered,
		rider.ID,
		events.TopicRiderRegistered,
		"rider-service",
		events.RiderRegistered{
			RiderID:   rider.ID,
			UserID:    rider.UserID,
			Name:      rider.Name,
			Timestamp: rider.CreatedAt,
		},
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, rider, []messaging.OutboxMessage{message}); err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, rider); err != nil {
		s.logger.Warn("cache warm-up failed after register", zap.String("rider_id", rider.ID), zap.Error(err))
	}

	return rider, nil
}

func (s *RiderService) UpdateProfile(ctx context.Context, cmd commands.UpdateProfileCommand) (*domain.Rider, error) {
	rider, err := s.getRider(ctx, cmd.RiderID)
	if err != nil {
		return nil, err
	}

	if err := rider.UpdateProfile(cmd.Name, cmd.Phone, cmd.Email, cmd.ProfilePhotoURL); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, rider, nil); err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, rider); err != nil {
		s.logger.Warn("cache update failed after profile update", zap.String("rider_id", rider.ID), zap.Error(err))
	}

	return rider, nil
}

func (s *RiderService) TopUpWallet(ctx context.Context, cmd commands.TopUpWalletCommand) (*domain.Rider, error) {
	rider, err := s.getRider(ctx, cmd.RiderID)
	if err != nil {
		return nil, err
	}

	if err := rider.TopUp(cmd.Amount); err != nil {
		return nil, err
	}

	message, err := messaging.NewOutboxMessage(
		events.TopicWalletToppedUp,
		rider.ID,
		events.TopicWalletToppedUp,
		"rider-service",
		events.WalletToppedUp{
			RiderID:   rider.ID,
			Amount:    cmd.Amount,
			Balance:   rider.WalletBalance,
			Timestamp: rider.UpdatedAt,
		},
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, rider, []messaging.OutboxMessage{message}); err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, rider); err != nil {
		s.logger.Warn("cache update failed after top up", zap.String("rider_id", rider.ID), zap.Error(err))
	}

	return rider, nil
}

func (s *RiderService) AddSavedAddress(ctx context.Context, cmd commands.AddSavedAddressCommand) (*domain.Rider, error) {
	rider, err := s.getRider(ctx, cmd.RiderID)
	if err != nil {
		return nil, err
	}

	rider.AddSavedAddress(domain.SavedAddress{
		ID:        cmd.AddressID,
		Label:     domain.AddressLabel(cmd.Label),
		Address:   cmd.Address,
		Latitude:  cmd.Latitude,
		Longitude: cmd.Longitude,
	})

	if err := s.repo.Update(ctx, rider, nil); err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, rider); err != nil {
		s.logger.Warn("cache update failed after add address", zap.String("rider_id", rider.ID), zap.Error(err))
	}

	return rider, nil
}

func (s *RiderService) RemoveSavedAddress(ctx context.Context, cmd commands.RemoveSavedAddressCommand) (*domain.Rider, error) {
	rider, err := s.getRider(ctx, cmd.RiderID)
	if err != nil {
		return nil, err
	}

	if err := rider.RemoveSavedAddress(cmd.AddressID); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, rider, nil); err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, rider); err != nil {
		s.logger.Warn("cache update failed after remove address", zap.String("rider_id", rider.ID), zap.Error(err))
	}

	return rider, nil
}

func (s *RiderService) GetRider(ctx context.Context, qry queries.GetRiderQuery) (*domain.Rider, error) {
	return s.getRider(ctx, qry.RiderID)
}

func (s *RiderService) getRider(ctx context.Context, id string) (*domain.Rider, error) {
	if rider, err := s.cache.Get(ctx, id); err == nil {
		return rider, nil
	}

	rider, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, rider); err != nil {
		s.logger.Warn("cache warm-up failed on read", zap.String("rider_id", id), zap.Error(err))
	}

	return rider, nil
}

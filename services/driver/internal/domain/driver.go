package domain

import "time"

// Location is a value object for GPS coordinates.
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Driver is the aggregate root of this bounded context.
// All business rules about a driver live here.
type Driver struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Name            string    `json:"name"`
	Phone           string    `json:"phone"`
	Status          Status    `json:"status"`
	Vehicle         Vehicle   `json:"vehicle"`
	CurrentLocation *Location `json:"current_location,omitempty"`
	Rating          float64   `json:"rating"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// NewDriver is a factory — the only way to create a valid Driver.
// It enforces required fields and sets sensible defaults.
func NewDriver(id, userID, name, phone string, vehicle Vehicle) (*Driver, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	if phone == "" {
		return nil, ErrInvalidPhone
	}
	now := time.Now()
	return &Driver{
		ID:        id,
		UserID:    userID,
		Name:      name,
		Phone:     phone,
		Status:    StatusOffline, // always starts offline
		Vehicle:   vehicle,
		Rating:    5.0,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ChangeStatus transitions the driver through the status state machine.
// Invalid transitions return a domain error — the caller should NOT
// decide what transitions are valid; only the domain knows that.
func (d *Driver) ChangeStatus(newStatus Status) error {
	if !d.Status.CanTransitionTo(newStatus) {
		return ErrInvalidStatusTransition
	}
	d.Status = newStatus
	d.UpdatedAt = time.Now()
	return nil
}

// UpdateLocation stores the driver's latest GPS position.
// Offline drivers cannot update location — they are not active.
func (d *Driver) UpdateLocation(lat, lng float64) error {
	if d.Status == StatusOffline {
		return ErrDriverOffline
	}
	d.CurrentLocation = &Location{Latitude: lat, Longitude: lng}
	d.UpdatedAt = time.Now()
	return nil
}

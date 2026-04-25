package domain

import "errors"

var (
	ErrDriverNotFound          = errors.New("driver not found")
	ErrDriverAlreadyExists     = errors.New("driver already exists")
	ErrInvalidName             = errors.New("driver name is required")
	ErrInvalidPhone            = errors.New("driver phone is required")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrDriverOffline           = errors.New("driver must be online or on-trip to update location")
)

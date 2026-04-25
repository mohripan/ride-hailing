package domain

import "errors"

var (
	ErrRiderNotFound      = errors.New("rider not found")
	ErrRiderAlreadyExists = errors.New("rider already exists")
	ErrInvalidName        = errors.New("rider name is required")
	ErrInvalidPhone       = errors.New("rider phone is required")
	ErrInsufficientWallet = errors.New("insufficient wallet balance")
	ErrInvalidTopUpAmount = errors.New("top-up amount must be greater than zero")
	ErrAddressNotFound    = errors.New("address not found")
)

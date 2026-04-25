package domain

import "time"

// Rider is the aggregate root of this bounded context.
type Rider struct {
	ID              string         `json:"id"`
	UserID          string         `json:"user_id"`
	Name            string         `json:"name"`
	Phone           string         `json:"phone"`
	Email           string         `json:"email"`
	ProfilePhotoURL string         `json:"profile_photo_url,omitempty"`
	Rating          float64        `json:"rating"`
	WalletBalance   float64        `json:"wallet_balance"`
	SavedAddresses  []SavedAddress `json:"saved_addresses"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

func NewRider(id, userID, name, phone, email string) (*Rider, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	if phone == "" {
		return nil, ErrInvalidPhone
	}
	now := time.Now()
	return &Rider{
		ID:             id,
		UserID:         userID,
		Name:           name,
		Phone:          phone,
		Email:          email,
		Rating:         5.0,
		WalletBalance:  0,
		SavedAddresses: []SavedAddress{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (r *Rider) UpdateProfile(name, phone, email, photoURL string) error {
	if name == "" {
		return ErrInvalidName
	}
	if phone == "" {
		return ErrInvalidPhone
	}
	r.Name = name
	r.Phone = phone
	r.Email = email
	r.ProfilePhotoURL = photoURL
	r.UpdatedAt = time.Now()
	return nil
}

// TopUp adds funds to the wallet. The payment gateway will call this
// after a successful payment confirmation — for now it's just arithmetic.
func (r *Rider) TopUp(amount float64) error {
	if amount <= 0 {
		return ErrInvalidTopUpAmount
	}
	r.WalletBalance += amount
	r.UpdatedAt = time.Now()
	return nil
}

// Deduct is used by trip-service when a trip completes.
func (r *Rider) Deduct(amount float64) error {
	if amount <= 0 || r.WalletBalance < amount {
		return ErrInsufficientWallet
	}
	r.WalletBalance -= amount
	r.UpdatedAt = time.Now()
	return nil
}

func (r *Rider) AddSavedAddress(addr SavedAddress) {
	r.SavedAddresses = append(r.SavedAddresses, addr)
	r.UpdatedAt = time.Now()
}

func (r *Rider) RemoveSavedAddress(addressID string) error {
	for i, a := range r.SavedAddresses {
		if a.ID == addressID {
			r.SavedAddresses = append(r.SavedAddresses[:i], r.SavedAddresses[i+1:]...)
			r.UpdatedAt = time.Now()
			return nil
		}
	}
	return ErrAddressNotFound
}

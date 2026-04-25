package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ride-hailing/services/rider/internal/domain"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type riderRow struct {
	ID              string    `db:"id"`
	UserID          string    `db:"user_id"`
	Name            string    `db:"name"`
	Phone           string    `db:"phone"`
	Email           string    `db:"email"`
	ProfilePhotoURL string    `db:"profile_photo_url"`
	Rating          float64   `db:"rating"`
	WalletBalance   float64   `db:"wallet_balance"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

type addressRow struct {
	ID        string  `db:"id"`
	RiderID   string  `db:"rider_id"`
	Label     string  `db:"label"`
	Address   string  `db:"address"`
	Latitude  float64 `db:"latitude"`
	Longitude float64 `db:"longitude"`
}

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(ctx context.Context, rider *domain.Rider) error {
	const q = `
        INSERT INTO riders (id, user_id, name, phone, email, profile_photo_url, rating, wallet_balance, created_at, updated_at)
        VALUES (:id, :user_id, :name, :phone, :email, :profile_photo_url, :rating, :wallet_balance, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, q, toRow(rider))
	return err
}

func (r *Repository) Update(ctx context.Context, rider *domain.Rider) error {
	// Update rider in a transaction so addresses stay in sync
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const updateRider = `
        UPDATE riders SET
            name              = :name,
            phone             = :phone,
            email             = :email,
            profile_photo_url = :profile_photo_url,
            rating            = :rating,
            wallet_balance    = :wallet_balance,
            updated_at        = :updated_at
        WHERE id = :id`
	if _, err := tx.NamedExecContext(ctx, updateRider, toRow(rider)); err != nil {
		return err
	}

	// Simplest correct approach: delete all addresses and re-insert.
	// For a service at this scale this is perfectly fine.
	if _, err := tx.ExecContext(ctx, `DELETE FROM saved_addresses WHERE rider_id = $1`, rider.ID); err != nil {
		return err
	}
	for _, addr := range rider.SavedAddresses {
		const insertAddr = `
            INSERT INTO saved_addresses (id, rider_id, label, address, latitude, longitude)
            VALUES ($1, $2, $3, $4, $5, $6)`
		if _, err := tx.ExecContext(ctx, insertAddr, addr.ID, rider.ID, string(addr.Label), addr.Address, addr.Latitude, addr.Longitude); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) FindByID(ctx context.Context, id string) (*domain.Rider, error) {
	var row riderRow
	err := r.db.GetContext(ctx, &row, `SELECT * FROM riders WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRiderNotFound
		}
		return nil, err
	}
	return r.hydrate(ctx, row)
}

func (r *Repository) FindByUserID(ctx context.Context, userID string) (*domain.Rider, error) {
	var row riderRow
	err := r.db.GetContext(ctx, &row, `SELECT * FROM riders WHERE user_id = $1`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRiderNotFound
		}
		return nil, err
	}
	return r.hydrate(ctx, row)
}

// hydrate loads the rider's saved addresses after fetching the main row.
func (r *Repository) hydrate(ctx context.Context, row riderRow) (*domain.Rider, error) {
	var addrRows []addressRow
	if err := r.db.SelectContext(ctx, &addrRows, `SELECT * FROM saved_addresses WHERE rider_id = $1`, row.ID); err != nil {
		return nil, err
	}

	rider := toRider(row)
	for _, a := range addrRows {
		rider.SavedAddresses = append(rider.SavedAddresses, domain.SavedAddress{
			ID:        a.ID,
			Label:     domain.AddressLabel(a.Label),
			Address:   a.Address,
			Latitude:  a.Latitude,
			Longitude: a.Longitude,
		})
	}
	return rider, nil
}

func toRow(r *domain.Rider) riderRow {
	return riderRow{
		ID:              r.ID,
		UserID:          r.UserID,
		Name:            r.Name,
		Phone:           r.Phone,
		Email:           r.Email,
		ProfilePhotoURL: r.ProfilePhotoURL,
		Rating:          r.Rating,
		WalletBalance:   r.WalletBalance,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

func toRider(row riderRow) *domain.Rider {
	return &domain.Rider{
		ID:              row.ID,
		UserID:          row.UserID,
		Name:            row.Name,
		Phone:           row.Phone,
		Email:           row.Email,
		ProfilePhotoURL: row.ProfilePhotoURL,
		Rating:          row.Rating,
		WalletBalance:   row.WalletBalance,
		SavedAddresses:  []domain.SavedAddress{},
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

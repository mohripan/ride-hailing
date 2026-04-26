package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ride-hailing/services/driver/internal/domain"
	"ride-hailing/shared/pkg/messaging"
	"ride-hailing/shared/pkg/outbox"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // registers the postgres driver
)

// driverRow is the DB data-transfer object. It mirrors the drivers table
// column-for-column. Keeping it separate from domain.Driver means a schema
// change never bleeds into domain code.
type driverRow struct {
	ID           string    `db:"id"`
	UserID       string    `db:"user_id"`
	Name         string    `db:"name"`
	Phone        string    `db:"phone"`
	Status       string    `db:"status"`
	VehicleMake  string    `db:"vehicle_make"`
	VehicleModel string    `db:"vehicle_model"`
	VehicleYear  int       `db:"vehicle_year"`
	PlateNumber  string    `db:"plate_number"`
	VehicleType  string    `db:"vehicle_type"`
	LocationLat  *float64  `db:"location_lat"`
	LocationLng  *float64  `db:"location_lng"`
	Rating       float64   `db:"rating"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(ctx context.Context, d *domain.Driver, messages []messaging.OutboxMessage) error {
	const q = `
		INSERT INTO drivers (
			id, user_id, name, phone, status,
			vehicle_make, vehicle_model, vehicle_year, plate_number, vehicle_type,
			location_lat, location_lng, rating, created_at, updated_at
		) VALUES (
			:id, :user_id, :name, :phone, :status,
			:vehicle_make, :vehicle_model, :vehicle_year, :plate_number, :vehicle_type,
			:location_lat, :location_lng, :rating, :created_at, :updated_at
		)`
	if len(messages) == 0 {
		_, err := r.db.NamedExecContext(ctx, q, toRow(d))
		return err
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.NamedExecContext(ctx, q, toRow(d)); err != nil {
		return err
	}
	if err := outbox.InsertTx(ctx, tx, messages); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) Update(ctx context.Context, d *domain.Driver, messages []messaging.OutboxMessage) error {
	const q = `
		UPDATE drivers SET
			name          = :name,
			phone         = :phone,
			status        = :status,
			vehicle_make  = :vehicle_make,
			vehicle_model = :vehicle_model,
			vehicle_year  = :vehicle_year,
			plate_number  = :plate_number,
			vehicle_type  = :vehicle_type,
			location_lat  = :location_lat,
			location_lng  = :location_lng,
			rating        = :rating,
			updated_at    = :updated_at
		WHERE id = :id`
	if len(messages) == 0 {
		_, err := r.db.NamedExecContext(ctx, q, toRow(d))
		return err
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.NamedExecContext(ctx, q, toRow(d)); err != nil {
		return err
	}
	if err := outbox.InsertTx(ctx, tx, messages); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) AppendOutbox(ctx context.Context, messages []messaging.OutboxMessage) error {
	if len(messages) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := outbox.InsertTx(ctx, tx, messages); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) FindByID(ctx context.Context, id string) (*domain.Driver, error) {
	var row driverRow
	err := r.db.GetContext(ctx, &row, `SELECT * FROM drivers WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrDriverNotFound
		}
		return nil, err
	}
	return toDriver(row), nil
}

func (r *Repository) FindByUserID(ctx context.Context, userID string) (*domain.Driver, error) {
	var row driverRow
	err := r.db.GetContext(ctx, &row, `SELECT * FROM drivers WHERE user_id = $1`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrDriverNotFound
		}
		return nil, err
	}
	return toDriver(row), nil
}

// toRow maps the domain aggregate to a flat DB row.
func toRow(d *domain.Driver) driverRow {
	row := driverRow{
		ID:           d.ID,
		UserID:       d.UserID,
		Name:         d.Name,
		Phone:        d.Phone,
		Status:       string(d.Status),
		VehicleMake:  d.Vehicle.Make,
		VehicleModel: d.Vehicle.Model,
		VehicleYear:  d.Vehicle.Year,
		PlateNumber:  d.Vehicle.PlateNumber,
		VehicleType:  string(d.Vehicle.Type),
		Rating:       d.Rating,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
	if d.CurrentLocation != nil {
		row.LocationLat = &d.CurrentLocation.Latitude
		row.LocationLng = &d.CurrentLocation.Longitude
	}
	return row
}

// toDriver maps a flat DB row back into the domain aggregate.
func toDriver(row driverRow) *domain.Driver {
	d := &domain.Driver{
		ID:     row.ID,
		UserID: row.UserID,
		Name:   row.Name,
		Phone:  row.Phone,
		Status: domain.Status(row.Status),
		Vehicle: domain.Vehicle{
			Make:        row.VehicleMake,
			Model:       row.VehicleModel,
			Year:        row.VehicleYear,
			PlateNumber: row.PlateNumber,
			Type:        domain.VehicleType(row.VehicleType),
		},
		Rating:    row.Rating,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.LocationLat != nil && row.LocationLng != nil {
		d.CurrentLocation = &domain.Location{
			Latitude:  *row.LocationLat,
			Longitude: *row.LocationLng,
		}
	}
	return d
}

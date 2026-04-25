package events

import "time"

const (
	TopicDriverStatusChanged   = "driver.status.changed"
	TopicDriverLocationUpdated = "driver.location.updated"
)

type DriverStatusChanged struct {
	DriverID  string    `json:"driver_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	Timestamp time.Time `json:"timestamp"`
}

type DriverLocationUpdated struct {
	DriverID  string    `json:"driver_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
}

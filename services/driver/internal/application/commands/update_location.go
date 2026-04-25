package commands

// UpdateLocationCommand carries a new GPS fix from the driver's device.
type UpdateLocationCommand struct {
	DriverID  string
	Latitude  float64
	Longitude float64
}

package domain

// VehicleType constrains what kinds of vehicles are allowed.
type VehicleType string

const (
	VehicleTypeMotorcycle VehicleType = "motorcycle"
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeVan        VehicleType = "van"
)

// Vehicle is a value object — it has no identity of its own,
// it only makes sense as part of a Driver.
type Vehicle struct {
	Make        string      `json:"make"`
	Model       string      `json:"model"`
	Year        int         `json:"year"`
	PlateNumber string      `json:"plate_number"`
	Type        VehicleType `json:"type"`
}

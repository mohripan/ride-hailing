package commands

// RegisterDriverCommand carries everything needed to onboard a new driver.
type RegisterDriverCommand struct {
	ID           string
	UserID       string
	Name         string
	Phone        string
	VehicleMake  string
	VehicleModel string
	VehicleYear  int
	PlateNumber  string
	VehicleType  string
}

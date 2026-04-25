package commands

type AddSavedAddressCommand struct {
	RiderID   string
	AddressID string
	Label     string
	Address   string
	Latitude  float64
	Longitude float64
}

type RemoveSavedAddressCommand struct {
	RiderID   string
	AddressID string
}

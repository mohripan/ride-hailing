package domain

// AddressLabel categorizes saved addresses for quick access in the UI.
type AddressLabel string

const (
	AddressLabelHome  AddressLabel = "home"
	AddressLabelWork  AddressLabel = "work"
	AddressLabelOther AddressLabel = "other"
)

// SavedAddress is a value object — a named location a rider frequently uses.
type SavedAddress struct {
	ID        string       `json:"id"`
	Label     AddressLabel `json:"label"`
	Address   string       `json:"address"`
	Latitude  float64      `json:"latitude"`
	Longitude float64      `json:"longitude"`
}

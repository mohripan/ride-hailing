package commands

// ChangeStatusCommand carries a requested status transition.
type ChangeStatusCommand struct {
	DriverID  string
	NewStatus string
}

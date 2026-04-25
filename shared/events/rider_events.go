package events

import "time"

const (
	TopicRiderRegistered = "rider.registered"
	TopicWalletToppedUp  = "rider.wallet.topped_up"
)

type RiderRegistered struct {
	RiderID   string    `json:"rider_id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

type WalletToppedUp struct {
	RiderID   string    `json:"rider_id"`
	Amount    float64   `json:"amount"`
	Balance   float64   `json:"balance"`
	Timestamp time.Time `json:"timestamp"`
}

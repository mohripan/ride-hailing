package messaging

const (
	ProcessingStatusSuccess = "SUCCESS"
	ProcessingStatusRetry   = "RETRY"
	ProcessingStatusDrop    = "DROP"
)

type Subscription struct {
	PubsubName      string              `json:"pubsubname"`
	Topic           string              `json:"topic"`
	Route           string              `json:"route,omitempty"`
	Routes          *SubscriptionRoutes `json:"routes,omitempty"`
	Metadata        map[string]string   `json:"metadata,omitempty"`
	DeadLetterTopic string              `json:"deadLetterTopic,omitempty"`
}

type SubscriptionRoutes struct {
	Rules   []SubscriptionRule `json:"rules,omitempty"`
	Default string             `json:"default,omitempty"`
}

type SubscriptionRule struct {
	Match string `json:"match"`
	Path  string `json:"path"`
}

type ProcessingResponse struct {
	Status string `json:"status"`
}

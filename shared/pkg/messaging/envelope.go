package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultPubSubName = "kafka-pubsub"
)

type Envelope struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Version    int             `json:"version"`
	Source     string          `json:"source"`
	Key        string          `json:"key"`
	OccurredAt time.Time       `json:"occurred_at"`
	Data       json.RawMessage `json:"data"`
}

type OutboxMessage struct {
	ID        string
	Topic     string
	Key       string
	EventType string
	Payload   []byte
	CreatedAt time.Time
}

func NewOutboxMessage(topic, key, eventType, source string, payload any) (OutboxMessage, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return OutboxMessage{}, fmt.Errorf("marshal %s payload: %w", eventType, err)
	}

	occurredAt := time.Now().UTC()
	envelope := Envelope{
		ID:         uuid.NewString(),
		Type:       eventType,
		Version:    1,
		Source:     source,
		Key:        key,
		OccurredAt: occurredAt,
		Data:       data,
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return OutboxMessage{}, fmt.Errorf("marshal %s envelope: %w", eventType, err)
	}

	return OutboxMessage{
		ID:        envelope.ID,
		Topic:     topic,
		Key:       key,
		EventType: eventType,
		Payload:   body,
		CreatedAt: occurredAt,
	}, nil
}

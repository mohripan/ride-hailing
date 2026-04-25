package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ride-hailing/shared/events"

	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Producer holds one writer per Kafka topic and implements domain.EventPublisher.
// We alias the external kafka-go package as kafkago because this package is
// also named "kafka" — Go would otherwise have a naming conflict.
type Producer struct {
	writers map[string]*kafkago.Writer
	logger  *zap.Logger
}

func NewProducer(brokers []string, logger *zap.Logger) *Producer {
	makeWriter := func(topic string) *kafkago.Writer {
		return &kafkago.Writer{
			Addr:         kafkago.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafkago.LeastBytes{},
			RequiredAcks: kafkago.RequireOne,
			Async:        false, // synchronous so we can log failures
		}
	}
	return &Producer{
		writers: map[string]*kafkago.Writer{
			events.TopicDriverStatusChanged:   makeWriter(events.TopicDriverStatusChanged),
			events.TopicDriverLocationUpdated: makeWriter(events.TopicDriverLocationUpdated),
		},
		logger: logger,
	}
}

func (p *Producer) Close() {
	for topic, w := range p.writers {
		if err := w.Close(); err != nil {
			p.logger.Warn("failed to close kafka writer", zap.String("topic", topic), zap.Error(err))
		}
	}
}

func (p *Producer) PublishStatusChanged(ctx context.Context, driverID, oldStatus, newStatus string) error {
	return p.publish(ctx, events.TopicDriverStatusChanged, driverID, events.DriverStatusChanged{
		DriverID:  driverID,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Timestamp: time.Now(),
	})
}

func (p *Producer) PublishLocationUpdated(ctx context.Context, driverID string, lat, lng float64) error {
	return p.publish(ctx, events.TopicDriverLocationUpdated, driverID, events.DriverLocationUpdated{
		DriverID:  driverID,
		Latitude:  lat,
		Longitude: lng,
		Timestamp: time.Now(),
	})
}

func (p *Producer) publish(ctx context.Context, topic, key string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event for topic %q: %w", topic, err)
	}
	writer, ok := p.writers[topic]
	if !ok {
		return fmt.Errorf("no writer registered for topic %q", topic)
	}
	return writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(key), // driver ID as key → same driver always hits same partition
		Value: data,
	})
}

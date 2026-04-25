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
		}
	}
	return &Producer{
		writers: map[string]*kafkago.Writer{
			events.TopicRiderRegistered: makeWriter(events.TopicRiderRegistered),
			events.TopicWalletToppedUp:  makeWriter(events.TopicWalletToppedUp),
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

func (p *Producer) PublishRiderRegistered(ctx context.Context, riderID, userID, name string) error {
	return p.publish(ctx, events.TopicRiderRegistered, riderID, events.RiderRegistered{
		RiderID:   riderID,
		UserID:    userID,
		Name:      name,
		Timestamp: time.Now(),
	})
}

func (p *Producer) PublishWalletToppedUp(ctx context.Context, riderID string, amount, balance float64) error {
	return p.publish(ctx, events.TopicWalletToppedUp, riderID, events.WalletToppedUp{
		RiderID:   riderID,
		Amount:    amount,
		Balance:   balance,
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
		Key:   []byte(key),
		Value: data,
	})
}

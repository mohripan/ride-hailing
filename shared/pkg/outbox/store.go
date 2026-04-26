package outbox

import (
	"context"
	"time"

	"ride-hailing/shared/pkg/messaging"

	"github.com/jmoiron/sqlx"
)

type Message struct {
	ID        string    `db:"id"`
	Topic     string    `db:"topic"`
	Key       string    `db:"message_key"`
	EventType string    `db:"event_type"`
	Payload   []byte    `db:"payload"`
	Attempts  int       `db:"attempts"`
	CreatedAt time.Time `db:"created_at"`
}

type Store struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) *Store {
	return &Store{db: db}
}

func InsertTx(ctx context.Context, tx *sqlx.Tx, messages []messaging.OutboxMessage) error {
	const insertMessage = `
		INSERT INTO outbox_messages (
			id, topic, message_key, event_type, payload, created_at, next_attempt_at
		) VALUES ($1, $2, $3, $4, $5, $6, $6)`

	for _, msg := range messages {
		if _, err := tx.ExecContext(ctx, insertMessage, msg.ID, msg.Topic, msg.Key, msg.EventType, msg.Payload, msg.CreatedAt); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) ClaimBatch(ctx context.Context, batchSize int, staleBefore time.Time) ([]Message, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	const claimQuery = `
		WITH claimed AS (
			SELECT id
			FROM outbox_messages
			WHERE published_at IS NULL
			  AND next_attempt_at <= NOW()
			  AND (processing_started_at IS NULL OR processing_started_at < $2)
			ORDER BY created_at
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE outbox_messages AS o
		SET processing_started_at = NOW(),
		    attempts = o.attempts + 1
		FROM claimed
		WHERE o.id = claimed.id
		RETURNING o.id, o.topic, o.message_key, o.event_type, o.payload, o.attempts, o.created_at`

	var messages []Message
	if err := tx.SelectContext(ctx, &messages, claimQuery, batchSize, staleBefore); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *Store) MarkPublished(ctx context.Context, id string, publishedAt time.Time) error {
	const query = `
		UPDATE outbox_messages
		SET published_at = $2,
		    processing_started_at = NULL,
		    last_error = NULL
		WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, id, publishedAt)
	return err
}

func (s *Store) MarkFailed(ctx context.Context, id, lastError string, nextAttemptAt time.Time) error {
	const query = `
		UPDATE outbox_messages
		SET processing_started_at = NULL,
		    last_error = $2,
		    next_attempt_at = $3
		WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, id, lastError, nextAttemptAt)
	return err
}

# Dapr pub/sub and outbox hardening

## What changed

- Outbound event publishing now uses a **transactional outbox** in each service database.
- Driver and rider no longer publish directly to Kafka from request handlers.
- Each service now runs an **in-process outbox relay** that publishes pending records to a local **Dapr sidecar**.
- Dapr is configured with a shared Kafka pub/sub component in `dapr/components/pubsub.yaml`.
- Consumer scaffolding was added in `dapr/subscriptions/` to standardize dead-letter topics, raw payload delivery, and future subscription routes.

## New flow

1. Application logic builds a shared event envelope.
2. Repository methods write domain data and outbox records in the same Postgres transaction when the operation has a DB write.
3. A background relay polls `outbox_messages`, publishes through the local Dapr HTTP sidecar, and marks records as published.
4. Failed publishes stay in the outbox and are retried with backoff.

For driver location updates, the durable write is the outbox record itself. The driver table is still not updated on every GPS ping.

## Files added for this milestone

- `shared/pkg/messaging/`
- `shared/pkg/outbox/`
- `shared/pkg/daprpubsub/`
- `dapr/components/pubsub.yaml`
- `dapr/subscriptions/location-service.yaml`
- `dapr/subscriptions/notification-service.yaml`

## Runtime configuration

`.env.example` now includes:

```env
DRIVER_DAPR_HTTP_PORT=3500
RIDER_DAPR_HTTP_PORT=3501
DAPR_PUBSUB_NAME=kafka-pubsub
OUTBOX_BATCH_SIZE=25
OUTBOX_POLL_INTERVAL=2s
OUTBOX_BASE_RETRY_DELAY=5s
OUTBOX_CLAIM_TIMEOUT=30s
```

## Local run commands

Use the Dapr sidecar run targets for local development:

1. `make up`
2. `make migrate-driver-up`
3. `make migrate-rider-up`
4. `make driver-run-dapr`
5. `make rider-run-dapr`

The sidecar commands use `dapr/components` as the resources path, so Dapr picks up the Kafka pub/sub component automatically.

## Outbox schema

Each service database now has an `outbox_messages` table with:

- `id`
- `topic`
- `message_key`
- `event_type`
- `payload`
- `attempts`
- `next_attempt_at`
- `processing_started_at`
- `published_at`
- `last_error`
- `created_at`

The relay uses:

- `processing_started_at` to recover abandoned in-flight records
- `attempts` plus exponential backoff for retries
- `published_at` to distinguish completed work from pending work

## Shared event envelope

Published messages use a shared JSON envelope from `shared/pkg/messaging`:

```json
{
  "id": "message-id",
  "type": "driver.status.changed",
  "version": 1,
  "source": "driver-service",
  "key": "aggregate-id",
  "occurred_at": "2026-04-26T00:00:00Z",
  "data": { }
}
```

This keeps producer payloads consistent and gives future consumers a stable idempotency key.

## Consumer scaffolding conventions

Future consumer services should:

- subscribe through Dapr declarative subscriptions under `dapr/subscriptions/`
- use `deadLetterTopic` for every subscription
- use `rawPayload: "true"` so consumers receive the shared envelope directly
- treat `Envelope.ID` as the idempotency key
- return one of the Dapr processing statuses from `shared/pkg/messaging`:
  - `SUCCESS`
  - `RETRY`
  - `DROP`

## Notes

- This milestone hardens the **publisher side** now and adds consumer conventions/config scaffolding for future services.
- It does not yet add live consumer applications; those should come with `location`, `matching`, `trip`, or `notification` implementation work.

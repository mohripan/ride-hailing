# Copilot Instructions

## Build and test commands

- Start local infrastructure: `make up`
- Stop local infrastructure: `make down`
- Stream infrastructure logs: `make logs`
- Sync workspace modules: `make tidy`
- Build the driver service: `make driver-build` or `cd services/driver && go build ./...`
- Build the rider service: `make rider-build` or `cd services/rider && go build ./...`
- Build shared packages: `cd shared && go build ./...`
- Run the driver service locally: `make driver-run`
- Run the rider service locally: `make rider-run`
- Run the driver test suite: `make driver-test`
- Run the rider test suite: `make rider-test`
- Run shared package tests: `cd shared && go test ./...`
- Run a single Go test: `cd services/driver && go test ./path/to/package -run TestName` (same pattern for `services/rider` and `shared`)
- Apply local schema migrations: `make migrate-driver-up` and `make migrate-rider-up`

## High-level architecture

- This repo is a Go workspace (`go.work`) with three active modules: `shared`, `services/driver`, and `services/rider`.
- `shared` is only for cross-service code. Right now that means event contracts in `shared/events` and the shared Zap logger factory in `shared/pkg/logger`.
- `services/driver` and `services/rider` use the same layered structure:
  - `cmd/main.go` loads env config, connects to Postgres/Redis/Kafka, wires dependencies, and starts the Gin HTTP server with graceful shutdown.
  - `internal/domain` contains aggregates, business rules, domain errors, and outbound port interfaces.
  - `internal/application` orchestrates use cases with command/query DTOs and calls domain ports; it should not become a second business-logic layer.
  - `internal/infrastructure` contains adapter implementations for Postgres, Redis, and Kafka.
  - `internal/interfaces/http` owns Gin routing, request binding, and HTTP error translation.
- The runtime data flow is HTTP -> application service -> domain aggregate/ports -> infrastructure adapters.
- Postgres is the source of truth, Redis is a cache-aside read model, and Kafka is used for integration events between services.
- Driver location updates are intentionally high-frequency cache/event writes: they update Redis and publish Kafka events, but do not persist every GPS ping to Postgres.
- Rider persistence is aggregate-oriented: updating a rider and its saved addresses happens together in one Postgres transaction.
- The directories `services/location`, `services/matching`, `services/notification`, `services/pricing`, and `services/trip` exist as placeholders but are not implemented yet.

## Key conventions

- Keep business rules in domain constructors and methods (`NewDriver`, `ChangeStatus`, `UpdateLocation`, `NewRider`, `TopUp`, `RemoveSavedAddress`) instead of duplicating them in handlers or application services.
- Domain packages define the repository, cache, and event publisher interfaces; infrastructure packages implement them. Domain code should not import infrastructure.
- HTTP handlers are thin: they bind/validate request payloads, generate UUIDs, build command/query objects, and delegate to the application layer.
- Each service centralizes transport-level error mapping in a `handleError` function inside `internal/interfaces/http/handler.go`.
- Read paths use cache-aside behavior: check Redis first, fall back to Postgres, then warm the cache.
- Cache failures and Kafka publish failures are logged but usually do not fail an otherwise successful write; persistence failures do fail the request.
- Shared event topic names and payload types live in `shared/events`; when adding or changing events, update shared contracts first and keep producer code aligned.
- Each service exposes Swagger from the Gin router at `/swagger/*any`, and the API routes live under `/api/v1`.
- Local defaults are environment-driven and match the compose stack: Postgres on `localhost:5432`, Redis on `localhost:6379`, Kafka on `localhost:9092`; rider defaults to port `8081`, driver defaults to `8080`.

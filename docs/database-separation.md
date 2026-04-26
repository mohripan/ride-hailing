# Service database separation

## What changed

- The local Postgres setup now creates two service-owned databases in the same container:
  - `ride_hailing_driver`
  - `ride_hailing_rider`
- Driver now defaults to `DRIVER_POSTGRES_DSN`.
- Rider now defaults to `RIDER_POSTGRES_DSN`.
- The migration targets in `Makefile` now apply each service's schema to its own database.

## Files updated

- `compose.yaml`
- `Makefile`
- `.env.example`
- `services/driver/config/config.go`
- `services/rider/config/config.go`
- `docker/postgres/init/01_create_service_databases.sql`

## How local setup works now

`compose.yaml` mounts `docker/postgres/init` into Postgres `docker-entrypoint-initdb.d`. On a fresh Postgres data volume, startup creates the `ride_hailing_driver` and `ride_hailing_rider` databases automatically.

Each service keeps using the same Postgres server on `localhost:5432`, but with its own DSN and its own schema migrations.

## How to apply it locally

For a fresh local setup:

1. `make nuke`
2. `make up`
3. `make migrate-driver-up`
4. `make migrate-rider-up`

If you already have an existing Postgres volume, the init SQL will not re-run automatically. In that case either:

1. reset the volume with `make nuke` and start fresh, or
2. create the two databases manually before running the migration targets.

## Expected environment variables

Use these DSNs when overriding defaults:

```env
DRIVER_POSTGRES_DSN=postgres://postgres:postgres@localhost:5432/ride_hailing_driver?sslmode=disable
RIDER_POSTGRES_DSN=postgres://postgres:postgres@localhost:5432/ride_hailing_rider?sslmode=disable
```

Redis and Kafka settings remain shared.

## Boundary rule going forward

Driver data should stay in the driver database and rider data should stay in the rider database. Cross-service reads should happen through APIs, projections, or Kafka-driven read models rather than direct table access.

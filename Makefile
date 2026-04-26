.PHONY: up down nuke logs driver-run driver-run-dapr driver-build driver-test migrate-driver-up rider-run rider-run-dapr rider-build rider-test migrate-rider-up tidy

# ── Infrastructure ────────────────────────────────────────────────────────────
up:
	docker compose up -d

down:
	docker compose down

nuke:
	docker compose down -v

logs:
	docker compose logs -f

# ── Driver Service ────────────────────────────────────────────────────────────
driver-run:
	cd services/driver && go run ./cmd/main.go

driver-run-dapr:
	cd services/driver && dapr run --app-id driver-service --app-port 8080 --dapr-http-port 3500 --resources-path ../../dapr/components -- go run ./cmd/main.go

driver-build:
	cd services/driver && go build -o bin/driver ./cmd/main.go

driver-test:
	cd services/driver && go test ./...

migrate-driver-up:
	docker compose cp services/driver/internal/infrastructure/postgres/migrations/001_create_drivers.sql postgres:/tmp/001.sql
	docker compose exec postgres psql -U postgres -d ride_hailing_driver -f /tmp/001.sql

# ── Rider Service ────────────────────────────────────────────────────────────
rider-run:
	cd services/rider && go run ./cmd/main.go

rider-run-dapr:
	cd services/rider && dapr run --app-id rider-service --app-port 8081 --dapr-http-port 3501 --resources-path ../../dapr/components -- go run ./cmd/main.go

rider-build:
	cd services/rider && go build -o bin/rider ./cmd/main.go

rider-test:
	cd services/rider && go test ./...

migrate-rider-up:
	docker compose cp services/rider/internal/infrastructure/postgres/migrations/001_create_riders.sql postgres:/tmp/rider_001.sql
	docker compose exec postgres psql -U postgres -d ride_hailing_rider -f /tmp/rider_001.sql

# ── Go Workspace ──────────────────────────────────────────────────────────────
tidy:
	cd shared && go mod tidy
	cd services/driver && go mod tidy
	cd services/rider && go mod tidy
	go work sync

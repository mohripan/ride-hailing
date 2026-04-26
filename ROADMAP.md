# Roadmap

## Direction

Build this into a Kafka-native, real-time delivery matching platform: rider demand comes in, driver supply is tracked continuously, matching happens from events and fast read models, and each service owns its data and publishes contracts to the rest of the system.

## Current baseline

- `services/driver` and `services/rider` are the only implemented services today.
- Redis is already positioned as a fast read/cache layer.
- Outbound events now use service-local outbox tables and Dapr pub/sub sidecars in front of Kafka.
- `services/location`, `services/matching`, `services/pricing`, `services/trip`, and `services/notification` are present as placeholders.
- Driver and rider now use separate Postgres databases in the same local Postgres container.

## Recommended sequence

### 1. Split data ownership by service - Done

**Goal:** move from a shared database to service-owned persistence.

**Deliverables**

- Separate Postgres databases for driver and rider in `compose.yaml`
- Separate DSNs and env vars per service
- Service-specific migration flow instead of assuming one shared `ride_hailing` database
- Clear rule that cross-service reads happen through APIs/events, not direct table access
- Migration/backfill plan for existing local data

**Why this comes first**

- It aligns the codebase with the service boundaries that already exist in the domain/application/infrastructure layers.
- It prevents future location, matching, trip, and pricing services from coupling themselves to one shared schema.

### 2. Harden Kafka as the integration backbone - Done

**Goal:** make events reliable enough to support matching and trip orchestration.

**Deliverables**

- Standard topic naming, event keys, and versioning rules in `shared/events`
- Transactional outbox pattern for DB write + event publish consistency
- Consumer idempotency and retry strategy
- Dead-letter topics for failed processing
- Event contract documentation per service

**Why next**

- Right now writes can succeed even if publishing fails. That is acceptable for early CRUD but weak for matching, trip state, and payments.

### 3. Build the location service

**Goal:** turn driver location events into a queryable real-time location index.

**Deliverables**

- `services/location` service with Kafka consumers for driver location updates
- Redis GEO or equivalent spatial index for nearby-driver lookups
- Driver availability projection combining location + online/on-trip status
- APIs/internal queries for nearest available drivers
- Staleness handling for inactive drivers

**Key dependency**

- Depends on driver events being reliable and well-versioned.

### 4. Build the matching service

**Goal:** assign the best available driver to a rider request quickly and deterministically.

**Deliverables**

- `services/matching` service
- Ride request intake and driver candidate selection
- Matching score based on distance, driver status, acceptance window, and future extensible factors
- Reservation/lock flow so the same driver is not matched twice
- Events such as `match.requested`, `match.proposed`, `match.accepted`, `match.expired`, `match.failed`

**Key dependency**

- Depends on a live location index and clean service-owned data boundaries.

### 5. Build the trip service

**Goal:** own trip lifecycle state from request through completion/cancellation.

**Deliverables**

- `services/trip` aggregate and persistence
- Trip lifecycle state machine
- Consumption of matching results and publication of trip events
- Cancellation flows from rider and driver sides
- Hooks for payment capture/settlement later

**Why separate it**

- Matching decides *who gets assigned*; trip owns *what happens after assignment*.

### 6. Build the pricing service

**Goal:** isolate fare estimation and pricing policy from trip orchestration.

**Deliverables**

- `services/pricing` service
- Fare quote API for rider request flow
- Base fare + distance/time model
- Surge or dynamic pricing inputs
- Final fare calculation events/queries for trip completion

### 7. Build the notification service

**Goal:** deliver rider/driver notifications from events rather than synchronous coupling.

**Deliverables**

- `services/notification` consumers for trip, match, and wallet events
- Push/SMS/email adapter abstraction
- Templates for driver assignment, arrival, trip completion, and wallet updates
- Delivery status tracking and retry handling

### 8. Add operational maturity

**Goal:** make the platform safe to evolve once more services are active.

**Deliverables**

- Integration tests around Kafka-driven flows
- Consumer contract tests for shared events
- Tracing/log correlation across services
- Metrics and dashboards for lag, match latency, acceptance rate, and trip throughput
- Local bootstrap commands for all implemented services
- Seed data and developer scripts for end-to-end simulation

## Immediate next milestones

If the goal is to make the next iteration meaningful, do these first:

1. Done - Separate the driver and rider databases.
2. Done - Add outbox-based event publishing so writes and Kafka stay consistent.
3. Implement the location service as the first real event-driven downstream service.
4. Implement the matching service on top of location projections.

## Proposed ownership boundaries

| Service | Owns |
| --- | --- |
| Driver | Driver profile, driver status, vehicle data |
| Rider | Rider profile, wallet, saved addresses |
| Location | Real-time driver location index and nearby-driver queries |
| Matching | Candidate selection, reservation, assignment decisions |
| Trip | Trip lifecycle and state transitions |
| Pricing | Fare quote and final pricing policy |
| Notification | Delivery of outbound user notifications |

## Architecture guardrails for next work

- Do not let services query each other's tables once the database split starts.
- Keep business rules inside domain aggregates and state machines, not handlers.
- Keep shared contracts in `shared/events`, but keep service persistence private.
- Prefer event-driven projections for read models needed by matching and dispatch.
- Treat matching, trip, and pricing as separate bounded contexts even if they are developed incrementally.

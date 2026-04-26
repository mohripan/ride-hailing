-- Driver table: flat columns instead of JSONB for simplicity and queryability.
-- vehicle_* columns store the value object inline.
-- location_lat/lng are nullable — a driver starts with no known location.

CREATE TABLE IF NOT EXISTS drivers (
    id            VARCHAR(36)    PRIMARY KEY,
    user_id       VARCHAR(36)    NOT NULL UNIQUE,
    name          VARCHAR(255)   NOT NULL,
    phone         VARCHAR(50)    NOT NULL UNIQUE,
    status        VARCHAR(20)    NOT NULL DEFAULT 'offline',

    vehicle_make  VARCHAR(100)   NOT NULL,
    vehicle_model VARCHAR(100)   NOT NULL,
    vehicle_year  INTEGER        NOT NULL,
    plate_number  VARCHAR(50)    NOT NULL UNIQUE,
    vehicle_type  VARCHAR(50)    NOT NULL,

    location_lat  DECIMAL(10,8),
    location_lng  DECIMAL(11,8),

    rating        DECIMAL(3,2)   NOT NULL DEFAULT 5.00,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_drivers_user_id ON drivers(user_id);
CREATE INDEX IF NOT EXISTS idx_drivers_status  ON drivers(status);

CREATE TABLE IF NOT EXISTS outbox_messages (
    id                    VARCHAR(36)   PRIMARY KEY,
    topic                 VARCHAR(255)  NOT NULL,
    message_key           VARCHAR(255)  NOT NULL,
    event_type            VARCHAR(255)  NOT NULL,
    payload               JSONB         NOT NULL,
    attempts              INTEGER       NOT NULL DEFAULT 0,
    next_attempt_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    processing_started_at TIMESTAMPTZ,
    published_at          TIMESTAMPTZ,
    last_error            TEXT,
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_driver_outbox_pending
    ON outbox_messages (next_attempt_at, created_at)
    WHERE published_at IS NULL;

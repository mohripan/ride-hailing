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

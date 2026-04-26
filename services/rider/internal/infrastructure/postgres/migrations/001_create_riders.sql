CREATE TABLE IF NOT EXISTS riders (
                                       id                VARCHAR(36)     PRIMARY KEY,
    user_id           VARCHAR(36)     NOT NULL UNIQUE,
    name              VARCHAR(255)    NOT NULL,
    phone             VARCHAR(50)     NOT NULL UNIQUE,
    email             VARCHAR(255)    NOT NULL DEFAULT '',
    profile_photo_url VARCHAR(500)    NOT NULL DEFAULT '',
    rating            DECIMAL(3,2)    NOT NULL DEFAULT 5.00,
    wallet_balance    DECIMAL(12,2)   NOT NULL DEFAULT 0.00,
    created_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ     NOT NULL DEFAULT NOW()
    );

-- Saved addresses stored in a separate table (one-to-many).
-- We don't use JSONB here because we may want to query by label later.
CREATE TABLE IF NOT EXISTS saved_addresses (
                                               id         VARCHAR(36)   PRIMARY KEY,
    rider_id   VARCHAR(36)   NOT NULL REFERENCES riders(id) ON DELETE CASCADE,
    label      VARCHAR(20)   NOT NULL DEFAULT 'other',
    address    VARCHAR(500)  NOT NULL,
    latitude   DECIMAL(10,8) NOT NULL,
    longitude  DECIMAL(11,8) NOT NULL
    );

CREATE INDEX IF NOT EXISTS idx_riders_user_id         ON riders(user_id);
CREATE INDEX IF NOT EXISTS idx_saved_addresses_rider  ON saved_addresses(rider_id);

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

CREATE INDEX IF NOT EXISTS idx_rider_outbox_pending
    ON outbox_messages (next_attempt_at, created_at)
    WHERE published_at IS NULL;

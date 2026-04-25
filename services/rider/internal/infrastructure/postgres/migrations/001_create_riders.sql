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
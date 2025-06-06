-- The `id` field is a 64-bit integer deterministically generated from `email_hash`
-- using a fast non-cryptographic hash function (e.g., xxHash64 or SHA-256 prefix).
-- This eliminates the need for sequence-based ID generation and enables rapid
-- ingestion at scale. While hash collisions are theoretically possible, the chance
-- of two distinct emails producing the same 64-bit ID is astronomically low
-- (less than 1 in 2^63), and the system is designed to tolerate this by simply
-- ignoring such insertions. This trade-off is accepted in favor of maximum performance.
DROP TABLE IF EXISTS users;
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,
    email_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TABLE IF EXISTS users_staging;
CREATE UNLOGGED TABLE IF NOT EXISTS users_staging (
  id BIGINT NOT NULL,
  email_hash TEXT NOT NULL
);

DROP TABLE IF EXISTS track_events;
CREATE TABLE IF NOT EXISTS track_events (
    user_id BIGINT NOT NULL,
    event_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


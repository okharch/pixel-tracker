CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email_hash TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS track_events (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    event_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNLOGGED TABLE IF NOT EXISTS staging_user_events (
    email_hash TEXT NOT NULL,
    event_hash TEXT NOT NULL
);

CREATE OR REPLACE FUNCTION process_staging_user_events()
RETURNS void AS $$
BEGIN
  INSERT INTO users(email_hash)
  SELECT DISTINCT email_hash
  FROM staging_user_events
  ON CONFLICT DO NOTHING;

  INSERT INTO track_events(user_id, event_hash, created_at)
  SELECT u.id, s.event_hash, now()
  FROM staging_user_events s
  JOIN users u ON u.email_hash = s.email_hash;

  TRUNCATE staging_user_events;
END;
$$ LANGUAGE plpgsql;

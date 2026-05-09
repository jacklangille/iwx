CREATE TABLE IF NOT EXISTS outbox_events (
  id BIGSERIAL PRIMARY KEY,
  event_id TEXT NOT NULL UNIQUE,
  event_type TEXT NOT NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  published_at TIMESTAMP WITHOUT TIME ZONE,
  attempt_count BIGINT NOT NULL DEFAULT 0,
  last_error TEXT
);

CREATE INDEX IF NOT EXISTS exchange_core_outbox_events_unpublished_idx
  ON outbox_events (published_at, id)
  WHERE published_at IS NULL;

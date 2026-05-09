CREATE TABLE IF NOT EXISTS position_locks (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  contract_id BIGINT NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
  side TEXT NOT NULL,
  quantity BIGINT NOT NULL,
  status TEXT NOT NULL,
  order_id BIGINT,
  reference_type TEXT NOT NULL DEFAULT '',
  reference_id TEXT NOT NULL DEFAULT '',
  correlation_id TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  released_at TIMESTAMP WITHOUT TIME ZONE
);

CREATE INDEX IF NOT EXISTS position_locks_user_idx
  ON position_locks (user_id, contract_id, side, id DESC);

CREATE INDEX IF NOT EXISTS position_locks_status_idx
  ON position_locks (status, contract_id, id DESC);

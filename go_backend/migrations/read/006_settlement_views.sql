CREATE TABLE IF NOT EXISTS settlement_entries (
  id BIGINT PRIMARY KEY,
  contract_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  entry_type TEXT NOT NULL,
  outcome TEXT NOT NULL,
  amount_cents BIGINT NOT NULL,
  quantity BIGINT NOT NULL,
  reference_id TEXT NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS read_settlement_entries_contract_idx
  ON settlement_entries (contract_id, id DESC);

CREATE INDEX IF NOT EXISTS read_settlement_entries_user_idx
  ON settlement_entries (user_id, id DESC);

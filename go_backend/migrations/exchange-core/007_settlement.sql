CREATE TABLE IF NOT EXISTS settlement_entries (
  id BIGSERIAL PRIMARY KEY,
  contract_id BIGINT NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL,
  entry_type TEXT NOT NULL,
  outcome TEXT NOT NULL,
  amount_cents BIGINT NOT NULL,
  quantity BIGINT NOT NULL DEFAULT 0,
  reference_id TEXT NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS settlement_entries_contract_idx
  ON settlement_entries (contract_id, id DESC);

CREATE INDEX IF NOT EXISTS settlement_entries_user_idx
  ON settlement_entries (user_id, id DESC);

CREATE TABLE IF NOT EXISTS settlement_applications (
  resolution_event_id TEXT PRIMARY KEY,
  contract_id BIGINT NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
  correlation_id TEXT,
  settled_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS settlement_applications_contract_idx
  ON settlement_applications (contract_id, settled_at DESC);

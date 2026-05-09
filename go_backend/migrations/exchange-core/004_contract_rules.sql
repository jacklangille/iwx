CREATE TABLE IF NOT EXISTS contract_rules (
  id BIGSERIAL PRIMARY KEY,
  contract_id BIGINT NOT NULL UNIQUE REFERENCES contracts(id) ON DELETE CASCADE,
  rule_version TEXT NOT NULL,
  metric TEXT NOT NULL,
  threshold BIGINT,
  measurement_unit TEXT,
  resolution_inclusive_side TEXT NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS contract_rules_contract_idx
  ON contract_rules (contract_id);

CREATE TABLE IF NOT EXISTS execution_applications (
  execution_id TEXT PRIMARY KEY,
  contract_id BIGINT NOT NULL,
  buyer_user_id BIGINT NOT NULL,
  seller_user_id BIGINT NOT NULL,
  occurred_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  applied_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS execution_applications_contract_idx
  ON execution_applications (contract_id, applied_at DESC);

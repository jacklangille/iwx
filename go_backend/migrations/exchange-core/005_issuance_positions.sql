CREATE TABLE IF NOT EXISTS issuance_batches (
  id BIGSERIAL PRIMARY KEY,
  contract_id BIGINT NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
  creator_user_id BIGINT NOT NULL,
  collateral_lock_id BIGINT NOT NULL UNIQUE REFERENCES collateral_locks(id),
  above_quantity BIGINT NOT NULL,
  below_quantity BIGINT NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS issuance_batches_contract_idx
  ON issuance_batches (contract_id, status, id DESC);

CREATE INDEX IF NOT EXISTS issuance_batches_creator_idx
  ON issuance_batches (creator_user_id, contract_id, id DESC);

CREATE TABLE IF NOT EXISTS positions (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  contract_id BIGINT NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
  side TEXT NOT NULL,
  available_quantity BIGINT NOT NULL DEFAULT 0,
  locked_quantity BIGINT NOT NULL DEFAULT 0,
  total_quantity BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  CONSTRAINT positions_user_contract_side_key UNIQUE (user_id, contract_id, side),
  CONSTRAINT positions_non_negative_quantities CHECK (
    available_quantity >= 0 AND locked_quantity >= 0 AND total_quantity >= 0
  ),
  CONSTRAINT positions_quantity_invariant CHECK (
    available_quantity + locked_quantity = total_quantity
  )
);

CREATE INDEX IF NOT EXISTS positions_user_contract_idx
  ON positions (user_id, contract_id, side);

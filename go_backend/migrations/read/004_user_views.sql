CREATE TABLE IF NOT EXISTS cash_accounts (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  currency TEXT NOT NULL,
  available_cents BIGINT NOT NULL,
  locked_cents BIGINT NOT NULL,
  total_cents BIGINT NOT NULL,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS read_cash_accounts_user_idx
  ON cash_accounts (user_id, currency);

CREATE TABLE IF NOT EXISTS collateral_locks (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  contract_id BIGINT NOT NULL,
  currency TEXT NOT NULL,
  amount_cents BIGINT NOT NULL,
  status TEXT NOT NULL,
  reference_id TEXT NOT NULL,
  description TEXT NOT NULL,
  reference_issuance_id BIGINT,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  released_at TIMESTAMP WITHOUT TIME ZONE
);

CREATE INDEX IF NOT EXISTS read_collateral_locks_user_idx
  ON collateral_locks (user_id, currency, id DESC);

CREATE TABLE IF NOT EXISTS order_cash_reservations (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  contract_id BIGINT NOT NULL,
  currency TEXT NOT NULL,
  amount_cents BIGINT NOT NULL,
  status TEXT NOT NULL,
  reference_type TEXT NOT NULL,
  reference_id TEXT NOT NULL,
  correlation_id TEXT NOT NULL,
  description TEXT NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  released_at TIMESTAMP WITHOUT TIME ZONE
);

CREATE INDEX IF NOT EXISTS read_order_cash_reservations_user_idx
  ON order_cash_reservations (user_id, currency, id DESC);

CREATE TABLE IF NOT EXISTS positions (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  contract_id BIGINT NOT NULL,
  side TEXT NOT NULL,
  available_quantity BIGINT NOT NULL,
  locked_quantity BIGINT NOT NULL,
  total_quantity BIGINT NOT NULL,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS read_positions_user_contract_idx
  ON positions (user_id, contract_id, side);

CREATE TABLE IF NOT EXISTS position_locks (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  contract_id BIGINT NOT NULL,
  side TEXT NOT NULL,
  quantity BIGINT NOT NULL,
  status TEXT NOT NULL,
  order_id BIGINT,
  reference_type TEXT NOT NULL,
  reference_id TEXT NOT NULL,
  correlation_id TEXT NOT NULL,
  description TEXT NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  released_at TIMESTAMP WITHOUT TIME ZONE
);

CREATE INDEX IF NOT EXISTS read_position_locks_user_idx
  ON position_locks (user_id, contract_id, side, id DESC);

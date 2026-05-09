CREATE TABLE IF NOT EXISTS order_commands (
  command_id TEXT PRIMARY KEY,
  contract_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  token_type TEXT NOT NULL,
  order_side TEXT NOT NULL,
  price TEXT NOT NULL,
  quantity BIGINT NOT NULL,
  status TEXT NOT NULL,
  error_message TEXT,
  result_status TEXT,
  result_order_id BIGINT,
  enqueued_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  started_at TIMESTAMP WITHOUT TIME ZONE,
  completed_at TIMESTAMP WITHOUT TIME ZONE,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS order_commands_contract_status_idx
  ON order_commands (contract_id, status, updated_at);

CREATE TABLE IF NOT EXISTS orders (
  id BIGSERIAL PRIMARY KEY,
  contract_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  token_type TEXT NOT NULL,
  order_side TEXT NOT NULL,
  price NUMERIC(10, 2) NOT NULL,
  quantity BIGINT NOT NULL,
  status TEXT NOT NULL,
  inserted_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS orders_match_lookup_idx
  ON orders (contract_id, token_type, order_side, status, price, inserted_at);

CREATE INDEX IF NOT EXISTS orders_contract_status_idx
  ON orders (contract_id, status, inserted_at);

CREATE TABLE IF NOT EXISTS market_snapshots (
  id BIGSERIAL PRIMARY KEY,
  contract_id BIGINT NOT NULL,
  best_above NUMERIC(10, 2),
  best_below NUMERIC(10, 2),
  mid_above NUMERIC(10, 2),
  mid_below NUMERIC(10, 2),
  inserted_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS market_snapshots_contract_inserted_at_idx
  ON market_snapshots (contract_id, inserted_at);

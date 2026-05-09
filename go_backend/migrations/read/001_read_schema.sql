CREATE TABLE IF NOT EXISTS contracts (
  id BIGINT PRIMARY KEY,
  creator_user_id BIGINT NOT NULL,
  name TEXT NOT NULL,
  region TEXT NOT NULL,
  metric TEXT NOT NULL,
  status TEXT NOT NULL,
  threshold BIGINT,
  multiplier BIGINT,
  measurement_unit TEXT,
  trading_period_start DATE,
  trading_period_end DATE,
  measurement_period_start DATE,
  measurement_period_end DATE,
  data_provider_name TEXT,
  data_provider_station_mode TEXT,
  description TEXT,
  inserted_at TIMESTAMP WITHOUT TIME ZONE,
  updated_at TIMESTAMP WITHOUT TIME ZONE
);

CREATE INDEX IF NOT EXISTS read_contracts_region_idx
  ON contracts (region);

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
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS orders (
  id BIGINT PRIMARY KEY,
  contract_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  token_type TEXT NOT NULL,
  order_side TEXT NOT NULL,
  price NUMERIC(10, 2) NOT NULL,
  quantity BIGINT NOT NULL,
  status TEXT NOT NULL,
  inserted_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS read_orders_contract_status_idx
  ON orders (contract_id, status, inserted_at);

CREATE TABLE IF NOT EXISTS market_snapshots (
  id BIGINT PRIMARY KEY,
  contract_id BIGINT NOT NULL,
  best_above NUMERIC(10, 2),
  best_below NUMERIC(10, 2),
  mid_above NUMERIC(10, 2),
  mid_below NUMERIC(10, 2),
  inserted_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS read_market_snapshots_contract_inserted_at_idx
  ON market_snapshots (contract_id, inserted_at);

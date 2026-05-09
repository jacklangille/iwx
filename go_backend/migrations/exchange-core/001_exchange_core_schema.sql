CREATE TABLE IF NOT EXISTS contracts (
  id BIGSERIAL PRIMARY KEY,
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
  inserted_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS contracts_region_idx
  ON contracts (region);

CREATE TABLE IF NOT EXISTS contract_commands (
  command_id TEXT PRIMARY KEY,
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
  command_status TEXT NOT NULL,
  error_message TEXT,
  result_contract_id BIGINT,
  enqueued_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  started_at TIMESTAMP WITHOUT TIME ZONE,
  completed_at TIMESTAMP WITHOUT TIME ZONE,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS contract_commands_status_updated_idx
  ON contract_commands (command_status, updated_at);

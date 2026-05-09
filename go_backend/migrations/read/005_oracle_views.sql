CREATE TABLE IF NOT EXISTS oracle_observations (
  id BIGINT PRIMARY KEY,
  contract_id BIGINT NOT NULL,
  provider_name TEXT NOT NULL,
  station_id TEXT NOT NULL,
  observed_metric TEXT NOT NULL,
  observation_window_start TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  observation_window_end TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  observed_value TEXT NOT NULL,
  normalized_value TEXT NOT NULL,
  observed_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  recorded_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS read_oracle_observations_contract_observed_idx
  ON oracle_observations (contract_id, observed_at DESC, recorded_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS contract_resolutions (
  id BIGINT PRIMARY KEY,
  contract_id BIGINT NOT NULL,
  provider_name TEXT NOT NULL,
  station_id TEXT NOT NULL,
  observed_metric TEXT NOT NULL,
  observation_window_start TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  observation_window_end TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  rule_version TEXT NOT NULL,
  resolved_value TEXT NOT NULL,
  outcome TEXT NOT NULL,
  resolved_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS read_contract_resolutions_contract_resolved_idx
  ON contract_resolutions (contract_id, resolved_at DESC, id DESC);

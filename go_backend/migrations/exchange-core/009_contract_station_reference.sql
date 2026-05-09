ALTER TABLE contracts
  ADD COLUMN IF NOT EXISTS station_id TEXT;

ALTER TABLE contract_commands
  ADD COLUMN IF NOT EXISTS station_id TEXT;

CREATE INDEX IF NOT EXISTS contracts_station_provider_idx
  ON contracts (data_provider_name, station_id);

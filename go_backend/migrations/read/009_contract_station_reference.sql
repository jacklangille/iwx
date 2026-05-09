ALTER TABLE contracts
  ADD COLUMN IF NOT EXISTS station_id TEXT;

CREATE INDEX IF NOT EXISTS read_contracts_station_provider_idx
  ON contracts (data_provider_name, station_id);

CREATE TABLE IF NOT EXISTS weather_stations (
  id BIGINT PRIMARY KEY,
  provider_name TEXT NOT NULL,
  station_id TEXT NOT NULL,
  display_name TEXT NOT NULL,
  region TEXT NOT NULL,
  latitude DOUBLE PRECISION,
  longitude DOUBLE PRECISION,
  supported_metrics TEXT NOT NULL DEFAULT '',
  active BOOLEAN NOT NULL DEFAULT TRUE,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  UNIQUE (provider_name, station_id)
);

CREATE INDEX IF NOT EXISTS weather_stations_active_region_idx
  ON weather_stations (active, region, provider_name, station_id);

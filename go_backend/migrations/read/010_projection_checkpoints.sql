ALTER TABLE contract_resolutions
  ADD COLUMN IF NOT EXISTS event_id TEXT;

UPDATE contract_resolutions
SET event_id = 'contract-resolved:' || id::text
WHERE event_id IS NULL;

ALTER TABLE contract_resolutions
  ALTER COLUMN event_id SET NOT NULL;

ALTER TABLE contract_resolutions
  ADD COLUMN IF NOT EXISTS published_at TIMESTAMP WITHOUT TIME ZONE;

CREATE UNIQUE INDEX IF NOT EXISTS read_contract_resolutions_event_id_idx
  ON contract_resolutions (event_id);

CREATE TABLE IF NOT EXISTS projection_checkpoints (
  projection_key TEXT PRIMARY KEY,
  event_id TEXT NOT NULL,
  version BIGINT NOT NULL,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

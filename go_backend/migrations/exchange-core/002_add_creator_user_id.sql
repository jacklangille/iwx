ALTER TABLE contracts
  ADD COLUMN IF NOT EXISTS creator_user_id BIGINT;

UPDATE contracts
SET creator_user_id = 0
WHERE creator_user_id IS NULL;

ALTER TABLE contracts
  ALTER COLUMN creator_user_id SET NOT NULL;

ALTER TABLE contract_commands
  ADD COLUMN IF NOT EXISTS creator_user_id BIGINT;

UPDATE contract_commands
SET creator_user_id = 0
WHERE creator_user_id IS NULL;

ALTER TABLE contract_commands
  ALTER COLUMN creator_user_id SET NOT NULL;

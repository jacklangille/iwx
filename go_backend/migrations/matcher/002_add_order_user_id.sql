ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS user_id BIGINT;

UPDATE orders
SET user_id = 0
WHERE user_id IS NULL;

ALTER TABLE orders
  ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE order_commands
  ADD COLUMN IF NOT EXISTS user_id BIGINT;

UPDATE order_commands
SET user_id = 0
WHERE user_id IS NULL;

ALTER TABLE order_commands
  ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS cash_reservation_id BIGINT,
  ADD COLUMN IF NOT EXISTS position_lock_id BIGINT,
  ADD COLUMN IF NOT EXISTS reservation_correlation_id TEXT NOT NULL DEFAULT '';

ALTER TABLE executions
  ADD COLUMN IF NOT EXISTS token_type TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS buyer_cash_reservation_id BIGINT,
  ADD COLUMN IF NOT EXISTS seller_position_lock_id BIGINT;

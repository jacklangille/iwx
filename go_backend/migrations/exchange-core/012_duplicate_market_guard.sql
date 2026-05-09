CREATE UNIQUE INDEX IF NOT EXISTS contracts_active_duplicate_market_idx
  ON contracts (
    LOWER(COALESCE(data_provider_name, '')),
    LOWER(COALESCE(station_id, '')),
    LOWER(metric),
    COALESCE(threshold, -9223372036854775808),
    COALESCE(trading_period_start, DATE '0001-01-01'),
    COALESCE(trading_period_end, DATE '0001-01-01'),
    COALESCE(measurement_period_start, DATE '0001-01-01'),
    COALESCE(measurement_period_end, DATE '0001-01-01')
  )
  WHERE status IN (
    'draft',
    'pending_approval',
    'pending_collateral',
    'active',
    'trading_closed',
    'awaiting_resolution',
    'resolved'
  );

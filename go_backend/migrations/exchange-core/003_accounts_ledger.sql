CREATE TABLE IF NOT EXISTS cash_accounts (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  currency TEXT NOT NULL,
  available_cents BIGINT NOT NULL DEFAULT 0,
  locked_cents BIGINT NOT NULL DEFAULT 0,
  total_cents BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  CONSTRAINT cash_accounts_user_currency_key UNIQUE (user_id, currency),
  CONSTRAINT cash_accounts_non_negative_balances CHECK (
    available_cents >= 0 AND locked_cents >= 0 AND total_cents >= 0
  ),
  CONSTRAINT cash_accounts_balance_invariant CHECK (
    available_cents + locked_cents = total_cents
  )
);

CREATE INDEX IF NOT EXISTS cash_accounts_user_idx
  ON cash_accounts (user_id, currency);

CREATE TABLE IF NOT EXISTS ledger_entries (
  id BIGSERIAL PRIMARY KEY,
  account_id BIGINT NOT NULL REFERENCES cash_accounts(id),
  user_id BIGINT NOT NULL,
  entry_type TEXT NOT NULL,
  amount_cents BIGINT NOT NULL,
  reference_type TEXT NOT NULL,
  reference_id TEXT NOT NULL,
  correlation_id TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  occurred_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ledger_entries_user_occurred_idx
  ON ledger_entries (user_id, occurred_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS ledger_entries_account_idx
  ON ledger_entries (account_id, id DESC);

CREATE TABLE IF NOT EXISTS collateral_locks (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  contract_id BIGINT NOT NULL,
  currency TEXT NOT NULL,
  amount_cents BIGINT NOT NULL,
  status TEXT NOT NULL,
  reference_id TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  reference_issuance_id BIGINT,
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  released_at TIMESTAMP WITHOUT TIME ZONE
);

CREATE INDEX IF NOT EXISTS collateral_locks_user_idx
  ON collateral_locks (user_id, currency, id DESC);

CREATE INDEX IF NOT EXISTS collateral_locks_contract_idx
  ON collateral_locks (contract_id, status, id DESC);

CREATE TABLE IF NOT EXISTS order_cash_reservations (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  contract_id BIGINT NOT NULL,
  currency TEXT NOT NULL,
  amount_cents BIGINT NOT NULL,
  status TEXT NOT NULL,
  reference_type TEXT NOT NULL,
  reference_id TEXT NOT NULL,
  correlation_id TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  released_at TIMESTAMP WITHOUT TIME ZONE
);

CREATE INDEX IF NOT EXISTS order_cash_reservations_user_idx
  ON order_cash_reservations (user_id, currency, id DESC);

CREATE INDEX IF NOT EXISTS order_cash_reservations_contract_status_idx
  ON order_cash_reservations (contract_id, status, id DESC);

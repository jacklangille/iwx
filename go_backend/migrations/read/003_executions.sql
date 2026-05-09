CREATE TABLE IF NOT EXISTS executions (
  id BIGINT PRIMARY KEY,
  execution_id TEXT NOT NULL UNIQUE,
  command_id TEXT NOT NULL,
  contract_id BIGINT NOT NULL,
  buy_order_id BIGINT NOT NULL,
  sell_order_id BIGINT NOT NULL,
  buyer_user_id BIGINT NOT NULL,
  seller_user_id BIGINT NOT NULL,
  price NUMERIC(10, 2) NOT NULL,
  quantity BIGINT NOT NULL,
  occurred_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS read_executions_contract_sequence_idx
  ON executions (contract_id, id DESC);

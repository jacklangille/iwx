# Architecture

## Status

This document defines the intended architecture for IWX and the current implementation boundary.

Current decision:

- the Go services are authoritative for backend evolution
- the frontend target is React

## Product Model

IWX is a centralized weather exchange with an internal ledger.

Users do not mint onchain assets. Instead:

1. a user proposes a weather contract
2. the platform validates and approves it
3. collateral is locked in exchange-controlled accounts
4. the platform issues paired internal claims:
   - `above`
   - `below`
5. users trade those claims on the exchange
6. an oracle resolves the outcome
7. settlement credits winners and releases collateral

## Service Boundaries

Target service set:

- `auth-service`
  - users
  - credentials
  - JWTs
  - sessions
- `exchange-core-service`
  - accounts
  - contracts
  - issuance
  - portfolio
  - settlement
- `oracle-service`
  - weather observations
  - normalized measurements
  - deterministic contract resolution
- `settlement-service`
  - consumes resolution events
  - applies payouts/refunds
  - closes contract state
- `matching-service`
  - order acceptance
  - order book
  - executions
  - market snapshots
- `read-api`
  - public/private query APIs
  - denormalized UI projections
  - UI-facing market/query APIs
- React frontend

## Critical Ownership Rule

The matcher should not own:

- cash balances
- collateral
- portfolio source-of-truth
- settlement
- oracle resolution
- contract moderation/approval lifecycle

The matcher should own:

- live tradable state for active markets
- open orders
- executions
- book sequencing
- matcher-local market snapshots

## Datastore Ownership

Target database layout:

- `auth-db`
  - users
  - credentials
- `exchange-core-db`
  - contracts
  - contract rules
  - accounts
  - ledger entries
  - collateral locks
  - issuance batches
  - positions
  - position locks
  - settlement records
- `oracle-db`
  - weather observations
  - contract resolutions
- `matcher-db`
  - orders
  - executions
  - market snapshots
  - matcher command state
- `read-db`
  - query-optimized projections for UI and APIs

## Contract Lifecycle

Target contract states:

- `draft`
- `pending_approval`
- `pending_collateral`
- `active`
- `trading_closed`
- `awaiting_resolution`
- `resolved`
- `settled`
- `cancelled`

## Trading Flow

Target centralized exchange flow:

1. user places order
2. gateway authenticates user
3. risk/reservation check runs
4. funds or inventory are reserved
5. matcher accepts the order
6. matcher executes trades
7. execution events update balances and positions
8. read models update market and user views

## Event Flow

The matcher should propagate successful matches through execution events.

Typical event consumers:

- accounts / ledger
- portfolio / positions
- read projections
- notifications / audit

This keeps the matcher as the source of truth for:

- what matched
- when it matched
- at what price and quantity

without making it the owner of balances.

## Current Transitional State

What is implemented today:

- Go auth service
- Go read API service
- Go exchange-core service
- Go matcher service
- Go migration runner
- async command submission over NATS JetStream
- matcher partitioning by `contract_id` for order commands
- matcher-owned:
  - `orders`
  - `executions`
  - `market_snapshots`
  - `order_commands`
- exchange-core-owned in code boundary:
  - `contracts`
  - `contract_rules`
  - `contract_commands`
  - `cash_accounts`
  - `ledger_entries`
  - `collateral_locks`
  - `issuance_batches`
  - `positions`
  - `position_locks`
  - `order_cash_reservations`
- authenticated command submission with JWT-backed user identity propagation:
  - `user_id` on order commands and orders
  - `creator_user_id` on contracts and contract commands
- pre-match reservation flow:
  - bid orders require cash reservations
  - ask orders require position locks
  - read-api reserves first, then enqueues to matcher
- matcher execution persistence:
  - every incoming order is written to matcher storage
  - every fill produces an execution row
  - every fill publishes an `execution_created` event
  - exchange-core consumes execution events to apply balance and portfolio deltas
  - read-api consumes execution events to refresh market projections and invalidate hot cache state
  - command redelivery is short-circuited by `command_id` replay
- oracle persistence:
  - raw/normalized observations are stored in `oracle-db`
  - final resolution payloads are stored in `oracle-db`
  - successful resolution publishes a `contract_resolved` event
- settlement persistence:
  - settlement consumes `contract_resolved`
  - payouts and refunds are written into exchange-core ledger + settlement tables
  - contract status transitions to `settled`
- read-api reads from dedicated read DB projections
- read-api user-state projections:
  - cash accounts
  - positions
  - position locks
  - collateral locks
  - cash reservations
- read-api delivery:
  - short-lived in-memory caching for hot market reads
  - browser-polled JSON endpoints for market, order command, and portfolio updates

What is transitional and expected to change:

- projection writes are currently service-driven syncs, not an event-driven projection pipeline yet
- the current UI uses direct browser polling against read endpoints rather than server-side push

## Immediate Architecture Direction

Near-term priorities:

1. keep Go as the only backend evolution path
2. harden service-to-service auth between Go services
3. add event-driven push only if the product actually needs lower-latency subscriptions
4. continue hardening execution-to-portfolio synchronization with integration and replay tooling
5. expand the React UI around contract lifecycle and portfolio workflows
6. harden service-level observability and dead-letter handling

## Related Docs

- roadmap: [TODO.md](/C:/Users/willj/src/iwx/TODO.md)
- domain model: [docs/domain-model.md](/C:/Users/willj/src/iwx/docs/domain-model.md)
- current Go backend notes: [go_backend/README.md](/C:/Users/willj/src/iwx/go_backend/README.md)

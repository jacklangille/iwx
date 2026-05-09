# Go Backend

This directory contains the current authoritative backend implementation for IWX.

The current frontend lives separately in [ui/](/C:/Users/willj/src/iwx/ui) as a Vite React application.

## Current Services

- `cmd/auth`
  - login and JWT issuance
- `cmd/read-api`
  - public/private read APIs and order submission
- `cmd/exchange-core`
  - contract metadata and lifecycle ownership
- `cmd/matcher`
  - asynchronous order matcher over NATS JetStream
- `cmd/oracle`
  - weather observations and deterministic contract resolution
- `cmd/settlement`
  - consumes resolution events and applies payouts / refunds

Legacy compatibility entrypoint:

- `cmd/server`
  - compatibility wrapper for the current read API process

## Current Behavior

Read API:

- `GET /api/contracts`
- `GET /api/contracts/:id/market_state`
- `GET /api/contracts/:id/market_snapshots`
- `GET /api/contracts/:id/executions`
- `GET /api/contracts/:id/observations`
- `GET /api/contracts/:id/resolution`
- `GET /api/contracts/:id/settlements`
- `GET /api/orders`
- `POST /api/orders` (authenticated)
- `GET /api/order_commands/:command_id` (authenticated owner only)
- `GET /api/me/account` (authenticated)
- `GET /api/me/positions` (authenticated)
- `GET /api/me/position_locks` (authenticated)
- `GET /api/me/collateral_locks` (authenticated)
- `GET /api/me/cash_reservations` (authenticated)
- `GET /api/me/portfolio` (authenticated)
- `GET /api/me/settlements` (authenticated)
- `GET /api/stream/contracts/:id/market`
- `GET /api/stream/me/order_commands/:id` (authenticated)
- `GET /api/stream/me/portfolio` (authenticated)

Exchange-core API:

- `GET /api/accounts/me` (authenticated)
- `GET /api/accounts/me/ledger` (authenticated)
- `GET /api/accounts/me/collateral_locks` (authenticated)
- `GET /api/accounts/me/cash_reservations` (authenticated)
- `POST /api/accounts/deposits` (authenticated)
- `POST /api/accounts/withdrawals` (authenticated)
- `POST /api/accounts/collateral_locks` (authenticated)
- `POST /api/accounts/collateral_locks/:id/release` (authenticated owner only)
- `POST /api/accounts/cash_reservations` (authenticated)
- `POST /api/accounts/cash_reservations/:id/release` (authenticated owner only)
- `GET /api/positions/me` (authenticated)
- `GET /api/positions/me/locks` (authenticated)
- `POST /api/contracts` (authenticated)
- `GET /api/contracts/:id` (authenticated owner only)
- `POST /api/contracts/:id/submit_for_approval` (authenticated owner only)
- `POST /api/contracts/:id/approve` (authenticated owner only)
- `GET /api/contracts/:id/collateral_requirement?paired_quantity=N` (authenticated owner only)
- `POST /api/contracts/:id/collateral_locks` (authenticated owner only)
- `GET /api/contracts/:id/issuance_batches` (authenticated owner only)
- `POST /api/contracts/:id/issuance_batches` (authenticated owner only)
- `POST /api/contracts/:id/activate` (authenticated owner only)
- `POST /api/contracts/:id/cancel` (authenticated owner only)
- `GET /api/contract_commands/:command_id` (authenticated owner only)

Auth API:

- `POST /api/auth/login`

Oracle API:

- `POST /api/oracle/observations`
- `GET /api/oracle/contracts/:id/observations`
- `POST /api/oracle/contracts/:id/resolve`
- `GET /api/oracle/contracts/:id/resolution`

## Current Phase State

Implemented so far:

- Phase 2 service boundaries
- Phase 3 database boundaries and migrations
- Phase 4 auth and identity propagation
- Phase 5 exchange-core accounts and ledger primitives
- Phase 6 exchange-core contract lifecycle
- Phase 7 issuance batches and creator position credits
- Phase 8 reservation-aware order intake
- Phase 9 matcher executions and idempotent command replay
- Phase 10 read-db user projections, hot market caching, and SSE delivery
- Phase 11 oracle observations, deterministic resolution, and contract-resolved events
- Phase 12 settlement consumption, payouts/refunds, and settlement history projections

Current auth behavior:

- `cmd/auth` issues JWTs with both username and numeric `user_id`
- read-api and exchange-core require bearer auth for write and command-owner endpoints
- order writes persist `user_id`
- contract writes persist `creator_user_id`
- command status endpoints are owner-scoped

Current accounting behavior:

- exchange-core owns cash accounts, ledger entries, collateral locks, and order cash reservations
- cash account balances are stored as `available_cents`, `locked_cents`, and `total_cents`
- exchange-core DB enforces `available_cents + locked_cents = total_cents`
- ledger entries are append-only accounting records for deposits, withdrawals, locks, and releases

Current contract lifecycle behavior:

- contract creation now always creates a `draft`
- exchange-core owns `contract_rules` separately from `contracts`
- lifecycle transitions are explicit:
  - `draft`
  - `pending_approval`
  - `pending_collateral`
  - `active`
  - `cancelled`
- collateral requirements are computed before activation
- cancelling before activation releases locked collateral

Current issuance behavior:

- issuance batches are tied to consumed collateral locks
- each issuance credits paired `above` and `below` positions to the creator
- creator positions are stored in `exchange-core` as internal inventory
- contract activation now requires issued supply, not just a raw collateral lock

Current reservation behavior:

- bid orders reserve cash in `exchange-core` before matcher enqueue
- ask orders reserve inventory through `position_locks` before matcher enqueue
- `read-api` generates the matcher `command_id` first and uses it as the reservation correlation key
- enqueue failure releases the reservation immediately
- matcher validation now requires reservation IDs on incoming orders

Current matcher execution behavior:

- every incoming order is persisted in matcher storage, even if it fills immediately
- each fill creates an `execution` row owned by the matcher
- matcher publishes `execution_created` events after successful fills
- exchange-core consumes execution events to debit reserved buyer cash, credit seller cash, debit seller locked inventory, and credit buyer positions
- read-api consumes execution events to refresh market projections from matcher storage into `read-db`
- duplicate JetStream deliveries are short-circuited by `command_id` replay instead of rematching the order

Current read-model behavior:

- `read-db` now stores user cash accounts, positions, position locks, collateral locks, and cash reservations
- exchange-core mutations project user state into `read-db`
- read-api serves authenticated portfolio/dashboard reads from `read-db` instead of reaching back into exchange-core
- hot market reads use a short-lived in-memory cache inside `read-api`
- read-api exposes polling-based SSE endpoints for market, command, and portfolio updates

Current oracle behavior:

- oracle observations are stored in an oracle-owned database
- contracts are resolved deterministically from the latest observation inside the contract measurement window
- inclusive threshold handling comes from `contract_rules.resolution_inclusive_side`
- successful resolution writes a `contract_resolution`, updates the contract status to `resolved`, projects oracle state into `read-db`, and publishes a `contract_resolved` event to NATS JetStream

Current settlement behavior:

- `cmd/settlement` consumes `contract_resolved` events from JetStream
- settlement unwinds open cash reservations and position locks for the contract
- winning holders are credited through cash account ledger entries
- collateral is consumed from locked creator balances for payouts
- unused or cancelled collateral is released back to available balances
- contract positions are zeroed and contract status moves to `settled`
- settlement entries are projected into `read-db` for contract and user history views

## Local Infrastructure

Local services are defined in [docker-compose.yml](/C:/Users/willj/src/iwx/docker-compose.yml):

- `read-postgres`
- `exchange-core-postgres`
- `oracle-postgres`
- `matcher-postgres-0`
- `matcher-postgres-1`
- `auth-postgres`
- `nats`
- `auth`
- `exchange-core`
- `read-api`
- `matcher`
- `oracle`
- `settlement`

Environment files:

- sample env: [go_backend/.env.example](/C:/Users/willj/src/iwx/go_backend/.env.example)
- local env: [go_backend/.env](/C:/Users/willj/src/iwx/go_backend/.env)

The Go services automatically load `.env` in local development.
Compose also injects strict per-service database URLs so a service cannot silently fall back into another service's database.

## Migrations

Schemas are now owned by Go migrations, not Docker init scripts.
Each service runs its required migrations automatically on boot.

Run all migrations:

```powershell
cd go_backend
go run .\cmd\migrate -target all
```

Individual targets:

- `auth`
- `exchange-core`
- `matcher`
- `oracle`
- `settlement` is event-driven and reuses `exchange-core` + `read` data stores
- `read`

## Matcher Partitioning

- order commands are routed by `contract_id % IWX_MATCHER_PARTITION_COUNT`
- each matcher instance consumes only the partitions in `IWX_MATCHER_OWNED_PARTITIONS`
- overlapping partition ownership is unsafe and must be avoided

Useful env vars:

- `IWX_MATCHER_PARTITION_COUNT`
- `IWX_MATCHER_OWNED_PARTITIONS`
- `IWX_MATCHER_INSTANCE_ID`
- `IWX_NATS_STREAM_PLACE_ORDER`
- `IWX_NATS_SUBJECT_PLACE_ORDER`

## Current Storage Topology

- auth DB
  - auth users and credentials
- exchange-core DB
  - contract metadata
  - contract rules
  - contract command state
  - cash accounts
  - ledger entries
  - collateral locks
  - issuance batches
  - positions
  - position locks
  - order cash reservations
- read DB
  - query projections for contracts, orders, executions, snapshots, order command status, cash accounts, and portfolio state
- matcher primary DB
  - matcher-owned write path for orders, executions, snapshots, and command state
- matcher replica DB
  - debugging and verification path only

## Configuration Boundary

- `IWX_AUTH_DATABASE_URL` is required for `auth`
- `IWX_EXCHANGE_CORE_DATABASE_URL` is required for `exchange-core`
- `IWX_READ_DATABASE_URL` is required for `read-api`
- `IWX_MATCHER_DATABASE_URL` is required for `matcher`
- `IWX_ORACLE_DATABASE_URL` is required for `oracle`
- services now fail fast on missing DB ownership config instead of falling back to `DATABASE_URL` or another service DB

## Package Layout

```text
go_backend/
  cmd/auth
  cmd/migrate
  cmd/read-api
  cmd/exchange-core
  cmd/matcher
  cmd/server
  internal/app
  internal/auth
  internal/authhttp
  internal/commands
  internal/config
  internal/domain
  internal/events
  internal/exchangecore
  internal/exchangecorehttp
  internal/httpapi
  internal/market
  internal/matching
  internal/messaging
  internal/messaging/natsbus
  internal/money
  internal/readmodel
  internal/realtime
  internal/store
  internal/store/postgres
  pkg/logging
```

## Related Docs

- target architecture: [docs/architecture.md](/C:/Users/willj/src/iwx/docs/architecture.md)
- domain model: [docs/domain-model.md](/C:/Users/willj/src/iwx/docs/domain-model.md)
- implementation roadmap: [TODO.md](/C:/Users/willj/src/iwx/TODO.md)

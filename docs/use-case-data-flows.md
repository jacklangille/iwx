# Use Case Data Flows

This document describes the current runtime data flows by reading the Go code paths directly. It does not treat older architecture notes as authoritative.

Primary code references:

- auth runtime: [go_backend/internal/auth/service.go](C:/Users/willj/src/iwx/go_backend/internal/auth/service.go), [go_backend/internal/authhttp/server.go](C:/Users/willj/src/iwx/go_backend/internal/authhttp/server.go)
- exchange-core runtime: [go_backend/internal/app/exchange_core.go](C:/Users/willj/src/iwx/go_backend/internal/app/exchange_core.go), [go_backend/internal/exchangecore](C:/Users/willj/src/iwx/go_backend/internal/exchangecore), [go_backend/internal/exchangecorehttp](C:/Users/willj/src/iwx/go_backend/internal/exchangecorehttp)
- matcher runtime: [go_backend/internal/app/matcher.go](C:/Users/willj/src/iwx/go_backend/internal/app/matcher.go), [go_backend/internal/matching](C:/Users/willj/src/iwx/go_backend/internal/matching), [go_backend/internal/store/postgres/matching_repository.go](C:/Users/willj/src/iwx/go_backend/internal/store/postgres/matching_repository.go)
- oracle runtime: [go_backend/internal/app/oracle.go](C:/Users/willj/src/iwx/go_backend/internal/app/oracle.go), [go_backend/internal/oracle/service.go](C:/Users/willj/src/iwx/go_backend/internal/oracle/service.go), [go_backend/internal/oraclehttp](C:/Users/willj/src/iwx/go_backend/internal/oraclehttp)
- settlement runtime: [go_backend/internal/app/settlement.go](C:/Users/willj/src/iwx/go_backend/internal/app/settlement.go), [go_backend/internal/settlement/service.go](C:/Users/willj/src/iwx/go_backend/internal/settlement/service.go)
- projector runtime: [go_backend/internal/app/projector.go](C:/Users/willj/src/iwx/go_backend/internal/app/projector.go), [go_backend/internal/projector/service.go](C:/Users/willj/src/iwx/go_backend/internal/projector/service.go), [go_backend/internal/readprojection/applier.go](C:/Users/willj/src/iwx/go_backend/internal/readprojection/applier.go)
- read API runtime: [go_backend/internal/httpapi/server.go](C:/Users/willj/src/iwx/go_backend/internal/httpapi/server.go), [go_backend/internal/readmodel/service.go](C:/Users/willj/src/iwx/go_backend/internal/readmodel/service.go)

## Runtime Boundaries

The current services and owned state are:

- `auth`
  - owns user lookup, password hash verification, signup, and JWT issuance
  - writes `auth-db`
- `exchange-core`
  - owns contracts, contract rules, account balances, ledger, collateral locks, issuance, positions, position locks, cash reservations, execution application, and settlement application
  - writes `exchange-core-db`
  - consumes `execution_created`
  - emits `projection_change`
- `matcher`
  - owns order command state, orders, executions, and market snapshots
  - writes `matcher-db`
  - consumes order commands from NATS JetStream
  - emits `execution_created`
  - emits `projection_change`
- `oracle`
  - owns station catalog, observations, resolutions, and its outbox
  - writes `oracle-db`
  - emits `projection_change`
  - emits `contract_resolved` through its outbox dispatcher
- `settlement`
  - owns no primary database
  - consumes `contract_resolved`
  - calls `exchange-core` over internal HTTP to apply settlement
  - emits `settlement_completed`
- `projector`
  - is the only writer to `read-db`
  - consumes `projection_change`
  - fetches authoritative bundles from owner services over internal HTTP
- `read-api`
  - only serves reads from `read-db`

Current bus-level events are defined in [go_backend/internal/events/types.go](C:/Users/willj/src/iwx/go_backend/internal/events/types.go):

- `execution_created`
- `contract_resolved`
- `settlement_completed`
- `projection_change`

## Cross-Service Patterns

Current write pattern:

1. client calls the owning service over HTTP
2. owner writes its own DB
3. owner may enqueue an outbox event or publish to NATS
4. downstream consumer applies owned side effects
5. owner emits `projection_change`
6. projector fetches fresh bundles and writes `read-db`
7. browser polls `read-api`

Current read pattern:

1. browser polls `read-api`
2. `read-api` serves only `read-db`
3. no read path reaches into owner databases directly

## Failure Semantics Legend

When thinking about partial completion in the current code, the important distinction is not just "can this fail," but "what state is left behind if it fails halfway through."

The current code mostly falls into four buckets:

- atomic owner transaction failure
  - the owner service transaction fails and rolls back
  - no durable state change should remain
  - retry is clean
- committed owner state plus unpublished side effect
  - owner DB commit succeeds
  - event enqueue or publish fails afterward
  - recovery depends on outbox durability or a replay path that can re-emit the missing event
- consumer-side partial apply with retry
  - a consumer starts handling an event
  - owned side effects fail before the consumer acks the message
  - JetStream redelivery retries the same event
  - correctness depends on idempotency keys such as `execution_id`, `command_id`, or settlement `event_id`
- read-model partial apply
  - projector updates some `read-db` tables and then fails before checkpointing
  - browser may temporarily observe an internally inconsistent read model
  - because the checkpoint is not advanced, the same projection event retries and should converge on the next successful apply

The strongest guarantees currently present in code are:

- matcher order command replay by `command_id`
- execution application idempotency by `execution_id`
- settlement application idempotency by resolution `event_id`
- projector checkpoint/version gating for duplicate and stale projection events

The weakest area is still `read-db` projection application, because one projection event can map to multiple sequential table replacements before checkpoint commit.

## Use Case 1: Sign Up

Entrypoint:

- `POST /api/auth/signup` in [go_backend/internal/authhttp/server.go](C:/Users/willj/src/iwx/go_backend/internal/authhttp/server.go)

Flow:

1. `authhttp.handleAuthSignup` parses `username` and `password`.
2. `auth.Service.Signup` validates:
   - username length >= 3
   - password length >= 8
3. `HashPassword` uses bcrypt.
4. `AuthUserRepository.CreateUser` inserts into `auth-db.users`.
5. `auth` signs a JWT containing:
   - `sub`
   - `uid`
   - `iss`
   - `iat`
   - `exp`
6. response returns `201 Created` with token and `user_id`.

State changes:

- `auth-db.users` insert

Events:

- none

Potential failure points:

- invalid JSON body
- username/password missing
- username too short or password too short
- bcrypt hashing failure
- duplicate username unique violation
- `auth` misconfiguration with missing JWT secret
- DB write failure in `auth-db`

## Use Case 2: Log In

Entrypoint:

- `POST /api/auth/login` in [go_backend/internal/authhttp/server.go](C:/Users/willj/src/iwx/go_backend/internal/authhttp/server.go)

Flow:

1. `authhttp.handleAuthLogin` parses username/password.
2. `auth.Service.Authenticate` loads the user from `auth-db`.
3. bcrypt compares the provided password with `password_hash`.
4. inactive users are rejected.
5. a JWT is signed and returned.

State changes:

- none

Events:

- none

Potential failure points:

- invalid JSON body
- missing credentials
- user not found
- bcrypt mismatch
- inactive user
- DB read failure in `auth-db`
- JWT signing failure

## Use Case 3: Browse Stations And Markets

Entrypoints:

- `GET /api/stations`
- `GET /api/contracts`
- `GET /api/contracts/:id/market_state`
- `GET /api/contracts/:id/executions`
- `GET /api/contracts/:id/market_snapshots`

Read-side code:

- route registration: [go_backend/internal/httpapi/server.go](C:/Users/willj/src/iwx/go_backend/internal/httpapi/server.go)
- query service: [go_backend/internal/readmodel/service.go](C:/Users/willj/src/iwx/go_backend/internal/readmodel/service.go)

Flow:

1. browser polls `read-api`.
2. `read-api` reads station, contract, order, execution, snapshot, and resolution projections from `read-db`.
3. market reads are optionally served through a short-lived in-memory cache in `readmodel.Service`.

State changes:

- none

Events:

- none on the request path

Potential failure points:

- `read-db` projection lag means stale UI data
- `read-db` query failure
- market cache may temporarily serve stale data until invalidated by the next projected market change
- contract-specific routes may 404 if projector has not applied the contract projection yet

## Use Case 4: Deposit Cash

Entrypoint:

- `POST /api/accounts/deposits` in [go_backend/internal/exchangecorehttp/accounts_handler.go](C:/Users/willj/src/iwx/go_backend/internal/exchangecorehttp/accounts_handler.go)

Core code:

- [go_backend/internal/exchangecore/accounts_service.go](C:/Users/willj/src/iwx/go_backend/internal/exchangecore/accounts_service.go)

Flow:

1. authenticated request hits `exchange-core`.
2. `accounts_handler` builds `store.DepositCashInput`.
3. `Service.DepositCash` validates:
   - `user_id`
   - `amount_cents > 0`
   - currency present
4. repository writes to `exchange-core-db`:
   - cash account updated
   - ledger entry inserted
5. `projectUser` emits `projection_change.user_state_changed`.
6. `exchange-core` outbox dispatcher publishes that event.
7. `projector` consumes `projection_change`.
8. projector fetches `GetUserStateBundle` from `exchange-core` internal HTTP.
9. projector replaces user cash account, positions, locks, reservations, and settlements in `read-db`.

State changes:

- `exchange-core-db.cash_accounts`
- `exchange-core-db.ledger_entries`
- `exchange-core-db.outbox_events`
- `read-db` user projections after projector apply

Potential failure points:

- missing or invalid auth
- invalid amount or currency
- `exchange-core-db` transaction failure
- projection change outbox enqueue/publish failure
- projector NATS consumer lag
- projector fetch failure against `exchange-core`
- read projection apply failure

Partial-completion analysis:

- if the cash account update or ledger insert fails inside `exchange-core-db`, the transaction should roll back cleanly and no deposit remains
- if the DB commit succeeds but the `projection_change` enqueue fails afterward, the user balance is durably updated in `exchange-core-db` but `read-db` stays stale until another later user-state event causes the projector to refresh
- if the outbox row exists but publish is delayed, the state is safe but stale; eventual recovery depends on the outbox dispatcher
- if the projector fetches the user bundle successfully but fails while replacing one of the user projection tables, `read-db` can temporarily show mismatched balance, position, or lock data until the same projection version is retried and checkpointed

## Use Case 5: Create Contract Draft

Entrypoint:

- `POST /api/contracts` in [go_backend/internal/exchangecorehttp/server.go](C:/Users/willj/src/iwx/go_backend/internal/exchangecorehttp/server.go)

Core code:

- contract validation and station binding: [go_backend/internal/exchangecore/service.go](C:/Users/willj/src/iwx/go_backend/internal/exchangecore/service.go)
- DB write path: [go_backend/internal/store/postgres/contract_write_repository.go](C:/Users/willj/src/iwx/go_backend/internal/store/postgres/contract_write_repository.go)

Flow:

1. authenticated request hits `exchange-core`.
2. `SubmitCreateContract` forces initial status to `draft`.
3. `ValidateCreateContract` validates required fields and date ordering.
4. `validateAndPopulateStation` calls `oracle` over internal HTTP through `oraclestationhttp.Client.FindStation`.
5. station validation enforces:
   - provider/station exists
   - station is active
   - station supports the requested metric
   - request region matches station region if provided
6. `rejectDuplicateMarket` checks `exchange-core-db` for an overlapping active or pending market with the same:
   - provider
   - station
   - metric
   - threshold
   - trading window
   - measurement window
7. repository transaction:
   - marks `contract_commands` processing
   - inserts `contracts`
   - inserts `contract_rules`
   - marks `contract_commands` succeeded
8. on success, `exchange-core` emits `projection_change.contract_changed`.
9. projector fetches `GetContractBundle` over internal HTTP and upserts the contract projection into `read-db`.

State changes:

- `exchange-core-db.contract_commands`
- `exchange-core-db.contracts`
- `exchange-core-db.contract_rules`
- `exchange-core-db.outbox_events`
- `read-db.contracts` after projection

Potential failure points:

- invalid auth
- invalid contract payload or date ordering
- station lookup HTTP failure to `oracle`
- unknown or inactive station
- unsupported metric for the chosen station
- duplicate market detected by service-level lookup
- duplicate market race caught by DB partial unique index
- `exchange-core-db` transaction failure
- contract projection outbox enqueue/publish failure
- projector fetch/apply failure

Partial-completion analysis:

- contract creation itself is transactional inside `exchange-core-db`; a failure before command success commit should leave no partially created contract/rule pair
- the duplicate-market DB unique index is the final guard if two requests race past the service-level duplicate check
- if the contract row commits but the projection change is not enqueued or not published, the draft exists durably in `exchange-core-db` but will be invisible in `read-db` until a later contract event reprojects it
- projector failure after fetching the contract bundle leaves the write side correct and the read side stale; retry is safe because contract projection is an upsert gated by version

## Use Case 6: Progress Contract To Active

Entrypoints:

- `POST /api/contracts/:id/submit_for_approval`
- `POST /api/contracts/:id/approve`
- `POST /api/contracts/:id/collateral_locks`
- `POST /api/contracts/:id/issuance_batches`
- `POST /api/contracts/:id/activate`
- `POST /api/contracts/:id/cancel`

Core code:

- lifecycle: [go_backend/internal/exchangecore/contracts_lifecycle.go](C:/Users/willj/src/iwx/go_backend/internal/exchangecore/contracts_lifecycle.go)
- issuance: [go_backend/internal/exchangecore/issuance_service.go](C:/Users/willj/src/iwx/go_backend/internal/exchangecore/issuance_service.go)

Flow:

1. creator-only lifecycle endpoints call `requireOwnedContract`.
2. `submit_for_approval` moves:
   - `draft -> pending_approval`
3. `approve` moves:
   - `pending_approval -> pending_collateral`
4. `collateral_locks` computes required collateral:
   - `max(multiplier * paired_quantity, $1000 floor)`
5. collateral lock transaction updates:
   - cash account locked balance
   - collateral lock row
   - ledger entry
6. `issuance_batches` consumes the collateral lock and credits paired positions:
   - `above`
   - `below`
7. `activate` requires:
   - contract status `pending_collateral`
   - at least one issued issuance batch
8. `cancel` is only allowed before issuance and releases locked collateral.
9. each balance/position/contract mutation emits either:
   - `projection_change.user_state_changed`
   - `projection_change.contract_changed`
10. projector refreshes the relevant contract and user bundles into `read-db`.

State changes:

- `exchange-core-db.contracts`
- `exchange-core-db.cash_accounts`
- `exchange-core-db.ledger_entries`
- `exchange-core-db.collateral_locks`
- `exchange-core-db.issuance_batches`
- `exchange-core-db.positions`
- `exchange-core-db.outbox_events`
- `read-db` contract and user projections

Potential failure points:

- non-owner trying to mutate the contract
- contract not found
- invalid lifecycle transition
- insufficient available cash for collateral lock
- invalid `paired_quantity`
- missing or invalid collateral lock id for issuance
- repository transaction failure for issuance or collateral changes
- projection event publish lag or failure
- projector bundle fetch or replace failure

Partial-completion analysis:

- collateral lock, issuance, and activation state transitions are owner-side operations; if their repository transaction fails, the write should roll back without a half-issued contract
- if a lifecycle step commits but its projection change is not durably enqueued, the contract can move forward in `exchange-core-db` while the UI still shows the old state
- cancellation is especially sensitive because it touches contract status, balances, and collateral lock state together; correctness depends on those being updated in one repository transaction
- projector retries are safe, but a failed user-state projection can briefly show collateral or positions from before issuance/activation even though the owner DB has already moved on

## Use Case 7: Place Buy Order

Entrypoint:

- `POST /api/orders` on `exchange-core` in [go_backend/internal/exchangecorehttp/orders_handler.go](C:/Users/willj/src/iwx/go_backend/internal/exchangecorehttp/orders_handler.go)

Flow:

1. authenticated request reaches `exchange-core`.
2. `ValidateOrderReservation` checks contract id, token type, side, price, quantity.
3. `command_id` is generated before reservation.
4. `ReserveOrderForMatching` creates an order cash reservation:
   - amount = `price_cents * quantity`
   - reference type `order_command`
   - correlation id = command id
5. `exchange-core` emits `projection_change.user_state_changed` for the user reservation state.
6. order command is enriched with:
   - `cash_reservation_id`
   - `reservation_correlation_id`
7. `matching.ValidatePlaceOrder` confirms reservation ids are present.
8. `MatcherClient.SubmitPlaceOrder` publishes the order command to JetStream using `MsgId(command_id)`.
9. if publish fails, `exchange-core` releases the reservation immediately.
10. matcher consumer receives the command and executes matching logic.
11. matcher always emits `projection_change.order_command_changed`.
12. matcher emits `projection_change.market_changed` if market state changed.
13. matcher emits `execution_created` for each fill.
14. `exchange-core` execution consumer applies fills to balances and positions and emits more `projection_change.user_state_changed`.
15. projector refreshes both market and user projections into `read-db`.
16. browser polls:
   - `GET /api/order_commands/:id`
   - `GET /api/contracts/:id/market_state`
   - `GET /api/me/portfolio`

State changes:

- `exchange-core-db.cash_accounts`
- `exchange-core-db.order_cash_reservations`
- `exchange-core-db.ledger_entries`
- `exchange-core-db.outbox_events`
- `matcher-db.order_commands`
- `matcher-db.orders`
- `matcher-db.executions`
- `matcher-db.market_snapshots`
- `matcher-db.outbox_events`
- `read-db` market and user projections

Potential failure points:

- invalid auth
- invalid order payload
- insufficient available cash for reservation
- reservation creation succeeds but matcher publish fails, requiring compensation release
- NATS unavailable during order publish
- matcher consumer lag
- matcher DB failure when persisting the order command or matching
- downstream execution application failure in `exchange-core`
- projector lag causing the command or market state to appear stale

Partial-completion analysis:

- reservation creation happens before matcher publish, so a buy order can fail after money is locked but before the command is published; the code attempts immediate compensation release on publish failure, but if that release fails there can be a temporarily stranded reservation in `exchange-core-db`
- matcher order processing is transactional inside `matcher-db`; if matching fails before commit, no partial order or execution state should remain
- matcher event emission is safer than before because it now uses an outbox, but the enqueue step is still after matching logic returns; if matcher state commits and event enqueue fails, recovery depends on command replay or a later market/order-command projection event, not a single fully transactional commit across business rows and outbox rows
- if `execution_created` publish is delayed, the order book in `matcher-db` may be ahead of balances and positions in `exchange-core-db`
- browser-visible state can split three ways during recovery:
  - matcher command state may already be `succeeded`
  - exchange-core balances may not yet reflect fills
  - `read-db` may lag both

## Use Case 8: Place Sell Order

Entrypoint:

- same `POST /api/orders` path in `exchange-core`

Flow differences from buy:

1. `ReserveOrderForMatching` creates a `position_lock` instead of a cash reservation.
2. lock is keyed by:
   - user
   - contract
   - side (`above` or `below`)
   - quantity
3. the order command is enriched with `position_lock_id`.
4. matcher requires `position_lock_id` for ask orders.
5. on fill, `exchange-core.ApplyExecution` consumes the seller’s locked inventory and credits cash.

State changes:

- `exchange-core-db.positions`
- `exchange-core-db.position_locks`
- `exchange-core-db.outbox_events`
- matcher state and `read-db` projection updates as in the buy flow

Potential failure points:

- seller lacks available inventory for the requested side/quantity
- position lock creation fails
- matcher publish fails after lock creation, requiring lock release
- duplicate or replayed executions must not re-consume locked inventory
- projector lag can make the user portfolio temporarily stale

Partial-completion analysis:

- the partial-failure shape is the same as the buy path, except the stranded pre-publish resource is inventory lock state instead of cash reservation state
- if lock creation commits and matcher publish fails, the system depends on the compensation release path to avoid leaving sell inventory unavailable
- if execution application is delayed, the seller can temporarily see inventory still locked or cash not yet credited even though matcher has already recorded the trade

## Use Case 9: Match Execution Application

This is a system-to-system flow triggered by matcher fills.

Source code:

- execution event publisher wrapper: [go_backend/internal/matching/execution_event_publishing_handler.go](C:/Users/willj/src/iwx/go_backend/internal/matching/execution_event_publishing_handler.go)
- exchange-core consumer: [go_backend/internal/messaging/natsbus/execution_consumer.go](C:/Users/willj/src/iwx/go_backend/internal/messaging/natsbus/execution_consumer.go)
- application service: [go_backend/internal/exchangecore/execution_service.go](C:/Users/willj/src/iwx/go_backend/internal/exchangecore/execution_service.go)

Flow:

1. matcher creates `domain.Execution` rows in `matcher-db`.
2. the wrapper emits `events.ExecutionCreated` carrying:
   - `execution_id`
   - `buy_order_id`
   - `sell_order_id`
   - buyer/seller user ids
   - reservation ids
3. matcher outbox stores the event.
4. matcher outbox dispatcher publishes `execution_created`.
5. `exchange-core` pulls the event from JetStream.
6. `ApplyExecution` writes execution side effects through `exchange-core-db`.
7. application is idempotent through execution application records in the repository.
8. affected users emit `projection_change.user_state_changed`.
9. projector refreshes user bundles into `read-db`.

State changes:

- `matcher-db.outbox_events`
- `exchange-core-db.execution_applications`
- `exchange-core-db.cash_accounts`
- `exchange-core-db.positions`
- `exchange-core-db.ledger_entries`
- `exchange-core-db.order_cash_reservations`
- `exchange-core-db.position_locks`
- `exchange-core-db.outbox_events`

Potential failure points:

- matcher outbox dispatch lag
- JetStream duplicate delivery
- exchange-core consumer failure leading to retry
- partial business apply is guarded by execution application idempotency, but DB failure before record commit still aborts the whole transaction
- user read model remains stale until projector catches up

Partial-completion analysis:

- this is one of the better-protected paths in the system
- `execution_created` is published from matcher outbox with stable `execution_id`, so duplicate bus delivery is expected and safe
- `exchange-core` applies the fill behind `execution_applications`; if the handler fails before that transaction commits, the message is retried and no application record should exist
- if the transaction commits but emitting `projection_change.user_state_changed` fails afterward, the execution side effects are durable while `read-db` remains stale; retry re-enters `ApplyExecution`, which should no-op on the execution application record and then re-emit projections
- that means the write side is protected against double fill, but browser-visible state can lag until the re-emission path succeeds

## Use Case 10: Record Oracle Observation

Entrypoint:

- `POST /api/oracle/observations` in [go_backend/internal/oraclehttp/handlers.go](C:/Users/willj/src/iwx/go_backend/internal/oraclehttp/handlers.go)

Flow:

1. request hits `oracle`.
2. `validateObservationInput` checks contract id, provider, station id, metric, windows, values, and `observed_at`.
3. `oracle` calls `exchange-core` internal client to verify the contract exists.
4. `oracleRepo.RecordObservation` inserts the observation into `oracle-db`.
5. `projectOracleState` emits `projection_change.oracle_state_changed`.
6. oracle outbox dispatcher publishes the projection change.
7. projector fetches oracle state bundle from `oracle` internal HTTP and updates `read-db`.

State changes:

- `oracle-db.oracle_observations`
- `oracle-db.outbox_events`
- `read-db` observations projection

Potential failure points:

- invalid observation payload
- contract lookup failure against `exchange-core`
- `oracle-db` insert failure
- oracle projection event publish lag
- projector fetch/apply failure

Partial-completion analysis:

- observation insertion is the owner-side durable step; if it fails, there is no partial observation
- if the observation commits but the oracle-state projection change is not enqueued or published, `oracle-db` contains the new observation while `read-db` does not
- later oracle events for the same contract should heal this because the projector fetches the full oracle bundle rather than incremental deltas

## Use Case 11: Resolve Contract

Entrypoint:

- `POST /api/oracle/contracts/:id/resolve` in [go_backend/internal/oraclehttp/handlers.go](C:/Users/willj/src/iwx/go_backend/internal/oraclehttp/handlers.go)

Flow:

1. `oracle` loads the contract from `exchange-core` internal HTTP.
2. if a resolution already exists, it reuses it.
3. otherwise `oracle` loads the contract rule from `exchange-core`.
4. it loads observations from `oracle-db`.
5. `latestObservationWithinWindow` selects the latest observation inside the measurement window.
6. normalized value is parsed and compared to threshold through `resolveOutcome`.
7. `oracleRepo.InsertResolution` writes the resolution with stable `event_id`.
8. `oracle` marks the contract status as `resolved` by calling `exchange-core.UpdateContractStatus` through the internal contract repository client.
9. `oracle` emits:
   - `projection_change.oracle_state_changed`
   - `projection_change.contract_changed`
10. if `resolution.PublishedAt` is nil, it enqueues a `contract_resolved` outbox event instead of publishing inline.
11. oracle outbox dispatcher publishes `contract_resolved` and marks the outbox event published.

State changes:

- `oracle-db.contract_resolutions`
- `oracle-db.outbox_events`
- `exchange-core-db.contracts.status = resolved`
- `read-db` oracle and contract projections after projector apply

Potential failure points:

- contract not found in `exchange-core`
- contract rule missing or threshold missing
- no observation in the contract measurement window
- invalid normalized value parsing
- `oracle-db` resolution insert failure
- internal `exchange-core` status update failure
- outbox enqueue failure
- outbox dispatch failure leaving the event unpublished until retry

Partial-completion analysis:

- this path now has a real oracle outbox, which is the main reason it is recoverable
- if resolution insert fails, there is no resolution and no downstream work
- if resolution insert succeeds but `exchange-core` contract status update fails, `oracle-db` can temporarily say the contract is resolved while `exchange-core-db` does not; a retry should reuse the same resolution row and retry the contract status update
- if status update succeeds but projection-change emission fails, the owner state is correct and only `read-db` is stale
- if outbox enqueue fails after resolution insert, the resolution exists but settlement will not start yet; retry should see the existing unpublished resolution and enqueue the missing outbox event
- if outbox publish fails, the outbox row remains pending and the dispatcher should retry without creating a second logical resolution

## Use Case 12: Settle Contract

This is driven asynchronously by `contract_resolved`.

Source code:

- consumer: [go_backend/internal/messaging/natsbus/settlement_consumer.go](C:/Users/willj/src/iwx/go_backend/internal/messaging/natsbus/settlement_consumer.go)
- service: [go_backend/internal/settlement/service.go](C:/Users/willj/src/iwx/go_backend/internal/settlement/service.go)
- exchange-core settlement application: [go_backend/internal/exchangecore/internal_service.go](C:/Users/willj/src/iwx/go_backend/internal/exchangecore/internal_service.go)

Flow:

1. `settlement` pulls `contract_resolved` from JetStream.
2. `HandleContractResolved` builds a correlation id:
   - trace id if present
   - else resolution event id
   - else a deterministic fallback
3. `settlement` calls `exchange-core` internal HTTP `SettleContract`.
4. `exchange-core-db` settlement application:
   - is keyed by `event_id`
   - is idempotent through `settlement_applications`
   - unwinds open reservations and locks
   - credits winning users
   - releases or consumes collateral
   - zeros positions
   - transitions contract to `settled`
5. `exchange-core` emits:
   - `projection_change.contract_changed`
   - `projection_change.settlement_changed`
   - `projection_change.user_state_changed` for affected users
6. `settlement` publishes `settlement_completed`.
7. projector refreshes contract and user settlement state into `read-db`.

State changes:

- `exchange-core-db.settlement_applications`
- `exchange-core-db.settlement_entries`
- `exchange-core-db.cash_accounts`
- `exchange-core-db.ledger_entries`
- `exchange-core-db.collateral_locks`
- `exchange-core-db.order_cash_reservations`
- `exchange-core-db.position_locks`
- `exchange-core-db.positions`
- `exchange-core-db.contracts`
- `exchange-core-db.outbox_events`
- `read-db` contract and user settlement projections

Potential failure points:

- `settlement` consumer lag
- malformed `contract_resolved` event
- internal HTTP failure calling `exchange-core`
- settlement business rule failure in `exchange-core`
- `settlement_completed` publish failure after settlement has already been applied
- projector lag or apply failure causing stale read-side settlement state

Partial-completion analysis:

- settlement itself does not own a DB, so the durable work happens inside `exchange-core`
- if `exchange-core` settlement application fails before commit, the consumer returns error and JetStream redelivers `contract_resolved`
- if `exchange-core` settlement application commits but `settlement_completed` publish fails afterward, the event handler returns failure and the same `contract_resolved` is retried
- retry is safe because `exchange-core` records `settlement_applications` keyed by the resolution `event_id`
- this means the largest remaining issue is not double settlement but delayed follow-up signaling; the contract can already be settled before `settlement_completed` is ever observed by downstream consumers
- projector lag can also make it look as if a contract is still unresolved or still holding balances/positions after settlement already committed

## Use Case 13: Projector Refresh

This is the central read-model maintenance flow.

Source code:

- consumer startup: [go_backend/internal/app/projector.go](C:/Users/willj/src/iwx/go_backend/internal/app/projector.go)
- bundle selection: [go_backend/internal/projector/service.go](C:/Users/willj/src/iwx/go_backend/internal/projector/service.go)
- read apply logic: [go_backend/internal/readprojection/applier.go](C:/Users/willj/src/iwx/go_backend/internal/readprojection/applier.go)

Flow:

1. owner service emits `projection_change` into its outbox.
2. owner outbox dispatcher publishes it to JetStream.
3. projector consumer receives the event.
4. projector switches by `ProjectionChange.Kind`:
   - `contract_changed` -> `exchange-core /internal/projection/contracts/:id`
   - `user_state_changed` -> `exchange-core /internal/projection/users/:id`
   - `settlement_changed` -> `exchange-core /internal/projection/settlements/contracts/:id`
   - `market_changed` -> `matcher /internal/projection/contracts/:id/market`
   - `order_command_changed` -> `matcher /internal/projection/order_commands/:id`
   - `oracle_state_changed` -> `oracle /internal/projection/contracts/:id/oracle`
   - `station_catalog_changed` -> `oracle /internal/projection/stations`
5. projector builds `events.ReadModelProjection`.
6. `readprojection.Applier` uses checkpoint keys and version numbers to decide whether to apply.
7. the applier replaces or upserts the relevant `read-db` tables.

State changes:

- `read-db.contracts`
- `read-db.orders`
- `read-db.executions`
- `read-db.market_snapshots`
- `read-db.order_commands`
- `read-db.cash_accounts`
- `read-db.positions`
- `read-db.position_locks`
- `read-db.collateral_locks`
- `read-db.cash_reservations`
- `read-db.oracle_observations`
- `read-db.contract_resolutions`
- `read-db.settlement_entries`
- `read-db.weather_stations`
- `read-db.projection_checkpoints`

Potential failure points:

- owner outbox publish lag
- projector consumer lag
- owner internal projection endpoint unavailable
- bundle fetch returns partial or missing data
- read projection replacement fails mid-apply
- checkpoint store failure
- stale or duplicate events are skipped by version gating, but a permanently missing newer event leaves `read-db` behind until another later event arrives

Partial-completion analysis:

- this is the most important partial-completion hotspot in the current code
- a single projection event can trigger several sequential write operations before checkpointing:
  - user-state projection replaces cash accounts, positions, position locks, collateral locks, cash reservations, and settlements one step at a time
  - market projection replaces orders, executions, and snapshots one step at a time
- if one replacement succeeds and a later one fails, `read-db` is temporarily internally inconsistent for that aggregate
- because the checkpoint is only recorded after the full apply completes, the same projection version will retry and should eventually converge
- the price of that design is temporary inconsistent reads, not permanent corruption
- if the owner-service bundle fetch itself returns an incomplete but syntactically valid view, the projector can faithfully project that incomplete state; the system assumes the bundle endpoints are authoritative snapshots

## Use Case 14: Browser Poll Read Cycle

Current UI behavior is browser polling, not SSE or websocket push.

Source code:

- `read-api` routes: [go_backend/internal/httpapi/server.go](C:/Users/willj/src/iwx/go_backend/internal/httpapi/server.go)
- React Query polling is in the Vite UI, not in backend push handlers

Flow:

1. browser periodically polls JSON endpoints.
2. `read-api` reads only from `read-db`.
3. UI freshness depends on:
   - projector lag
   - browser polling interval
   - any short-lived `read-api` cache entry for market state

Potential failure points:

- browser can observe eventual consistency gaps between write completion and projection completion
- stale data remains visible until the next poll
- if projector stops, reads remain available but stop changing

## Summary

The current architecture is rigorous in one specific way:

- every write-side use case has a single owner service
- cross-service propagation happens by explicit event or internal HTTP call
- `read-api` is query-only
- projector is the sole writer to `read-db`

The main operational risks still visible in the code are:

- outbox dispatch lag
- projector lag
- internal HTTP dependencies during projection and oracle/settlement workflows
- eventual consistency between owner DBs and `read-db`
- browser polling delay rather than immediate push

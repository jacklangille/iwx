# TODO

## End Goal

Build this as a centralized weather exchange with an internal ledger, not an onchain protocol.

- Users do not mint tokens into wallets.
- Users create exchange-managed weather contracts.
- Users post collateral into platform-controlled accounts.
- The platform issues internal `above` and `below` claim inventory.
- Trading, resolution, and settlement are all exchange-owned workflows.

## Target Platform Shape

Core service boundaries:

- `auth-service`
- `exchange-core-service`
- `matching-service`
- `read-api`
- React frontend

Near-term pragmatic shape:

- `auth-service`
- `exchange-core-service`
  - accounts
  - contracts
  - issuance
  - portfolio
  - oracle
  - settlement
- `matching-service`
- `read-api`

## Phase 0: Stabilize The Current Go Direction

Status: completed on the repo documentation and planning layer.

- Decide that the Go services are the only authoritative backend path.
- Freeze and remove obsolete backend implementation paths.
- Keep the current Go API and matcher as the canonical trading path.
- Remove obsolete backend assumptions from local docs and workflows.
- Add a top-level architecture doc describing:
  - auth-service
  - exchange-core-service
  - matching-service
  - read-api
  - React frontend

## Phase 1: Define Core Domain Models

Status: completed on the shared domain-definition layer.

- Finalize the centralized-exchange domain vocabulary:
  - contract
  - issuance batch
  - cash account
  - ledger entry
  - collateral lock
  - position
  - position lock
  - order
  - execution
  - market snapshot
  - oracle observation
  - resolution
  - settlement entry
- Define contract lifecycle states:
  - draft
  - pending_approval
  - pending_collateral
  - active
  - trading_closed
  - awaiting_resolution
  - resolved
  - settled
  - cancelled
- Define execution-side invariants:
  - matcher never owns balances
  - matcher never owns collateral
  - matcher only accepts/rejects/matches orders
- Define settlement rules precisely:
  - above payout rule
  - below payout rule
  - tie/threshold handling
  - cancellation/refund behavior

## Phase 2: Fix Service Boundaries In Code

Status: completed on the service-boundary layer.

- Split current Go code into service-oriented packages or binaries:
  - `cmd/auth`
  - `cmd/exchange-core`
  - `cmd/matcher`
  - `cmd/read-api`
- Move contract creation out of the matcher path and into exchange-core ownership.
- Keep matcher responsible only for:
  - orders
  - executions
  - book state
  - market snapshots
- Move contract metadata and lifecycle into exchange-core.
- Add explicit inter-service contracts:
  - contract activated
  - collateral locked
  - issuance completed
  - order accepted/rejected
  - execution created
  - contract resolved
  - settlement completed

## Phase 3: Introduce Proper Datastores Per Boundary

Status: completed on the storage-boundary and migration layer.

- Create separate DBs:
  - `auth-db`
  - `exchange-core-db`
  - `matcher-db`
  - `read-db`
- Stop treating the matcher replica as the main read model.
- Move `contracts` out of matcher ownership and into exchange-core DB.
- Keep matcher DB for:
  - orders
  - executions
  - snapshots
  - matcher command state
- Put read projections in `read-db` only.
- Add migration tooling for all Go-owned DBs.
- Remove dependence on Docker init SQL as the primary migration mechanism.

## Phase 4: Build Auth Properly

Status: completed on the authentication and identity-propagation layer.

- Finish the auth service:
  - users
  - password hashing
  - JWT issuance
  - session/token validation
- Add user identity propagation between services.
- Add service-to-service auth or signed internal tokens.
- Add user IDs to contract creation, orders, positions, and settlements.

## Phase 5: Build Accounts And Ledger

Status: completed on the exchange-core accounting layer.

- Create cash accounts model:
  - available balance
  - locked balance
  - total balance
- Add ledger entries as append-only accounting records.
- Add collateral locks tied to contract proposals and issuance.
- Add reservation records for buy-side order cash holds.
- Define all accounting entry types:
  - deposit
  - withdrawal
  - collateral_lock
  - collateral_release
  - order_cash_reserve
  - order_cash_release
  - settlement_credit
  - settlement_debit
- Add invariant checks so balances are never derived ad hoc from UI reads.

## Phase 6: Rework Contract Creation Flow

Status: completed on the exchange-core contract lifecycle layer.

- Replace current direct `create contract command` with a full lifecycle:
  - draft created by user
  - validation
  - moderation/approval
  - collateral requirement computed
  - collateral lock requested
  - activation
- Split contract metadata from contract rules if needed.
- Add explicit creator identity.
- Add rule versioning for resolution semantics.
- Add trading window enforcement.
- Add contract cancellation flow before activation.

## Phase 7: Add Issuance Service Logic

Status: completed on the exchange-core issuance and portfolio-credit layer.

- Create issuance batches tied to contracts and collateral locks.
- Issue paired internal claims:
  - above units
  - below units
- Record total issued supply per contract.
- Credit creator portfolio with issued positions.
- Add burn/redeem rules if inventory reduction before resolution is needed.
- Emit issuance events for downstream projections.

## Phase 8: Build Portfolio And Reservation Logic

Status: completed on the exchange-core reservation and pre-enqueue order flow layer.

- Add positions table keyed by:
  - `user_id`
  - `contract_id`
  - `side` (`above` / `below`)
- Add position locks for sell-side reservations.
- Add cash reservations for buy-side orders.
- Add APIs/internal methods for:
  - reserve cash
  - reserve inventory
  - release reservation
  - apply execution delta
- Make matcher order acceptance depend on prior reservation success, not raw user input alone.

## Phase 9: Refactor Matcher Around Executions Only

Status: completed on the matcher execution persistence and idempotent replay layer.

- Change matcher writes from `orders + snapshots only` to:
  - orders
  - executions
  - snapshots
- Add execution events as the canonical trade output.
- Keep single-writer partitioning by `contract_id`.
- Make matcher input an already risk-checked/reserved order command.
- Make matcher output execution events, not balance changes.
- Add idempotency guarantees for command replay and JetStream redelivery.
- Add durable recovery semantics for hot contract partitions.

## Phase 10: Build Read Models Properly

Status: completed on the read-db user projection, in-memory hot read cache, and SSE delivery layer for current platform state.

- Create read-api projections in `read-db` for:
  - contract catalog
  - market summaries
  - order book summaries
  - chart series
  - user positions
  - user balances
  - pending commands
  - settlement history
- Move hot market state to cache or in-memory projection, not replica Postgres.
- Add websocket or SSE push from read-api for:
  - market updates
  - order status
  - portfolio changes
  - settlement updates

## Phase 11: Add Oracle Service

Status: completed on the oracle service, observation storage, deterministic resolution, read projections, and contract-resolved event emission layer.

- Build oracle ingestion pipeline:
  - approved providers
  - normalized station data
  - metric normalization
  - observation windows
- Add deterministic resolution logic.
- Store:
  - source provider
  - station ID
  - raw observation
  - normalized value
  - rule version
  - final outcome
- Emit contract resolution events.
- Add replay and auditability for all resolutions.

## Phase 12: Build Settlement Service

Status: completed on the settlement consumer, exchange-core payout/refund processing, settlement history projections, and contract-settled transition layer.

- Consume resolution events.
- Determine winners and losers.
- Credit winning holders.
- Release unused collateral.
- Close positions.
- Create settlement ledger entries.
- Mark contracts settled.
- Handle cancelled or invalid markets and refunds.

## Phase 13: Migrate UI To React

Status: completed on the React frontend and legacy frontend removal layer.

- Vite React UI exists in `ui/`
- current screens:
  - trade / market detail
  - stations placeholder
  - portfolio / balances / settlements
- current React features:
  - React Query data fetching
  - bearer-token login flow
  - async order submission with command-status polling
  - SSE market updates
- remaining work:
  - contract creation / lifecycle screens
  - richer portfolio and settlement workflows

- Decide frontend structure:
  - React SPA
  - or React app served separately from read-api
- Define UI API contracts before building screens.
- Build or refine React screens for:
  - contract list
  - market detail page
  - order entry
  - command status
  - portfolio
  - balances
  - settlement history
- Build auth flow in React:
  - login
  - token storage
  - refresh/session handling
- Build async UX around command submission:
  - submit
  - queued
  - processing
  - succeeded/failed
- Add websocket or SSE client for live market updates.

## Phase 14: Operational Hardening

Current status:

- shared JSON log output is now enabled through `pkg/logging`
- HTTP services now generate and return:
  - `X-Request-ID`
  - `X-Trace-ID`
  - `X-Correlation-ID`
- request and trace IDs are propagated through:
  - read-api
  - auth
  - exchange-core
  - oracle HTTP
  - matcher order envelopes
  - oracle `contract_resolved` events
  - settlement `settlement_completed` events
- async matcher, oracle, and settlement logs now include trace-aware structured entries
- remaining work:
  - explicit metrics endpoints / time-series metrics
  - dead-letter handling
  - replay tooling
  - backup/recovery docs
  - admin tooling

- Add structured logging everywhere.
- Add tracing and correlation IDs across services.
- Add metrics:
  - queue lag
  - matcher latency
  - projection lag
  - failed commands
  - resolution timing
- Add dead-letter handling for failed events.
- Add replay tools for rebuilding read models.
- Add backup and recovery plan for each DB.
- Add admin tooling for:
  - contract approval
  - contract cancellation
  - manual resolution review
  - settlement replays

## Phase 15: Testing And Verification

Status: completed on the focused Go service verification layer.

- added service-level tests for:
  - contract lifecycle transitions and cancellation release behavior
  - reservation creation and release logic
  - matcher reservation validation rules
  - oracle resolution event propagation with trace IDs
  - settlement correlation and completion-event propagation
  - request / trace header propagation helpers

- Add service-level tests for:
  - ledger invariants
  - collateral locking
  - issuance
  - reservation logic
  - matching
  - settlement
- Add event-flow integration tests across services.
- Add idempotency tests for NATS redelivery.
- Add contract lifecycle end-to-end tests.
- Add UI integration tests for async command flows.

## Recommended Immediate Next 10 Tasks

1. Add service-to-service auth or signed internal tokens between Go services.
2. Start the React frontend with login, contract list, market detail, and async command UX.
3. Add integration tests for login, authenticated commands, and command ownership.
4. Add invariant and conflict tests for issuance, activation, reservation, and pre-activation cancellation.
5. Add lifecycle rules for re-issuance, burn, or redeem semantics before settlement.
6. Teach matcher/execution consumers to consume or release reservations from actual execution events.
7. Add cancellation/amend flows that unwind open-order reservations safely.
8. Replace polling-based SSE with event-driven push off matcher, oracle, and exchange-core events.
9. Start the React frontend’s auth and portfolio flows against the new read APIs.
10. Add replay/backfill tooling for oracle, settlement, and read projections.

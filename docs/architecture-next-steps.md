# Architecture Next Steps

This document captures the next major architecture improvements for IWX after the current service split, event hardening, and React migration work.

The goal is to make the next rounds of implementation incremental and explicit so they can be done one by one without re-deciding the direction each time.

## Current State Summary

The current platform shape is:

- `auth-service`
- `exchange-core-service`
- `matching-service`
- `oracle-service`
- `settlement-service`
- `read-api`
- React frontend

Current strengths:

- service/data ownership is mostly correct
- read paths are separated from write ownership
- matcher no longer owns balances or collateral
- execution application is idempotent
- settlement application is now idempotent by resolution event id
- oracle now has a real outbox for `contract_resolved`
- read projections now have stable ids and version checkpoints

Current weakness:

- reliability patterns are still inconsistent across services
- producer outboxes are now in place, but the UI still relies on straightforward browser polling
- production and local infra still need clearer separation in a few places

## Priority Order

Recommended execution order:

1. generalized outbox for `matcher` and `exchange-core`
2. dedicated projector path for `read-db`
3. optional event-driven realtime push
4. replay and repair tooling
5. simplify local infra
6. duplicate-market prevention
7. real oracle ingestion pipeline
8. architecture-level failure testing

## 1. Generalized Outbox For Matcher And Exchange-Core

Status: completed.

### Why

Right now the strongest publish reliability is on the oracle side. Matcher execution publishing and exchange-core read-model publishing still rely on direct publish plus idempotent downstream consumers.

That is workable, but it is not uniform.

### Target

Introduce a reusable outbox pattern for any service that:

- commits owned database state
- then needs to emit downstream events

### Scope

Apply it first to:

- `matcher -> execution_created`
- `exchange-core -> read-model projection events`

Potentially later to:

- exchange-core lifecycle/account events
- settlement completion events

### Deliverables

- reusable outbox schema and repository pattern
- background dispatcher per service
- stable event ids everywhere
- retry and failure tracking
- removal of direct publish-after-write for matcher/exchange-core paths

### Definition Of Done

- service commit and event publication are decoupled by durable outbox storage
- restarts do not lose unpublished events
- duplicate publish attempts remain safe

## 2. Dedicated Projector Path For Read DB

Status: completed.

### Why

Right now multiple services still emit read-model projection events directly. That means producers know about projection shapes and read concerns.

That works, but it is not the cleanest architecture.

### Target

Move toward:

- services emit business events only
- a dedicated projector service or projector worker consumes those events
- projector writes `read-db`

### Benefits

- less coupling between domain services and UI/read concerns
- simpler replay model
- clearer event catalog
- easier read-model rebuilds

### Deliverables

- projector service boundary
- projector-owned event handling
- removal of projection-specific emission logic from core domain services

### Definition Of Done

- `read-db` is updated only by projector code
- producers no longer need to know projection event formats

## 3. Optional Event-Driven Realtime Push

### Why

The UI currently polls normal JSON endpoints. That is simpler, but it still adds lag and repeated reads if lower-latency live updates become important.

### Target

If needed later, make read-api push updates based on applied events or projection changes instead of relying on browser polling.

### Scope

Use event-driven push for:

- market updates
- order command updates
- portfolio updates
- settlement updates

### Deliverables

- event-driven websocket or SSE broadcaster
- hook from successful projection application to fanout
- removal of polling-based refresh loops where possible

### Definition Of Done

- UI receives updates based on actual applied state changes
- no periodic poll loop is required for normal freshness

## 4. Replay And Repair Tooling

### Why

Event-driven systems need operational recovery tooling, not just code-level idempotency.

### Target

Add operator-friendly mechanisms to inspect, replay, and repair system state.

### Deliverables

- inspect unpublished outbox events
- replay outbox by event id or range
- rebuild read projections
- inspect failed application records
- inspect lagging consumers

### Definition Of Done

- an operator can recover from a partial event failure without direct manual DB surgery

## 5. Simplify Local Infra

Status: completed for local development.

### Why

The current local setup still contains transitional complexity, especially around matcher HA shape that the product path does not currently require.

### Target

Optimize local development for simplicity, not future HA simulation.

### Candidate changes

- remove matcher replica from local Compose
- reduce duplicated env/config drift
- keep one matcher Postgres for dev
- preserve HA patterns only in deployment docs, not default local runtime

### Definition Of Done

- local setup is easier to boot, inspect, and reset
- no product path depends on dev-only HA scaffolding

## 6. Duplicate-Market Prevention

Status: completed.

### Why

The `$1,000` collateral floor helps with spam, but it does not stop duplicate market definitions.

### Target

Prevent semantically identical overlapping markets.

### Candidate uniqueness dimensions

- station
- metric
- threshold
- measurement window
- trading window
- active/pending lifecycle overlap

### Definition Of Done

- users cannot create redundant active/pending markets that represent the same weather contract

## 7. Real Oracle Ingestion Pipeline

### Why

The oracle owns the station catalog, which is correct, but weather observations are still manually posted.

### Target

Make oracle a real ingestion service with provider adapters and normalization logic.

### Deliverables

- provider adapters
- station mapping
- scheduled pull jobs
- normalization pipeline
- source metadata capture
- deterministic ingest rules

### Definition Of Done

- oracle can ingest approved provider data automatically
- resolution uses internally owned normalized observations

## 8. Architecture-Level Failure Testing

### Why

Unit tests are not enough for event-driven correctness.

### Target

Add tests focused on crash/replay/ordering failure modes.

### Important scenarios

- DB commit succeeds but publish is delayed
- consumer fails after partial local work
- duplicate `execution_created`
- duplicate `contract_resolved`
- out-of-order projection events
- replay after restart

### Definition Of Done

- key reliability assumptions are verified with integration-style failure tests

## Suggested Execution Cadence

Use this order for implementation:

1. generalized outbox for matcher
2. generalized outbox for exchange-core
3. projector service extraction
4. event-driven push delivery
5. replay and repair tooling
6. local infra simplification
7. duplicate-market prevention
8. oracle ingestion pipeline
9. failure-mode integration testing

## Working Mode

When implementing these, prefer:

- one improvement per change set
- explicit migration files
- no mixed refactor + feature bundles unless necessary
- verification after each step with `go test ./...` and `go build ./...`

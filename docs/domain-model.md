# Domain Model

## Purpose

This document completes Phase 1 by defining the shared centralized-exchange domain model that the Go services should converge on.

It is intentionally platform-oriented, not onchain-oriented.

## Core Vocabulary

Core domain entities:

- `contract`
- `contract_rule`
- `cash_account`
- `ledger_entry`
- `collateral_lock`
- `issuance_batch`
- `position`
- `position_lock`
- `order`
- `execution`
- `market_snapshot`
- `oracle_observation`
- `contract_resolution`
- `settlement_entry`

## Contract Semantics

A contract is an exchange-issued market on a weather outcome.

The contract defines:

- market identity
- weather metric
- threshold
- measurement unit
- trading window
- measurement window
- resolution rule version

The contract itself is not a token. It is the market container that later supports internally-issued `above` and `below` claims.

## Contract Lifecycle

Canonical contract states:

- `draft`
- `pending_approval`
- `pending_collateral`
- `active`
- `trading_closed`
- `awaiting_resolution`
- `resolved`
- `settled`
- `cancelled`

Lifecycle intent:

- `draft`
  - user has proposed a market but it is not reviewable or tradable yet
- `pending_approval`
  - waiting for platform moderation/approval
- `pending_collateral`
  - approved market definition but collateral has not yet been fully locked
- `active`
  - collateral locked, issuance complete, market open for trading
- `trading_closed`
  - no new trading, waiting for observation window completion
- `awaiting_resolution`
  - trading closed, oracle resolution pending
- `resolved`
  - final weather outcome has been determined
- `settled`
  - payouts and collateral release completed
- `cancelled`
  - market voided and any locked value refunded or released

## Claims And Positions

The system issues internal paired claims:

- `above`
- `below`

These are not blockchain tokens. They are internal exchange positions.

Each user position must distinguish:

- available quantity
- locked quantity
- total quantity

Sell orders reserve position inventory before the matcher sees the order.

## Cash And Ledger

All cash-like behavior is internal account accounting.

Each account tracks:

- available balance
- locked balance
- total balance

All balance-changing behavior should be justified by append-only ledger entries.

Required ledger entry families:

- `deposit`
- `withdrawal`
- `collateral_lock`
- `collateral_release`
- `order_cash_reserve`
- `order_cash_release`
- `settlement_credit`
- `settlement_debit`

## Collateral And Issuance

Collateral is not onchain escrow. It is an internal liability lock.

Expected flow:

1. creator proposes a contract
2. platform approves the definition
3. required collateral is computed
4. cash is locked in a collateral lock
5. issuance batch creates paired `above` and `below` claims
6. creator receives internal inventory in portfolio storage

Issuance must be represented explicitly as issuance batches tied to:

- contract
- creator
- collateral lock
- issued quantity on both sides

## Orders And Executions

Orders represent intent to trade.

Executions represent actual matches and are the durable downstream trading event.

An execution should carry:

- contract ID
- buy order ID
- sell order ID
- buyer user ID
- seller user ID
- price
- quantity
- sequence
- occurred-at timestamp

## Matcher Invariants

The matcher must not own:

- cash balances
- collateral
- portfolio source-of-truth
- settlement
- oracle resolution
- contract moderation lifecycle

The matcher may own:

- active market order books
- open orders
- execution sequencing
- execution creation
- market snapshots needed for market-data views

Operational invariant:

- matcher input should already have passed reservation checks
- matcher output should be executions, not balance mutations

## Oracle And Resolution

Resolution must be deterministic and auditable.

Store for each resolution:

- provider name
- station ID
- observed metric
- observation window
- rule version
- resolved value
- final outcome

Canonical outcomes:

- `above`
- `below`
- `cancelled`

## Settlement Rules

Baseline rule set for now:

- `above` pays if the resolved metric is strictly above the threshold
- `below` pays if the resolved metric is at or below the threshold
- tie behavior belongs to `below` by default

If the platform later supports different threshold semantics, that must be versioned in `contract_rule.rule_version`.

Settlement behavior:

- winning holders receive payout credits
- losing holders receive no payout
- remaining locked collateral is released
- cancelled markets refund and unwind without winner/loser payouts

## Read-Model Boundaries

`market_snapshots` are read-model data.

They are valid for:

- chart series
- recent market state
- market summary history

They are not the source of truth for:

- balances
- positions
- collateral
- settlement

## Code Alignment

Shared domain definitions live in:

- [go_backend/internal/domain/types.go](/C:/Users/willj/src/iwx/go_backend/internal/domain/types.go)
- [go_backend/internal/domain/states.go](/C:/Users/willj/src/iwx/go_backend/internal/domain/states.go)
- [go_backend/internal/domain/exchange_models.go](/C:/Users/willj/src/iwx/go_backend/internal/domain/exchange_models.go)

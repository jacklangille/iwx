# IWX

IWX is being rebuilt as a centralized weather exchange with a Go backend and a React frontend.

## Current Repo Status

- The Go services are the authoritative backend path.
- The backend code lives in `go_backend/`.
- The frontend code lives in `ui/`.

Current Go runtime pieces:

- `cmd/auth`
- `cmd/read-api`
- `cmd/exchange-core`
- `cmd/matcher`
- `cmd/oracle`
- `cmd/settlement`
- local Postgres + NATS infrastructure via [docker-compose.yml](/C:/Users/willj/src/iwx/docker-compose.yml)
- Vite React UI in [ui/](/C:/Users/willj/src/iwx/ui)

Important architecture notes:

- matcher currently owns order matching, executions, and market snapshots
- exchange-core currently owns contract lifecycle, contract rules, and contract command state
- oracle now owns weather observations and final contract resolution records
- matcher now publishes `execution_created` events after successful fills
- exchange-core now consumes `execution_created` to apply balance and position changes
- read-api now consumes `execution_created` to refresh market projections and invalidate hot cache entries
- settlement now consumes `contract_resolved` events and settles contracts against exchange-core balances and positions
- read-api now serves both market reads and user dashboard projections from `read-db`
- authenticated writes now carry `user_id` / `creator_user_id` through the Go services
- long term, contract lifecycle, collateral, issuance, and settlement should continue expanding inside `exchange-core-service`
- the target platform architecture is documented in [docs/architecture.md](/C:/Users/willj/src/iwx/docs/architecture.md)

## Local Dev

Start the full local stack:

```powershell
docker compose up -d
```

This now starts:

- Postgres for `auth`, `exchange-core`, `oracle`, `read`, and `matcher`
- NATS with JetStream
- `auth`
- `exchange-core`
- `read-api`
- `matcher`
- `oracle`
- `settlement`

Useful local endpoints:

- read-api: `http://127.0.0.1:8080`
- auth: `http://127.0.0.1:8081`
- exchange-core: `http://127.0.0.1:8082`
- oracle: `http://127.0.0.1:8083`

If you want to run the Go processes manually instead of through Compose:

```powershell
cd go_backend
go run .\cmd\read-api
go run .\cmd\exchange-core
go run .\cmd\matcher
go run .\cmd\auth
go run .\cmd\oracle
go run .\cmd\settlement
```

Start the React UI in another shell:

```powershell
cd ui
npm install
npm run dev
```

The Vite app proxies local API traffic to:

- read-api: `http://127.0.0.1:8080`
- auth: `http://127.0.0.1:8081`
- exchange-core: `http://127.0.0.1:8082`

See [ui/.env.example](/C:/Users/willj/src/iwx/ui/.env.example) for frontend overrides.

Authenticated write endpoints currently are:

- `POST /api/auth/login`
- `GET /api/accounts/me`
- `GET /api/accounts/me/ledger`
- `GET /api/accounts/me/collateral_locks`
- `GET /api/accounts/me/cash_reservations`
- `POST /api/accounts/deposits`
- `POST /api/accounts/withdrawals`
- `POST /api/accounts/collateral_locks`
- `POST /api/accounts/collateral_locks/:id/release`
- `POST /api/accounts/cash_reservations`
- `POST /api/accounts/cash_reservations/:id/release`
- `POST /api/contracts`
- `GET /api/contracts/:id`
- `POST /api/contracts/:id/submit_for_approval`
- `POST /api/contracts/:id/approve`
- `GET /api/contracts/:id/collateral_requirement`
- `POST /api/contracts/:id/collateral_locks`
- `GET /api/contracts/:id/issuance_batches`
- `POST /api/contracts/:id/issuance_batches`
- `POST /api/contracts/:id/activate`
- `POST /api/contracts/:id/cancel`
- `GET /api/positions/me`
- `GET /api/positions/me/locks`
- `GET /api/contract_commands/:command_id`
- `POST /api/orders`
- `GET /api/order_commands/:command_id`
- `GET /api/contracts/:id/executions`
- `GET /api/contracts/:id/observations`
- `GET /api/contracts/:id/resolution`
- `GET /api/contracts/:id/settlements`
- `GET /api/me/account`
- `GET /api/me/positions`
- `GET /api/me/position_locks`
- `GET /api/me/collateral_locks`
- `GET /api/me/cash_reservations`
- `GET /api/me/portfolio`
- `GET /api/me/settlements`
- `GET /api/stream/contracts/:id/market`
- `GET /api/stream/me/order_commands/:id`
- `GET /api/stream/me/portfolio`
- internal oracle endpoints live on `cmd/oracle`:
  - `POST /api/oracle/observations`
  - `GET /api/oracle/contracts/:id/observations`
  - `POST /api/oracle/contracts/:id/resolve`
  - `GET /api/oracle/contracts/:id/resolution`

The Go services load `go_backend/.env` automatically in local development.
Each service also runs its required database migrations on boot.

Operational notes:

- Go service logs now emit JSON lines to stdout and per-service log files.
- HTTP services return `X-Request-ID`, `X-Trace-ID`, and `X-Correlation-ID`.
- Order, oracle-resolution, and settlement flows now propagate trace IDs across service boundaries.

## Planning Docs

- target architecture: [docs/architecture.md](/C:/Users/willj/src/iwx/docs/architecture.md)
- domain model: [docs/domain-model.md](/C:/Users/willj/src/iwx/docs/domain-model.md)
- implementation roadmap: [TODO.md](/C:/Users/willj/src/iwx/TODO.md)

# iwx

Minimal Phoenix web application.

## Requirements

- Elixir 1.15 or newer
- Erlang/OTP compatible with your Elixir version
- PostgreSQL
- Node.js and npm

## Run locally

```sh
mix setup
cd assets && npm install && cd ..
mix ecto.create
mix ecto.migrate
mix phx.server
```

The app runs at <http://localhost:4000>.

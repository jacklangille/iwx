import Config

config :frontend, FrontendWeb.Endpoint,
  http: [ip: {127, 0, 0, 1}, port: 4002],
  secret_key_base: "nYSoJMCOQsK18WujEFA/IeVasfad4ctFSgt4SwFW2wGIdqpmuac8AdaiwmEfJ3ah",
  server: false

config :frontend, Frontend.Repo,
  username: "jwl",
  password: "",
  hostname: "localhost",
  database: "frontend_test#{System.get_env("MIX_TEST_PARTITION")}",
  pool: Ecto.Adapters.SQL.Sandbox,
  pool_size: 10

config :logger, level: :warning
config :phoenix, :plug_init_mode, :runtime
config :phoenix_live_view, enable_expensive_runtime_checks: true
config :phoenix, sort_verified_routes_query_params: true

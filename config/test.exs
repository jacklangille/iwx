import Config

config :frontend, FrontendWeb.Endpoint,
  http: [ip: {127, 0, 0, 1}, port: 4002],
  secret_key_base: "nYSoJMCOQsK18WujEFA/IeVasfad4ctFSgt4SwFW2wGIdqpmuac8AdaiwmEfJ3ah",
  server: false

config :logger, level: :warning
config :phoenix, :plug_init_mode, :runtime
config :phoenix_live_view, enable_expensive_runtime_checks: true
config :phoenix, sort_verified_routes_query_params: true

import Config

config :frontend, FrontendWeb.Endpoint,
  http: [ip: {127, 0, 0, 1}],
  check_origin: false,
  code_reloader: true,
  debug_errors: true,
  secret_key_base: "aFbsxI7j4Mfx/8capXzB6ShxUMvLHMrAru1+gaNiOqqu8jgrXfu0+sNnpxM/CP6m",
  watchers: [
    npm: ["run", "dev", cd: Path.expand("../assets", __DIR__)]
  ],
  live_reload: [
    patterns: [
      ~r"priv/static/.*(js|css|png|jpeg|jpg|gif|svg)$",
      ~r"lib/frontend_web/(controllers|components|router)/.*(ex|heex)$",
      ~r"lib/frontend_web/templates/.*(eex|heex)$"
    ]
  ]

config :frontend, Frontend.Repo,
  username: "jwl",
  password: "",
  hostname: "localhost",
  database: "frontend_dev",
  show_sensitive_data_on_connection_error: true,
  pool_size: 10

config :logger, :default_formatter, format: "[$level] $message\n"
config :phoenix, :stacktrace_depth, 20
config :phoenix, :plug_init_mode, :runtime

config :phoenix_live_view,
  debug_heex_annotations: true,
  debug_attributes: true,
  enable_expensive_runtime_checks: true

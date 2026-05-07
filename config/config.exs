# This file is responsible for configuring your application
# and its dependencies with the aid of the Config module.
#
# This configuration file is loaded before any dependency and
# is restricted to this project.

import Config

config :frontend, ecto_repos: [Frontend.Repo]

config :frontend, FrontendWeb.Endpoint,
  url: [host: "localhost"],
  adapter: Bandit.PhoenixAdapter,
  pubsub_server: Frontend.PubSub

config :logger, :default_formatter,
  format: "$time $metadata[$level] $message\n",
  metadata: [:request_id]

import_config "#{config_env()}.exs"

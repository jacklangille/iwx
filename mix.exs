defmodule Frontend.MixProject do
  use Mix.Project

  def project do
    [
      app: :frontend,
      version: "0.1.0",
      elixir: "~> 1.15",
      start_permanent: Mix.env() == :prod,
      aliases: aliases(),
      deps: deps(),
      compilers: [:phoenix_live_view] ++ Mix.compilers(),
      listeners: [Phoenix.CodeReloader]
    ]
  end

  def application do
    [
      mod: {Frontend.Application, []},
      extra_applications: [:logger]
    ]
  end

  def cli do
    [
      preferred_envs: [precommit: :test]
    ]
  end

  defp deps do
    [
      {:phoenix, "~> 1.8.5"},
      {:phoenix_html, "~> 4.1"},
      {:phoenix_live_view, "~> 1.1.0"},
      {:phoenix_live_reload, "~> 1.5", only: :dev},
      {:bandit, "~> 1.5"},
      {:postgrex, ">= 0.0.0"},
      {:ecto_sql, "~> 3.11"},
      {:jason, "~> 1.4"}
    ]
  end

  defp aliases do
    [
      setup: ["deps.get"],
      test: ["test"],
      precommit: ["compile --warnings-as-errors", "deps.unlock --unused", "format", "test"]
    ]
  end
end

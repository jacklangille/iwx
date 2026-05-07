defmodule Frontend.Application do
  use Application

  @impl true
  def start(_type, _args) do
    children = [
      {Phoenix.PubSub, name: Frontend.PubSub},
      Frontend.Repo,
      FrontendWeb.Endpoint
    ]

    Supervisor.start_link(children, strategy: :one_for_one, name: Frontend.Supervisor)
  end

  @impl true
  def config_change(changed, _new, removed) do
    FrontendWeb.Endpoint.config_change(changed, removed)
    :ok
  end
end

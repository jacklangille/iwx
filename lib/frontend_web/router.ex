defmodule FrontendWeb.Router do
  use FrontendWeb, :router

  pipeline :browser do
    plug :accepts, ["html"]
    plug :put_root_layout, html: {FrontendWeb.Layouts, :root}
  end

  pipeline :api do
    plug :accepts, ["json"]
  end

  scope "/", FrontendWeb do
    pipe_through :browser

    get "/", PageController, :home
    get "/stations", StationExplorerController, :index
  end

  scope "/api", FrontendWeb do
    pipe_through :api
    get "/orders", OrderController, :index
    post "/orders", OrderController, :create
    get "/contracts", ContractController, :index
    get "/contracts/:id/market_state", MarketStateController, :show
    get "/contracts/:id/market_snapshots", MarketSnapshotController, :index
  end
end

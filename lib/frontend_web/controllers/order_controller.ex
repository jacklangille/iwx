defmodule FrontendWeb.OrderController do
  use FrontendWeb, :controller

  alias Frontend.Exchange
  alias Frontend.Exchange.Order
  alias FrontendWeb.MarketBroadcaster
  alias FrontendWeb.OrderJSON

  def create(conn, params) do
    case Exchange.place_order(params) do
      {:ok, %Order{} = order} ->
        MarketBroadcaster.broadcast_market_state(order.contract_id)
        json(conn, OrderJSON.show(%{order: order}))

      {:ok, %{status: "filled"}} ->
        MarketBroadcaster.broadcast_market_state(params["contract_id"])
        json(conn, OrderJSON.show(%{order: %{status: "filled"}}))

      {:error, changeset} ->
        conn
        |> put_status(:bad_request)
        |> json(%{
          errors: Ecto.Changeset.traverse_errors(changeset, fn {msg, _opts} -> msg end)
        })
    end
  end

  def index(conn, params) do
    orders = Exchange.list_orders(params)

    json(conn, OrderJSON.index(%{orders: orders}))
  end
end

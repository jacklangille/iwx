defmodule FrontendWeb.OrderController do
  use FrontendWeb, :controller

  alias Frontend.Exchange
  alias Frontend.Exchange.Order

  def create(conn, params) do
    case Exchange.place_order(params) do
      {:ok, %Order{} = order} ->
        json(conn, %{
          id: order.id,
          contract_id: order.contract_id,
          token_type: order.token_type,
          order_side: order.order_side,
          price: Decimal.to_string(order.price, :normal),
          quantity: order.quantity,
          status: order.status
        })

      {:ok, %{status: "filled"}} ->
        json(conn, %{
          status: "filled"
        })

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

    json(
      conn,
      Enum.map(orders, fn order ->
        %{
          id: order.id,
          contract_id: order.contract_id,
          token_type: order.token_type,
          order_side: order.order_side,
          price: Decimal.to_string(order.price, :normal),
          quantity: order.quantity,
          status: order.status
        }
      end)
    )
  end
end

defmodule FrontendWeb.OrderJSON do
  alias Frontend.Exchange.Order
  alias FrontendWeb.NumberJSON

  def index(%{orders: orders}) do
    Enum.map(orders, &data/1)
  end

  def show(%{order: %Order{} = order}), do: data(order)
  def show(%{order: %{status: "filled"}}), do: %{status: "filled"}

  defp data(%Order{} = order) do
    %{
      id: order.id,
      contract_id: order.contract_id,
      token_type: order.token_type,
      order_side: order.order_side,
      price: NumberJSON.number_to_string(order.price),
      quantity: order.quantity,
      status: order.status
    }
  end
end

defmodule Frontend.Exchange.Matching do
  import Ecto.Query

  alias Ecto.Changeset
  alias Frontend.Exchange.Order
  alias Frontend.Exchange.OrderBook
  alias Frontend.Exchange.Orders
  alias Frontend.Exchange.Snapshots
  alias Frontend.Repo

  def process_order(attrs) do
    changeset = Order.changeset(%Order{}, attrs)

    if changeset.valid? do
      order = Changeset.apply_changes(changeset)
      attrs = normalized_attrs(attrs, order)

      case consume_matches(
             order.contract_id,
             order.token_type,
             order.order_side,
             order.price,
             order.quantity
           ) do
        0 ->
          record_market_snapshot(order.contract_id)
          %{status: "filled"}

        remaining_quantity ->
          open_order =
            attrs
            |> Map.put("quantity", remaining_quantity)
            |> Orders.insert_open()

          record_market_snapshot(order.contract_id)
          open_order
      end
    else
      Repo.rollback(changeset)
    end
  end

  defp normalized_attrs(_attrs, order) do
    %{
      "contract_id" => order.contract_id,
      "token_type" => order.token_type,
      "order_side" => order.order_side,
      "price" => order.price,
      "quantity" => order.quantity
    }
  end

  defp consume_matches(_contract_id, _token_type, _order_side, _price, 0), do: 0

  defp consume_matches(contract_id, token_type, order_side, price, remaining_quantity) do
    case find_match(contract_id, token_type, order_side, price) do
      nil ->
        remaining_quantity

      matched_order ->
        matched_quantity = min(remaining_quantity, matched_order.quantity)

        Orders.update_matched!(matched_order, matched_quantity)

        consume_matches(
          contract_id,
          token_type,
          order_side,
          price,
          remaining_quantity - matched_quantity
        )
    end
  end

  defp find_match(contract_id, token_type, "bid", price) do
    Repo.one(
      from(o in Order,
        where:
          o.contract_id == ^contract_id and
            o.token_type == ^token_type and
            o.order_side == "ask" and
            o.status == "open" and
            o.price <= ^price,
        order_by: [asc: o.price, asc: o.inserted_at],
        limit: 1,
        lock: "FOR UPDATE"
      )
    )
  end

  defp find_match(contract_id, token_type, "ask", price) do
    Repo.one(
      from(o in Order,
        where:
          o.contract_id == ^contract_id and
            o.token_type == ^token_type and
            o.order_side == "bid" and
            o.status == "open" and
            o.price >= ^price,
        order_by: [desc: o.price, asc: o.inserted_at],
        limit: 1,
        lock: "FOR UPDATE"
      )
    )
  end

  defp record_market_snapshot(contract_id) do
    contract_id
    |> Orders.list_open_for_contract()
    |> OrderBook.summary()
    |> then(&Snapshots.record(contract_id, &1))
  end
end

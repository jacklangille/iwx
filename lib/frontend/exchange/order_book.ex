defmodule Frontend.Exchange.OrderBook do
  def components(orders) do
    order_book = build(orders)
    best = best_quotes(order_book)

    %{
      orders: orders,
      order_book: order_book,
      best: best,
      mid: midpoint_quotes(best)
    }
  end

  def summary(orders) do
    %{best: best, mid: mid} = components(orders)

    mid_price =
      cond do
        mid.above != nil and mid.below != nil -> average_two(mid.above, mid.below)
        mid.above != nil -> mid.above
        mid.below != nil -> mid.below
        true -> nil
      end

    %{
      best: best,
      mid: mid,
      mid_price: mid_price
    }
  end

  def liquidity_totals(orders) do
    Enum.reduce(orders, %{above: 0, below: 0}, fn order, totals ->
      case outcome_side(order.token_type) do
        :above -> %{totals | above: totals.above + order.quantity}
        :below -> %{totals | below: totals.below + order.quantity}
        _other -> totals
      end
    end)
  end

  def above_below_bid_gap(%{above: %{bid: nil}}), do: nil
  def above_below_bid_gap(%{below: %{bid: nil}}), do: nil

  def above_below_bid_gap(%{above: %{bid: above_bid}, below: %{bid: below_bid}}) do
    above_bid
    |> Decimal.sub(below_bid)
    |> Decimal.abs()
  end

  defp build(orders) do
    %{
      above: %{
        bid: aggregate_levels(orders, :above, :bid, :desc),
        ask: aggregate_levels(orders, :above, :ask, :asc)
      },
      below: %{
        bid: aggregate_levels(orders, :below, :bid, :desc),
        ask: aggregate_levels(orders, :below, :ask, :asc)
      }
    }
  end

  defp aggregate_levels(orders, outcome_side, quote_side, sort_direction) do
    stored_token_type = stored_token_type(outcome_side)
    stored_order_side = stored_order_side(quote_side)

    orders
    |> Enum.filter(&(&1.token_type == stored_token_type and &1.order_side == stored_order_side))
    |> Enum.reduce(%{}, fn order, levels ->
      key = Decimal.to_string(order.price, :normal)

      Map.update(
        levels,
        key,
        %{price: order.price, quantity: order.quantity},
        fn level -> %{level | quantity: level.quantity + order.quantity} end
      )
    end)
    |> Map.values()
    |> Enum.sort_by(& &1.price, &compare_decimals(&1, &2, sort_direction))
  end

  defp compare_decimals(left, right, :asc), do: Decimal.compare(left, right) != :gt
  defp compare_decimals(left, right, :desc), do: Decimal.compare(left, right) != :lt

  defp best_quotes(order_book) do
    %{
      above: %{
        bid: best_level_price(order_book.above.bid),
        ask: best_level_price(order_book.above.ask)
      },
      below: %{
        bid: best_level_price(order_book.below.bid),
        ask: best_level_price(order_book.below.ask)
      }
    }
  end

  defp best_level_price([level | _levels]), do: level.price
  defp best_level_price([]), do: nil

  defp midpoint_quotes(best) do
    %{
      above: midpoint(best.above.bid, best.above.ask),
      below: midpoint(best.below.bid, best.below.ask)
    }
  end

  defp midpoint(bid, ask) do
    cond do
      bid != nil and ask != nil ->
        bid
        |> Decimal.add(ask)
        |> Decimal.div(Decimal.new("2"))

      bid != nil ->
        bid

      ask != nil ->
        ask

      true ->
        nil
    end
  end

  defp average_two(a, b) do
    a
    |> Decimal.add(b)
    |> Decimal.div(Decimal.new("2"))
  end

  defp stored_token_type(:above), do: "above"
  defp stored_token_type(:below), do: "below"

  defp stored_order_side(:bid), do: "bid"
  defp stored_order_side(:ask), do: "ask"

  defp outcome_side("above"), do: :above
  defp outcome_side("below"), do: :below
  defp outcome_side(_stored_token_type), do: nil
end

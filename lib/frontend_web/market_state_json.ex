defmodule FrontendWeb.MarketStateJSON do
  alias FrontendWeb.NumberJSON

  def serialize(market_state) do
    %{
      contract_id: market_state.contract_id,
      sequence: market_state.sequence,
      as_of: timestamp_to_string(market_state.as_of),
      order_book: %{
        above: %{
          bid: Enum.map(market_state.order_book.above.bid, &serialize_level/1),
          ask: Enum.map(market_state.order_book.above.ask, &serialize_level/1)
        },
        below: %{
          bid: Enum.map(market_state.order_book.below.bid, &serialize_level/1),
          ask: Enum.map(market_state.order_book.below.ask, &serialize_level/1)
        }
      },
      summary: %{
        best: %{
          above: %{
            bid: NumberJSON.number_to_string(market_state.summary.best.above.bid),
            ask: NumberJSON.number_to_string(market_state.summary.best.above.ask)
          },
          below: %{
            bid: NumberJSON.number_to_string(market_state.summary.best.below.bid),
            ask: NumberJSON.number_to_string(market_state.summary.best.below.ask)
          }
        },
        mid: %{
          above: NumberJSON.number_to_string(market_state.summary.mid.above),
          below: NumberJSON.number_to_string(market_state.summary.mid.below)
        },
        liquidity: market_state.summary.liquidity,
        above_below_bid_gap: NumberJSON.number_to_string(market_state.summary.above_below_bid_gap)
      }
    }
  end

  def serialize_chart_point(nil), do: nil

  def serialize_chart_point(chart_point) do
    %{
      bucket_start: timestamp_to_string(chart_point.bucket_start),
      inserted_at: timestamp_to_string(chart_point.inserted_at),
      mid_above: NumberJSON.number_to_string(chart_point.mid_above),
      mid_below: NumberJSON.number_to_string(chart_point.mid_below),
      best_above: NumberJSON.number_to_string(chart_point.best_above),
      best_below: NumberJSON.number_to_string(chart_point.best_below)
    }
  end

  defp serialize_level(level) do
    %{
      price: NumberJSON.number_to_string(level.price),
      quantity: level.quantity
    }
  end

  defp timestamp_to_string(nil), do: nil
  defp timestamp_to_string(%NaiveDateTime{} = value), do: NaiveDateTime.to_iso8601(value)
  defp timestamp_to_string(%DateTime{} = value), do: DateTime.to_iso8601(value)
end

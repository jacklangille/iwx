defmodule FrontendWeb.ContractController do
  use FrontendWeb, :controller

  alias Frontend.Repo
  alias Frontend.Contracts.Contract
  alias Frontend.Exchange

  def index(conn, _params) do
    contracts =
      Repo.all(Contract)
      |> Enum.map(fn c ->
        summary = Exchange.market_summary(c.id)
        as_of = Exchange.latest_market_snapshot_timestamp(c.id)
        sequence = Exchange.latest_market_sequence(c.id)

        %{
          id: c.id,
          as_of: as_of,
          sequence: sequence,
          name: c.name,
          region: c.region,
          metric: c.metric,
          status: c.status,
          threshold: c.threshold,
          multiplier: c.multiplier,
          measurement_unit: c.measurement_unit,
          trading_period_start: c.trading_period_start,
          trading_period_end: c.trading_period_end,
          measurement_period_start: c.measurement_period_start,
          measurement_period_end: c.measurement_period_end,
          data_provider_name: c.data_provider_name,
          data_provider_station_mode: c.data_provider_station_mode,
          description: c.description,
          best_above_bid: number_to_string(summary.best.above.bid),
          best_below_bid: number_to_string(summary.best.below.bid),
          mid_above: number_to_string(summary.mid.above),
          mid_below: number_to_string(summary.mid.below),
          mid_price: number_to_string(summary.mid_price)
        }
      end)

    json(conn, contracts)
  end

  defp number_to_string(nil), do: nil
  defp number_to_string(%Decimal{} = value), do: Decimal.to_string(value, :normal)

  defp number_to_string(value) when is_float(value),
    do: :erlang.float_to_binary(value, decimals: 2)

  defp number_to_string(value) when is_integer(value), do: Integer.to_string(value)
end

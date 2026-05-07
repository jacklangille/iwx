defmodule FrontendWeb.ContractJSON do
  alias Frontend.Contracts.Contract
  alias FrontendWeb.NumberJSON

  def index(%{contract_summaries: contract_summaries}) do
    Enum.map(contract_summaries, &data/1)
  end

  defp data(%{contract: %Contract{} = contract, market: market}) do
    summary = market.summary

    %{
      id: contract.id,
      as_of: market.as_of,
      sequence: market.sequence,
      name: contract.name,
      region: contract.region,
      metric: contract.metric,
      status: contract.status,
      threshold: contract.threshold,
      multiplier: contract.multiplier,
      measurement_unit: contract.measurement_unit,
      trading_period_start: contract.trading_period_start,
      trading_period_end: contract.trading_period_end,
      measurement_period_start: contract.measurement_period_start,
      measurement_period_end: contract.measurement_period_end,
      data_provider_name: contract.data_provider_name,
      data_provider_station_mode: contract.data_provider_station_mode,
      description: contract.description,
      best_above_bid: NumberJSON.number_to_string(summary.best.above.bid),
      best_below_bid: NumberJSON.number_to_string(summary.best.below.bid),
      mid_above: NumberJSON.number_to_string(summary.mid.above),
      mid_below: NumberJSON.number_to_string(summary.mid.below),
      mid_price: NumberJSON.number_to_string(summary.mid_price)
    }
  end
end

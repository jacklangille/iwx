defmodule FrontendWeb.MarketBroadcaster do
  alias Frontend.Exchange
  alias FrontendWeb.Endpoint
  alias FrontendWeb.MarketStateJSON

  @market_state_updated_event "market_state_updated"

  def broadcast_market_state(contract_id) do
    chart_point =
      contract_id
      |> Exchange.current_market_chart_point()
      |> MarketStateJSON.serialize_chart_point()

    market_state =
      contract_id
      |> Exchange.market_state()
      |> MarketStateJSON.serialize()

    payload = %{
      type: @market_state_updated_event,
      contract_id: market_state.contract_id,
      sequence: market_state.sequence,
      as_of: market_state.as_of,
      market_state: market_state,
      chart_point: chart_point
    }

    Endpoint.broadcast(
      "contract:#{market_state.contract_id}",
      @market_state_updated_event,
      payload
    )
  end
end

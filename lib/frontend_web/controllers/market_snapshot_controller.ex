defmodule FrontendWeb.MarketSnapshotController do
  use FrontendWeb, :controller

  alias Frontend.Exchange.ChartConfig
  alias Frontend.Exchange
  alias FrontendWeb.MarketStateJSON

  def index(conn, %{"id" => contract_id} = params) do
    chart_config = ChartConfig.from_params(params)
    snapshots = Exchange.list_market_chart_series(contract_id, Map.to_list(chart_config))

    json(conn, %{
      config: chart_config,
      points: Enum.map(snapshots, &MarketStateJSON.serialize_chart_point/1)
    })
  end
end

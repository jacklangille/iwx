defmodule FrontendWeb.MarketStateController do
  use FrontendWeb, :controller

  alias Frontend.Exchange
  alias FrontendWeb.MarketStateJSON

  def show(conn, %{"id" => contract_id}) do
    market_state = Exchange.market_state(contract_id)

    json(conn, MarketStateJSON.serialize(market_state))
  end
end

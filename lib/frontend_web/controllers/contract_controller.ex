defmodule FrontendWeb.ContractController do
  use FrontendWeb, :controller

  alias Frontend.Contracts
  alias FrontendWeb.ContractJSON

  def index(conn, _params) do
    contract_summaries = Contracts.list_contract_summaries()

    json(conn, ContractJSON.index(%{contract_summaries: contract_summaries}))
  end
end

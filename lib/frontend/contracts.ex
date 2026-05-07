defmodule Frontend.Contracts do
  import Ecto.Query

  alias Frontend.Contracts.Contract
  alias Frontend.Exchange
  alias Frontend.Repo

  def list_contracts do
    Repo.all(Contract)
  end

  def list_contract_summaries do
    list_contracts()
    |> Enum.map(&contract_summary/1)
  end

  def contract_exists?(contract_id) do
    Repo.exists?(from(c in Contract, where: c.id == ^contract_id))
  end

  defp contract_summary(%Contract{} = contract) do
    summary = Exchange.market_summary(contract.id)

    %{
      contract: contract,
      market: %{
        as_of: Exchange.latest_market_snapshot_timestamp(contract.id),
        sequence: Exchange.latest_market_sequence(contract.id),
        summary: summary
      }
    }
  end
end

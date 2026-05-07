defmodule Frontend.Exchange.Snapshots do
  import Ecto.Query

  alias Frontend.Exchange.MarketSnapshot
  alias Frontend.Repo

  def record(contract_id, summary) do
    %MarketSnapshot{}
    |> MarketSnapshot.changeset(%{
      contract_id: contract_id,
      best_above: summary.best.above.bid,
      best_below: summary.best.below.bid,
      mid_above: summary.mid.above,
      mid_below: summary.mid.below
    })
    |> Repo.insert!()
  end

  def list_since(contract_id, window_start) do
    Repo.all(
      from(s in MarketSnapshot,
        where: s.contract_id == ^contract_id and s.inserted_at >= ^window_start,
        order_by: [asc: s.inserted_at]
      )
    )
  end

  def latest_timestamp(contract_id) do
    Repo.one(
      from(s in MarketSnapshot,
        where: s.contract_id == ^contract_id,
        order_by: [desc: s.inserted_at],
        limit: 1,
        select: s.inserted_at
      )
    )
  end

  def latest_sequence(contract_id) do
    Repo.one(
      from(s in MarketSnapshot,
        where: s.contract_id == ^contract_id,
        order_by: [desc: s.inserted_at],
        limit: 1,
        select: s.id
      )
    )
  end
end

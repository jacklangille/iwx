defmodule Frontend.Exchange.MarketSnapshot do
  use Ecto.Schema
  import Ecto.Changeset

  schema "market_snapshots" do
    field(:best_above, :decimal)
    field(:best_below, :decimal)
    field(:mid_above, :decimal)
    field(:mid_below, :decimal)

    belongs_to(:contract, Frontend.Contracts.Contract)

    timestamps(updated_at: false)
  end

  def changeset(snapshot, attrs) do
    snapshot
    |> cast(attrs, [:contract_id, :best_above, :best_below, :mid_above, :mid_below])
    |> validate_required([:contract_id])
  end
end

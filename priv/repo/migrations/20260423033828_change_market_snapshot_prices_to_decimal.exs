defmodule Frontend.Repo.Migrations.ChangeMarketSnapshotPricesToDecimal do
  use Ecto.Migration

  def up do
    alter table(:market_snapshots) do
      modify :best_above, :decimal, precision: 10, scale: 2
      modify :best_below, :decimal, precision: 10, scale: 2
      modify :mid_above, :decimal, precision: 10, scale: 2
      modify :mid_below, :decimal, precision: 10, scale: 2
    end
  end

  def down do
    alter table(:market_snapshots) do
      modify :best_above, :integer
      modify :best_below, :integer
      modify :mid_above, :float
      modify :mid_below, :float
    end
  end
end

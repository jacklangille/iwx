defmodule Frontend.Repo.Migrations.CreateMarketSnapshots do
  use Ecto.Migration

  def change do
    create table(:market_snapshots) do
      add :contract_id, references(:contracts, on_delete: :delete_all), null: false
      add :best_above, :integer
      add :best_below, :integer
      add :mid_above, :float
      add :mid_below, :float

      timestamps(updated_at: false)
    end

    create index(:market_snapshots, [:contract_id, :inserted_at])
  end
end

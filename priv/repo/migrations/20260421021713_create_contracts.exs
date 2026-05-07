defmodule Frontend.Repo.Migrations.CreateContracts do
  use Ecto.Migration

  def change do
    create table(:contracts) do
      add :name, :string, null: false
      add :region, :string, null: false
      add :metric, :string, null: false
      add :status, :string, null: false
      add :threshold, :integer
      add :multiplier, :integer

      timestamps()
    end
  end
end

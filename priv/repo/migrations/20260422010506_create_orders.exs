defmodule Frontend.Repo.Migrations.CreateOrders do
  use Ecto.Migration
  def change do
    create table(:orders) do
      add :contract_id, :integer, null: false
      add :side, :string, null: false
      add :price, :integer, null: false
      add :quantity, :integer, null: false
      add :status, :string, null: false

      timestamps()
    end

    create index(:orders, [:contract_id])
  end
end

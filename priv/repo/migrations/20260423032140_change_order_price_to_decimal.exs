defmodule Frontend.Repo.Migrations.ChangeOrderPriceToDecimal do
  use Ecto.Migration

  def up do
    alter table(:orders) do
      modify :price, :decimal, precision: 10, scale: 2, null: false
    end
  end

  def down do
    alter table(:orders) do
      modify :price, :integer, null: false
    end
  end
end

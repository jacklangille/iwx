defmodule Frontend.Repo.Migrations.AddContractFkToOrders do
  use Ecto.Migration

  def up do
    execute """
    DELETE FROM orders
    WHERE contract_id NOT IN (SELECT id FROM contracts)
    """

    drop_if_exists index(:orders, [:contract_id])

    alter table(:orders) do
      modify :contract_id, references(:contracts, on_delete: :delete_all),
        null: false,
        from: :integer
    end

    create index(:orders, [:contract_id])
  end

  def down do
    drop_if_exists index(:orders, [:contract_id])

    alter table(:orders) do
      modify :contract_id, :integer, null: false
    end

    create index(:orders, [:contract_id])
  end
end

defmodule Frontend.Repo.Migrations.ReplaceSideWithTokenTypeAndOrderSideOnOrders do
  use Ecto.Migration

  def change do
    alter table(:orders) do
      add :token_type, :string
      add :order_side, :string
    end

    execute """
    UPDATE orders
    SET token_type = CASE side
      WHEN 'long' THEN 'above'
      WHEN 'short' THEN 'below'
      ELSE side
    END
    """

    execute "UPDATE orders SET order_side = 'bid'"

    alter table(:orders) do
      modify :token_type, :string, null: false
      modify :order_side, :string, null: false
      remove :side
    end
  end
end

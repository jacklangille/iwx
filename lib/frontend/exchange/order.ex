defmodule Frontend.Exchange.Order do
  use Ecto.Schema
  import Ecto.Changeset

  schema "orders" do
    field(:token_type, :string)
    field(:order_side, :string)
    field(:price, :decimal)
    field(:quantity, :integer)
    field(:status, :string)

    belongs_to(:contract, Frontend.Contracts.Contract)

    timestamps()
  end

  def changeset(order, attrs) do
    order
    |> cast(attrs, [:contract_id, :token_type, :order_side, :price, :quantity])
    |> validate_required([:contract_id, :token_type, :order_side, :price, :quantity])
    |> validate_inclusion(:token_type, ["above", "below"])
    |> validate_inclusion(:order_side, ["bid", "ask"])
    |> validate_number(:price, greater_than: 0)
    |> validate_number(:quantity, greater_than: 0)
    |> foreign_key_constraint(:contract_id)
    |> put_change(:status, "open")
  end
end

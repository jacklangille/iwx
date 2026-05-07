defmodule Frontend.Exchange.Orders do
  import Ecto.Query

  alias Frontend.Exchange.Order
  alias Frontend.Repo

  def list_open(params \\ %{}) do
    case Map.get(params, "contract_id") do
      nil -> Repo.all(from(o in Order, where: o.status == "open"))
      contract_id -> list_open_for_contract(contract_id)
    end
  end

  def list_open_for_contract(contract_id) do
    Repo.all(
      from(o in Order,
        where: o.contract_id == ^contract_id and o.status == "open"
      )
    )
  end

  def insert_open(attrs) do
    case Repo.insert(Order.changeset(%Order{}, attrs)) do
      {:ok, order} -> order
      {:error, changeset} -> Repo.rollback(changeset)
    end
  end

  def update_matched!(matched_order, matched_quantity) do
    updated_attrs =
      if matched_quantity == matched_order.quantity do
        %{quantity: 0, status: "filled"}
      else
        %{quantity: matched_order.quantity - matched_quantity}
      end

    matched_order
    |> Ecto.Changeset.change(updated_attrs)
    |> Repo.update!()
  end
end

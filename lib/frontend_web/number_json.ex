defmodule FrontendWeb.NumberJSON do
  def number_to_string(nil), do: nil
  def number_to_string(%Decimal{} = value), do: Decimal.to_string(value, :normal)

  def number_to_string(value) when is_float(value),
    do: :erlang.float_to_binary(value, decimals: 2)

  def number_to_string(value) when is_integer(value), do: Integer.to_string(value)
end

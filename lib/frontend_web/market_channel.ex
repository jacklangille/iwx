defmodule FrontendWeb.MarketChannel do
  use Phoenix.Channel

  alias Frontend.Exchange

  @impl true
  def join("contract:" <> contract_id, _payload, socket) do
    case Integer.parse(contract_id) do
      {id, ""} ->
        join_contract_topic(id, contract_id, socket)

      _other ->
        {:error, %{reason: "invalid_contract"}}
    end
  end

  def join(_topic, _payload, _socket), do: {:error, %{reason: "invalid_topic"}}

  defp join_contract_topic(id, contract_id, socket) do
    cond do
      Integer.to_string(id) != contract_id ->
        {:error, %{reason: "invalid_contract"}}

      Exchange.contract_exists?(id) ->
        {:ok, assign(socket, :contract_id, id)}

      true ->
        {:error, %{reason: "not_found"}}
    end
  end
end

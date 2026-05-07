defmodule Frontend.Exchange.ChartConfig do
  @lookback_seconds 24 * 60 * 60
  @bucket_seconds 5 * 60
  @supported_lookback_seconds [30 * 24 * 60 * 60, 5 * 24 * 60 * 60, 24 * 60 * 60]
  @supported_bucket_seconds [5 * 60, 60 * 60, 24 * 60 * 60]

  def defaults do
    %{
      lookback_seconds: @lookback_seconds,
      bucket_seconds: @bucket_seconds
    }
  end

  def from_params(params) do
    defaults()
    |> maybe_put_supported(
      :lookback_seconds,
      params["lookback_seconds"],
      @supported_lookback_seconds
    )
    |> maybe_put_supported(:bucket_seconds, params["bucket_seconds"], @supported_bucket_seconds)
  end

  defp maybe_put_supported(config, key, value, supported_values) do
    parsed_value = parse_positive_integer(value)

    if parsed_value in supported_values do
      Map.put(config, key, parsed_value)
    else
      config
    end
  end

  defp parse_positive_integer(value) when is_binary(value) do
    case Integer.parse(value) do
      {integer, ""} when integer > 0 -> integer
      _ -> nil
    end
  end

  defp parse_positive_integer(_value), do: nil
end

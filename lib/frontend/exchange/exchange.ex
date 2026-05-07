defmodule Frontend.Exchange do
  alias Frontend.Exchange.ChartConfig
  alias Frontend.Exchange.ChartSeries
  alias Frontend.Exchange.Matching
  alias Frontend.Exchange.OrderBook
  alias Frontend.Exchange.Orders
  alias Frontend.Exchange.Snapshots
  alias Frontend.Repo

  def place_order(attrs) do
    Repo.transaction(fn -> Matching.process_order(attrs) end)
  end

  def list_orders(attrs \\ %{}) do
    Orders.list_open(attrs)
  end

  def market_state(contract_id) do
    %{orders: orders, order_book: order_book, best: best, mid: mid} =
      contract_id
      |> Orders.list_open_for_contract()
      |> OrderBook.components()

    %{
      contract_id: normalize_contract_id(contract_id),
      sequence: latest_market_sequence(contract_id),
      as_of: latest_market_snapshot_timestamp(contract_id),
      order_book: order_book,
      summary: %{
        best: best,
        mid: mid,
        liquidity: OrderBook.liquidity_totals(orders),
        above_below_bid_gap: OrderBook.above_below_bid_gap(best)
      }
    }
  end

  def market_summary(contract_id) do
    contract_id
    |> Orders.list_open_for_contract()
    |> OrderBook.summary()
  end

  def record_market_snapshot(contract_id) do
    contract_id
    |> market_summary()
    |> then(&Snapshots.record(contract_id, &1))
  end

  def list_market_chart_series(contract_id, opts \\ []) do
    chart_config =
      ChartConfig.defaults()
      |> Map.merge(Map.new(opts))

    lookback_seconds = chart_config.lookback_seconds
    bucket_seconds = chart_config.bucket_seconds

    window_end = NaiveDateTime.utc_now() |> NaiveDateTime.truncate(:second)
    window_start = NaiveDateTime.add(window_end, -lookback_seconds, :second)
    first_bucket = ChartSeries.bucket_start(window_start, bucket_seconds)
    last_bucket = ChartSeries.bucket_start(window_end, bucket_seconds)

    contract_id
    |> Snapshots.list_since(window_start)
    |> ChartSeries.bucket_snapshots(bucket_seconds, first_bucket, last_bucket)
  end

  def current_market_chart_point(contract_id, opts \\ []) do
    contract_id
    |> list_market_chart_series(opts)
    |> List.last()
  end

  def latest_market_snapshot_timestamp(contract_id) do
    Snapshots.latest_timestamp(contract_id)
  end

  def latest_market_sequence(contract_id) do
    Snapshots.latest_sequence(contract_id)
  end

  defp normalize_contract_id(contract_id) when is_integer(contract_id), do: contract_id

  defp normalize_contract_id(contract_id) when is_binary(contract_id) do
    case Integer.parse(contract_id) do
      {id, ""} -> id
      _other -> contract_id
    end
  end

  defp normalize_contract_id(contract_id), do: contract_id
end

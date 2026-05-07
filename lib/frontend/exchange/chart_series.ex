defmodule Frontend.Exchange.ChartSeries do
  def bucket_snapshots(snapshots, bucket_seconds, first_bucket, last_bucket) do
    snapshots_by_bucket = group_snapshots_by_bucket(snapshots, bucket_seconds)
    buckets = bucket_range(first_bucket, last_bucket, bucket_seconds)

    {_carry, rows} =
      Enum.reduce(buckets, {%{}, []}, fn bucket, {carry, rows} ->
        bucket_snapshot =
          snapshots_by_bucket
          |> Map.get(bucket, [])
          |> reduce_bucket_snapshot(bucket)

        carry = carry_forward_snapshot(carry, bucket_snapshot)

        row = %{
          bucket_start: bucket,
          inserted_at: carry[:inserted_at] || bucket,
          mid_above: carry[:mid_above],
          mid_below: carry[:mid_below],
          best_above: carry[:best_above],
          best_below: carry[:best_below]
        }

        {carry, [row | rows]}
      end)

    Enum.reverse(rows)
  end

  def bucket_start(nil, _bucket_seconds), do: nil

  def bucket_start(%DateTime{} = timestamp, bucket_seconds) do
    timestamp
    |> DateTime.to_unix()
    |> div(bucket_seconds)
    |> Kernel.*(bucket_seconds)
    |> DateTime.from_unix!()
    |> DateTime.to_naive()
    |> NaiveDateTime.truncate(:second)
  end

  def bucket_start(%NaiveDateTime{} = timestamp, bucket_seconds) do
    timestamp
    |> DateTime.from_naive!("Etc/UTC")
    |> DateTime.to_unix()
    |> div(bucket_seconds)
    |> Kernel.*(bucket_seconds)
    |> DateTime.from_unix!()
    |> DateTime.to_naive()
    |> NaiveDateTime.truncate(:second)
  end

  defp group_snapshots_by_bucket(snapshots, bucket_seconds) do
    Enum.group_by(snapshots, fn snapshot ->
      bucket_start(snapshot.inserted_at, bucket_seconds)
    end)
  end

  defp bucket_range(first_bucket, last_bucket, bucket_seconds) do
    Stream.iterate(first_bucket, &NaiveDateTime.add(&1, bucket_seconds, :second))
    |> Enum.take_while(&(NaiveDateTime.compare(&1, last_bucket) != :gt))
  end

  defp reduce_bucket_snapshot([], _bucket), do: %{}

  defp reduce_bucket_snapshot(snapshots, bucket) do
    Enum.reduce(snapshots, %{bucket_start: bucket}, fn snapshot, carry ->
      carry
      |> put_non_nil(:inserted_at, snapshot.inserted_at)
      |> put_non_nil(:mid_above, snapshot.mid_above)
      |> put_non_nil(:mid_below, snapshot.mid_below)
      |> put_non_nil(:best_above, snapshot.best_above)
      |> put_non_nil(:best_below, snapshot.best_below)
    end)
  end

  defp carry_forward_snapshot(carry, bucket_snapshot) do
    carry
    |> put_non_nil(:inserted_at, bucket_snapshot[:inserted_at])
    |> put_non_nil(:mid_above, bucket_snapshot[:mid_above])
    |> put_non_nil(:mid_below, bucket_snapshot[:mid_below])
    |> put_non_nil(:best_above, bucket_snapshot[:best_above])
    |> put_non_nil(:best_below, bucket_snapshot[:best_below])
  end

  defp put_non_nil(map, _key, nil), do: map
  defp put_non_nil(map, key, value), do: Map.put(map, key, value)
end

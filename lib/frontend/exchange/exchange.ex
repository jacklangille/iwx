defmodule Frontend.Exchange do
  alias Frontend.Contracts.Contract
  alias Frontend.Exchange.ChartConfig
  alias Frontend.Exchange.Order
  alias Frontend.Repo
  alias Frontend.Exchange.MarketSnapshot
  alias FrontendWeb.Endpoint
  alias FrontendWeb.MarketStateJSON
  import Ecto.Query

  @market_state_updated_event "market_state_updated"

  def place_order(attrs) do
    case Repo.transaction(fn -> process_order(attrs) end) do
      {:ok, result} ->
        attrs
        |> Map.fetch!("contract_id")
        |> broadcast_market_state()

        {:ok, result}

      {:error, reason} ->
        {:error, reason}
    end
  end

  def list_orders(attrs \\ %{}) do
    case Map.get(attrs, "contract_id") do
      nil ->
        Repo.all(from(o in Order, where: o.status == "open"))

      contract_id ->
        Repo.all(
          from(o in Order,
            where: o.contract_id == ^contract_id and o.status == "open"
          )
        )
    end
  end

  def contract_exists?(contract_id) do
    Repo.exists?(from(c in Contract, where: c.id == ^contract_id))
  end

  def broadcast_market_state(contract_id) do
    chart_point =
      contract_id
      |> current_market_chart_point()
      |> MarketStateJSON.serialize_chart_point()

    market_state =
      contract_id
      |> market_state()
      |> MarketStateJSON.serialize()

    payload = %{
      type: @market_state_updated_event,
      contract_id: market_state.contract_id,
      sequence: market_state.sequence,
      as_of: market_state.as_of,
      market_state: market_state,
      chart_point: chart_point
    }

    Endpoint.broadcast(
      "contract:#{market_state.contract_id}",
      @market_state_updated_event,
      payload
    )
  end

  def market_state(contract_id) do
    %{orders: orders, order_book: order_book, best: best, mid: mid} =
      market_components(contract_id)

    %{
      contract_id: normalize_contract_id(contract_id),
      sequence: latest_market_sequence(contract_id),
      as_of: latest_market_snapshot_timestamp(contract_id),
      order_book: order_book,
      summary: %{
        best: best,
        mid: mid,
        liquidity: liquidity_totals(orders),
        above_below_bid_gap: above_below_bid_gap(best)
      }
    }
  end

  def market_summary(contract_id) do
    %{best: best, mid: mid} = market_components(contract_id)

    mid_price =
      cond do
        mid.above != nil and mid.below != nil -> average_two(mid.above, mid.below)
        mid.above != nil -> mid.above
        mid.below != nil -> mid.below
        true -> nil
      end

    %{
      best: best,
      mid: mid,
      mid_price: mid_price
    }
  end

  def record_market_snapshot(contract_id) do
    summary = market_summary(contract_id)

    %MarketSnapshot{}
    |> MarketSnapshot.changeset(%{
      contract_id: contract_id,
      best_above: summary.best.above.bid,
      best_below: summary.best.below.bid,
      mid_above: summary.mid.above,
      mid_below: summary.mid.below
    })
    |> Repo.insert!()
  end

  def list_market_chart_series(contract_id, opts \\ []) do
    chart_config =
      ChartConfig.defaults()
      |> Map.merge(Map.new(opts))

    lookback_seconds = chart_config.lookback_seconds
    bucket_seconds = chart_config.bucket_seconds

    window_end = NaiveDateTime.utc_now() |> NaiveDateTime.truncate(:second)
    window_start = NaiveDateTime.add(window_end, -lookback_seconds, :second)
    first_bucket = bucket_start(window_start, bucket_seconds)
    last_bucket = bucket_start(window_end, bucket_seconds)

    contract_id
    |> list_market_snapshots_since(window_start)
    |> bucket_market_snapshots(bucket_seconds, first_bucket, last_bucket)
  end

  def current_market_chart_point(contract_id, opts \\ []) do
    contract_id
    |> list_market_chart_series(opts)
    |> List.last()
  end

  def latest_market_snapshot_timestamp(contract_id) do
    Repo.one(
      from(s in MarketSnapshot,
        where: s.contract_id == ^contract_id,
        order_by: [desc: s.inserted_at],
        limit: 1,
        select: s.inserted_at
      )
    )
  end

  def latest_market_sequence(contract_id) do
    Repo.one(
      from(s in MarketSnapshot,
        where: s.contract_id == ^contract_id,
        order_by: [desc: s.inserted_at],
        limit: 1,
        select: s.id
      )
    )
  end

  defp list_market_snapshots_since(contract_id, window_start) do
    Repo.all(
      from(s in MarketSnapshot,
        where: s.contract_id == ^contract_id and s.inserted_at >= ^window_start,
        order_by: [asc: s.inserted_at]
      )
    )
  end

  defp bucket_market_snapshots(snapshots, bucket_seconds, first_bucket, last_bucket) do
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

  defp bucket_start(nil, _bucket_seconds), do: nil

  defp bucket_start(%DateTime{} = timestamp, bucket_seconds) do
    timestamp
    |> DateTime.to_unix()
    |> div(bucket_seconds)
    |> Kernel.*(bucket_seconds)
    |> DateTime.from_unix!()
    |> DateTime.to_naive()
    |> NaiveDateTime.truncate(:second)
  end

  defp bucket_start(%NaiveDateTime{} = timestamp, bucket_seconds) do
    timestamp
    |> DateTime.from_naive!("Etc/UTC")
    |> DateTime.to_unix()
    |> div(bucket_seconds)
    |> Kernel.*(bucket_seconds)
    |> DateTime.from_unix!()
    |> DateTime.to_naive()
    |> NaiveDateTime.truncate(:second)
  end

  defp list_open_orders_for_contract(contract_id) do
    Repo.all(
      from(o in Order,
        where: o.contract_id == ^contract_id and o.status == "open"
      )
    )
  end

  defp market_components(contract_id) do
    orders = list_open_orders_for_contract(contract_id)
    order_book = build_order_book(orders)
    best = best_quotes(order_book)

    %{
      orders: orders,
      order_book: order_book,
      best: best,
      mid: midpoint_quotes(best)
    }
  end

  defp build_order_book(orders) do
    %{
      above: %{
        bid: aggregate_levels(orders, :above, :bid, :desc),
        ask: aggregate_levels(orders, :above, :ask, :asc)
      },
      below: %{
        bid: aggregate_levels(orders, :below, :bid, :desc),
        ask: aggregate_levels(orders, :below, :ask, :asc)
      }
    }
  end

  defp aggregate_levels(orders, outcome_side, quote_side, sort_direction) do
    stored_token_type = stored_token_type(outcome_side)
    stored_order_side = stored_order_side(quote_side)

    orders
    |> Enum.filter(&(&1.token_type == stored_token_type and &1.order_side == stored_order_side))
    |> Enum.reduce(%{}, fn order, levels ->
      key = Decimal.to_string(order.price, :normal)

      Map.update(
        levels,
        key,
        %{price: order.price, quantity: order.quantity},
        fn level ->
          %{level | quantity: level.quantity + order.quantity}
        end
      )
    end)
    |> Map.values()
    |> Enum.sort_by(& &1.price, &compare_decimals(&1, &2, sort_direction))
  end

  defp compare_decimals(left, right, :asc), do: Decimal.compare(left, right) != :gt
  defp compare_decimals(left, right, :desc), do: Decimal.compare(left, right) != :lt

  defp best_quotes(order_book) do
    %{
      above: %{
        bid: best_level_price(order_book.above.bid),
        ask: best_level_price(order_book.above.ask)
      },
      below: %{
        bid: best_level_price(order_book.below.bid),
        ask: best_level_price(order_book.below.ask)
      }
    }
  end

  defp best_level_price([level | _levels]), do: level.price
  defp best_level_price([]), do: nil

  defp midpoint_quotes(best) do
    %{
      above: midpoint(best.above.bid, best.above.ask),
      below: midpoint(best.below.bid, best.below.ask)
    }
  end

  defp liquidity_totals(orders) do
    Enum.reduce(orders, %{above: 0, below: 0}, fn order, totals ->
      case outcome_side(order.token_type) do
        :above -> %{totals | above: totals.above + order.quantity}
        :below -> %{totals | below: totals.below + order.quantity}
        _other -> totals
      end
    end)
  end

  defp above_below_bid_gap(%{above: %{bid: nil}}), do: nil
  defp above_below_bid_gap(%{below: %{bid: nil}}), do: nil

  defp above_below_bid_gap(%{above: %{bid: above_bid}, below: %{bid: below_bid}}) do
    above_bid
    |> Decimal.sub(below_bid)
    |> Decimal.abs()
  end

  defp normalize_contract_id(contract_id) when is_integer(contract_id), do: contract_id

  defp normalize_contract_id(contract_id) when is_binary(contract_id) do
    case Integer.parse(contract_id) do
      {id, ""} -> id
      _other -> contract_id
    end
  end

  defp normalize_contract_id(contract_id), do: contract_id

  defp midpoint(bid, ask) do
    cond do
      bid != nil and ask != nil ->
        bid
        |> Decimal.add(ask)
        |> Decimal.div(Decimal.new("2"))

      bid != nil ->
        bid

      ask != nil ->
        ask

      true ->
        nil
    end
  end

  defp average_two(a, b) do
    a
    |> Decimal.add(b)
    |> Decimal.div(Decimal.new("2"))
  end

  defp process_order(attrs) do
    contract_id = attrs["contract_id"]
    outcome_side = outcome_side(attrs["token_type"])
    quote_side = quote_side(attrs["order_side"])
    price = to_decimal(attrs["price"])
    quantity = attrs["quantity"]

    attrs = Map.put(attrs, "price", price)

    case consume_matches(contract_id, outcome_side, quote_side, price, quantity) do
      0 ->
        record_market_snapshot(contract_id)
        %{status: "filled"}

      remaining_quantity ->
        order =
          attrs
          |> Map.put("quantity", remaining_quantity)
          |> insert_open_order()

        record_market_snapshot(contract_id)
        order
    end
  end

  defp consume_matches(_contract_id, _outcome_side, _quote_side, _price, 0), do: 0

  defp consume_matches(contract_id, outcome_side, quote_side, price, remaining_quantity) do
    case find_match(contract_id, outcome_side, quote_side, price) do
      nil ->
        remaining_quantity

      matched_order ->
        matched_quantity = min(remaining_quantity, matched_order.quantity)

        update_matched_order!(matched_order, matched_quantity)

        consume_matches(
          contract_id,
          outcome_side,
          quote_side,
          price,
          remaining_quantity - matched_quantity
        )
    end
  end

  defp find_match(contract_id, outcome_side, :bid, price) do
    stored_token_type = stored_token_type(outcome_side)
    stored_resting_quote_side = stored_order_side(:ask)

    Repo.one(
      from(o in Order,
        where:
          o.contract_id == ^contract_id and
            o.token_type == ^stored_token_type and
            o.order_side == ^stored_resting_quote_side and
            o.status == "open" and
            o.price <= ^price,
        order_by: [asc: o.price, asc: o.inserted_at],
        limit: 1
      )
    )
  end

  defp find_match(contract_id, outcome_side, :ask, price) do
    stored_token_type = stored_token_type(outcome_side)
    stored_resting_quote_side = stored_order_side(:bid)

    Repo.one(
      from(o in Order,
        where:
          o.contract_id == ^contract_id and
            o.token_type == ^stored_token_type and
            o.order_side == ^stored_resting_quote_side and
            o.status == "open" and
            o.price >= ^price,
        order_by: [desc: o.price, asc: o.inserted_at],
        limit: 1
      )
    )
  end

  defp find_match(_, _, _, _), do: nil

  defp update_matched_order!(matched_order, matched_quantity) do
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

  defp insert_open_order(attrs) do
    case Repo.insert(Order.changeset(%Order{}, attrs)) do
      {:ok, order} ->
        order

      {:error, changeset} ->
        Repo.rollback(changeset)
    end
  end

  defp to_decimal(%Decimal{} = value), do: value
  defp to_decimal(value), do: Decimal.new(to_string(value))

  defp stored_token_type(:above), do: "above"
  defp stored_token_type(:below), do: "below"
  defp stored_token_type(_outcome_side), do: nil

  defp stored_order_side(:bid), do: "bid"
  defp stored_order_side(:ask), do: "ask"
  defp stored_order_side(_quote_side), do: nil

  defp outcome_side("above"), do: :above
  defp outcome_side("below"), do: :below
  defp outcome_side(_stored_token_type), do: nil

  defp quote_side("bid"), do: :bid
  defp quote_side("ask"), do: :ask
  defp quote_side(_stored_order_side), do: nil
end

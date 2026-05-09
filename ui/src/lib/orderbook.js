function toNumber(value) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : null;
}

function normalizeLevels(levels, side) {
  return (levels || []).map((level) => {
    const price = toNumber(level.price);
    const quantity = toNumber(level.quantity);
    const priceKey = price == null ? String(level.price) : String(price);

    return {
      side,
      key: `${side}:${priceKey}`,
      price,
      quantity,
      total: price != null && quantity != null ? price * quantity : null,
    };
  });
}

function buildSpread(bids, asks) {
  const bestBid = bids[0]?.price;
  const bestAsk = asks[0]?.price;

  if (bestBid == null || bestAsk == null) {
    return { spread: null, percent: null };
  }

  const spread = bestAsk - bestBid;
  const midpoint = (bestBid + bestAsk) / 2;

  if (!Number.isFinite(spread) || !Number.isFinite(midpoint) || midpoint <= 0) {
    return { spread: null, percent: null };
  }

  return {
    spread,
    percent: (spread / midpoint) * 100,
  };
}

export function buildOrderBookLadder(orderBook, view) {
  const activeBook = view === "above" ? orderBook?.above : orderBook?.below;
  const visibleBids = normalizeLevels(activeBook?.bid, "bid");
  const visibleAsks = normalizeLevels(activeBook?.ask, "ask");
  const displayedAsks = [...visibleAsks].reverse();
  const maxVisibleQuantity = [...visibleAsks, ...visibleBids].reduce(
    (max, level) => Math.max(max, level.quantity || 0),
    0,
  );

  const withDisplayData = (level) => ({
    ...level,
    barWidth:
      maxVisibleQuantity > 0 && level.quantity != null
        ? (level.quantity / maxVisibleQuantity) * 100
        : 0,
  });

  return {
    asks: displayedAsks.map(withDisplayData),
    bids: visibleBids.map(withDisplayData),
    spread: buildSpread(visibleBids, visibleAsks),
  };
}

export function renderOrderBookToggle({ state, elements }) {
  const { viewSelect } = elements.orderBook;

  if (viewSelect) {
    viewSelect.value = state.ui.orderBookView;
  }
}

const previousVisibleLevelKeysByView = {
  above: null,
  below: null,
};

function toNumber(value) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : null;
}

function formatPrice(value) {
  const parsed = toNumber(value);
  return parsed == null ? "-" : parsed.toFixed(2);
}

function formatSize(value) {
  const parsed = toNumber(value);
  if (parsed == null) return "-";
  return Number.isInteger(parsed) ? String(parsed) : parsed.toFixed(2);
}

function formatTotal(value) {
  const parsed = toNumber(value);
  return parsed == null ? "-" : parsed.toFixed(2);
}

function formatSpreadPercent(value) {
  const parsed = toNumber(value);
  return parsed == null ? "-" : `${parsed.toFixed(2)}%`;
}

function normalizeLevels(levels, side) {
  return (levels ?? []).map((level) => {
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

function buildSpread(visibleBids, visibleAsks) {
  const bestBid = visibleBids[0]?.price;
  const bestAsk = visibleAsks[0]?.price;

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

export function buildOrderBookLadder({ orderBook, view }) {
  const activeBook = view === "above" ? orderBook?.above : orderBook?.below;
  const visibleBids = normalizeLevels(activeBook?.bid, "bid");
  const visibleAsks = normalizeLevels(activeBook?.ask, "ask");
  const displayedAsks = [...visibleAsks].reverse();
  const maxVisibleQuantity = [...visibleAsks, ...visibleBids].reduce(
    (max, level) => Math.max(max, level.quantity ?? 0),
    0,
  );
  const previousVisibleLevelKeys = previousVisibleLevelKeysByView[view];
  const currentVisibleLevelKeys = new Set(
    [...visibleAsks, ...visibleBids].map((level) => level.key),
  );
  const newVisibleLevelKeys =
    previousVisibleLevelKeys == null
      ? new Set()
      : new Set(
          [...currentVisibleLevelKeys].filter(
            (key) => !previousVisibleLevelKeys.has(key),
          ),
        );

  previousVisibleLevelKeysByView[view] = currentVisibleLevelKeys;

  const withDisplayData = (level) => ({
    ...level,
    isNew: newVisibleLevelKeys.has(level.key),
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

function renderLevelRow(level) {
  const row = document.createElement("div");
  row.className = [
    "order-book__row",
    `order-book__row--${level.side}`,
    level.isNew && "order-book__row--new",
  ]
    .filter(Boolean)
    .join(" ");
  row.style.setProperty("--bar-width", `${level.barWidth}%`);

  const price = document.createElement("span");
  price.className = "order-book__cell order-book__cell--price";
  price.textContent = formatPrice(level.price);

  const quantity = document.createElement("span");
  quantity.className = "order-book__cell order-book__cell--size";
  quantity.textContent = formatSize(level.quantity);

  const total = document.createElement("span");
  total.className = "order-book__cell order-book__cell--total";
  total.textContent = formatTotal(level.total);

  row.appendChild(price);
  row.appendChild(quantity);
  row.appendChild(total);

  return row;
}

function renderSpreadRow(spread) {
  const row = document.createElement("div");
  row.className = "order-book__spread";
  row.dataset.orderbookSpreadRow = "";

  const label = document.createElement("span");
  label.className = "order-book__spread-label";
  label.textContent = "Spread";

  const absolute = document.createElement("span");
  absolute.className = "order-book__spread-value";
  absolute.textContent = formatPrice(spread.spread);

  const percent = document.createElement("span");
  percent.className = "order-book__spread-percent";
  percent.textContent = formatSpreadPercent(spread.percent);

  row.appendChild(label);
  row.appendChild(absolute);
  row.appendChild(percent);

  return row;
}

function centerSpreadRow(container) {
  const spreadRow = container.querySelector("[data-orderbook-spread-row]");
  if (!spreadRow) return;

  const maxScroll = container.scrollHeight - container.clientHeight;
  if (maxScroll <= 0) return;

  const targetScroll =
    spreadRow.offsetTop -
    container.clientHeight / 2 +
    spreadRow.clientHeight / 2;

  container.scrollTop = Math.min(Math.max(targetScroll, 0), maxScroll);
}

function scheduleSpreadCentering(container) {
  requestAnimationFrame(() => centerSpreadRow(container));
}

export function renderOrderBook({ state, elements, centerOnSpread = false }) {
  const orderBook = state.market?.order_book;
  if (!orderBook) return;

  const container = elements.orderBook.list;
  if (!container) return;

  container.innerHTML = "";

  const ladder = buildOrderBookLadder({
    orderBook,
    view: state.ui.orderBookView,
  });

  for (const level of ladder.asks) {
    container.appendChild(renderLevelRow(level));
  }

  container.appendChild(renderSpreadRow(ladder.spread));

  for (const level of ladder.bids) {
    container.appendChild(renderLevelRow(level));
  }

  if (centerOnSpread) {
    scheduleSpreadCentering(container);
  }
}

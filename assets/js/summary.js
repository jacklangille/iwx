export function renderSummary({ state, elements }) {
  const summary = state.market?.summary;

  const {
    bestAbove: bestAboveEl,
    bestBelow: bestBelowEl,
    askAbove: askAboveEl,
    askBelow: askBelowEl,
    midAbove: midAboveEl,
    midBelow: midBelowEl,
    liquidityAbove: liquidityAboveEl,
    liquidityBelow: liquidityBelowEl,
  } = elements.summary;

  const aboveLiquidity =
    summary?.liquidity?.above != null ? Number(summary.liquidity.above) : null;
  const belowLiquidity =
    summary?.liquidity?.below != null ? Number(summary.liquidity.below) : null;

  const midAbove =
    summary?.mid?.above != null ? Number(summary.mid.above) : null;
  const midBelow =
    summary?.mid?.below != null ? Number(summary.mid.below) : null;

  const bestAboveBid =
    summary?.best?.above?.bid != null ? Number(summary.best.above.bid) : null;
  const bestAboveAsk =
    summary?.best?.above?.ask != null ? Number(summary.best.above.ask) : null;

  const bestBelowBid =
    summary?.best?.below?.bid != null ? Number(summary.best.below.bid) : null;
  const bestBelowAsk =
    summary?.best?.below?.ask != null ? Number(summary.best.below.ask) : null;

  const bestAbove = bestAboveBid;
  const bestBelow = bestBelowBid;

  if (bestAboveEl) {
    bestAboveEl.textContent = bestAbove != null ? bestAbove.toFixed(2) : "-";
  }

  if (bestBelowEl) {
    bestBelowEl.textContent = bestBelow != null ? bestBelow.toFixed(2) : "-";
  }

  if (askAboveEl) {
    askAboveEl.textContent =
      bestAboveAsk != null ? bestAboveAsk.toFixed(2) : "-";
  }

  if (askBelowEl) {
    askBelowEl.textContent =
      bestBelowAsk != null ? bestBelowAsk.toFixed(2) : "-";
  }

  if (midAboveEl) {
    midAboveEl.textContent = midAbove != null ? midAbove.toFixed(2) : "-";
  }

  if (midBelowEl) {
    midBelowEl.textContent = midBelow != null ? midBelow.toFixed(2) : "-";
  }

  if (liquidityAboveEl) {
    liquidityAboveEl.textContent =
      aboveLiquidity != null ? String(aboveLiquidity) : "-";
  }

  if (liquidityBelowEl) {
    liquidityBelowEl.textContent =
      belowLiquidity != null ? String(belowLiquidity) : "-";
  }
}

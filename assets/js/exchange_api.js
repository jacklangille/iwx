export async function fetchMarketChart(contractId, config = {}) {
  const params = new URLSearchParams();

  if (config.lookbackSeconds) {
    params.set("lookback_seconds", String(config.lookbackSeconds));
  }

  if (config.bucketSeconds) {
    params.set("bucket_seconds", String(config.bucketSeconds));
  }

  const query = params.toString();
  const path = `/api/contracts/${contractId}/market_snapshots${query ? `?${query}` : ""}`;
  const res = await fetch(path);
  return res.json();
}

export async function fetchMarketState(contractId) {
  const res = await fetch(`/api/contracts/${contractId}/market_state`);
  return res.json();
}

export async function submitOrder({ contractId, tokenType, orderSide, price, quantity }) {
  const res = await fetch("/api/orders", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      contract_id: contractId,
      token_type: tokenType,
      order_side: orderSide,
      price,
      quantity,
    }),
  });

  return res.json();
}

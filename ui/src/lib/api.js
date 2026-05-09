const READ_API_BASE_URL =
  import.meta.env.VITE_READ_API_BASE_URL || "/read-api";
const AUTH_API_BASE_URL =
  import.meta.env.VITE_AUTH_API_BASE_URL || "/auth-api";
const EXCHANGE_CORE_API_BASE_URL =
  import.meta.env.VITE_EXCHANGE_CORE_API_BASE_URL || "/exchange-core-api";

async function requestJson(baseUrl, path, options = {}) {
  const headers = new Headers(options.headers || {});

  if (options.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${baseUrl}${path}`, {
    ...options,
    headers,
  });

  const contentType = response.headers.get("Content-Type") || "";
  const payload = contentType.includes("application/json")
    ? await response.json()
    : await response.text();

  if (!response.ok) {
    const error = new Error(
      typeof payload === "string"
        ? payload
        : payload?.error || "request failed",
    );
    error.status = response.status;
    error.payload = payload;
    throw error;
  }

  return payload;
}

function authHeaders(token) {
  return token ? { Authorization: `Bearer ${token}` } : {};
}

export function listContracts() {
  return requestJson(READ_API_BASE_URL, "/api/contracts");
}

export async function listStations(activeOnly = true) {
  const params = new URLSearchParams();
  if (activeOnly) {
    params.set("active", "true");
  }

  const response = await requestJson(
    READ_API_BASE_URL,
    `/api/stations${params.size ? `?${params.toString()}` : ""}`,
  );

  return response.stations || [];
}

export function getMarketState(contractId) {
  return requestJson(READ_API_BASE_URL, `/api/contracts/${contractId}/market_state`);
}

export function getMarketSnapshots(contractId, config) {
  const params = new URLSearchParams();

  if (config?.lookbackSeconds) {
    params.set("lookback_seconds", String(config.lookbackSeconds));
  }
  if (config?.bucketSeconds) {
    params.set("bucket_seconds", String(config.bucketSeconds));
  }

  const query = params.toString();
  return requestJson(
    READ_API_BASE_URL,
    `/api/contracts/${contractId}/market_snapshots${query ? `?${query}` : ""}`,
  );
}

export function getPortfolio(token) {
  return requestJson(READ_API_BASE_URL, "/api/me/portfolio", {
    headers: authHeaders(token),
  });
}

export function getUserSettlements(token) {
  return requestJson(READ_API_BASE_URL, "/api/me/settlements", {
    headers: authHeaders(token),
  });
}

export function depositFunds(token, body) {
  return requestJson(EXCHANGE_CORE_API_BASE_URL, "/api/accounts/deposits", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(body),
  });
}

export function submitOrder(token, body) {
  return requestJson(EXCHANGE_CORE_API_BASE_URL, "/api/orders", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(body),
  });
}

export function getOrderCommand(token, commandId) {
  return requestJson(EXCHANGE_CORE_API_BASE_URL, `/api/order_commands/${commandId}`, {
    headers: authHeaders(token),
  });
}

export function createContract(token, body) {
  return requestJson(EXCHANGE_CORE_API_BASE_URL, "/api/contracts", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(body),
  });
}

export function getContractCommand(token, commandId) {
  return requestJson(EXCHANGE_CORE_API_BASE_URL, `/api/contract_commands/${commandId}`, {
    headers: authHeaders(token),
  });
}

export function getContractDetails(token, contractId) {
  return requestJson(EXCHANGE_CORE_API_BASE_URL, `/api/contracts/${contractId}`, {
    headers: authHeaders(token),
  });
}

export function submitContractForApproval(token, contractId) {
  return requestJson(EXCHANGE_CORE_API_BASE_URL, `/api/contracts/${contractId}/submit_for_approval`, {
    method: "POST",
    headers: authHeaders(token),
  });
}

export function approveContract(token, contractId) {
  return requestJson(EXCHANGE_CORE_API_BASE_URL, `/api/contracts/${contractId}/approve`, {
    method: "POST",
    headers: authHeaders(token),
  });
}

export function getCollateralRequirement(token, contractId, pairedQuantity, currency = "USD") {
  const params = new URLSearchParams({
    paired_quantity: String(pairedQuantity),
    currency,
  });
  return requestJson(
    EXCHANGE_CORE_API_BASE_URL,
    `/api/contracts/${contractId}/collateral_requirement?${params.toString()}`,
    {
      headers: authHeaders(token),
    },
  );
}

export function lockContractCollateral(token, contractId, body) {
  return requestJson(EXCHANGE_CORE_API_BASE_URL, `/api/contracts/${contractId}/collateral_locks`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(body),
  });
}

export function listIssuanceBatches(token, contractId) {
  return requestJson(EXCHANGE_CORE_API_BASE_URL, `/api/contracts/${contractId}/issuance_batches`, {
    headers: authHeaders(token),
  });
}

export function createIssuanceBatch(token, contractId, body) {
  return requestJson(EXCHANGE_CORE_API_BASE_URL, `/api/contracts/${contractId}/issuance_batches`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(body),
  });
}

export function activateContract(token, contractId) {
  return requestJson(EXCHANGE_CORE_API_BASE_URL, `/api/contracts/${contractId}/activate`, {
    method: "POST",
    headers: authHeaders(token),
  });
}

export function login(username, password) {
  return requestJson(AUTH_API_BASE_URL, "/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  });
}

export function signup(username, password) {
  return requestJson(AUTH_API_BASE_URL, "/api/auth/signup", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  });
}

export function contractMarketStreamUrl(contractId) {
  return `${READ_API_BASE_URL}/api/stream/contracts/${contractId}/market`;
}

export { EXCHANGE_CORE_API_BASE_URL };

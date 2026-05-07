export async function fetchContracts() {
  const response = await fetch("/api/contracts");

  if (!response.ok) {
    throw new Error(`Failed to fetch contracts (${response.status})`);
  }

  const payload = await response.json();

  if (Array.isArray(payload)) return payload;
  if (Array.isArray(payload.data)) return payload.data;
  if (Array.isArray(payload.contracts)) return payload.contracts;

  return [];
}

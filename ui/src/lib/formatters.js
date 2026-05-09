export function formatDateRange(period) {
  if (!period?.start || !period?.end) return "-";

  const start = new Date(`${period.start}T00:00:00`);
  const end = new Date(`${period.end}T00:00:00`);

  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return "-";

  const startText = start.toLocaleDateString([], {
    month: "short",
    day: "numeric",
  });
  const endText = end.toLocaleDateString([], {
    month: "short",
    day: "numeric",
    year: "numeric",
  });

  return `${startText} \u2013 ${endText}`;
}

export function formatMeasurementUnit(unit) {
  if (unit === "deg_f") return "\u00B0F";
  if (unit === "deg_c") return "\u00B0C";
  return "";
}

function titleCase(value) {
  return String(value || "")
    .split(/[_\s]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function displayUnit(unit) {
  if (unit === "deg_f") return "\u00B0F";
  if (unit === "deg_c") return "\u00B0C";
  return unit || "";
}

export function buildContractLabel(contract) {
  if (!contract) return "";

  const measurementRange = formatDateRange({
    start: contract.measurement_period_start,
    end: contract.measurement_period_end,
  });

  return [
    contract.region,
    measurementRange !== "-" ? measurementRange : null,
    contract.metric ? `${contract.metric} Index` : null,
  ]
    .filter(Boolean)
    .join(" ");
}

export function buildContractDescription(contract) {
  if (!contract) return "No contract description available.";
  if (contract.description) return contract.description;

  const metric = titleCase(contract.metric || "metric");
  const threshold = contract.threshold != null ? String(contract.threshold) : "the configured threshold";
  const unit = displayUnit(contract.measurement_unit);
  const payoutCents = Number(contract.multiplier || 100);
  const payoutText = Number.isFinite(payoutCents) ? `$${(payoutCents / 100).toFixed(2)}` : "the configured payout";
  const thresholdText = unit ? `${threshold}${unit}` : threshold;

  return `If ${metric} resolves above ${thresholdText}, ABOVE pays ${payoutText} per claim. If it resolves at or below ${thresholdText}, BELOW pays ${payoutText} per claim.`;
}

export function contractStatusClassName(status) {
  const value = String(status || "").toLowerCase();

  if (value === "open" || value === "active") return "market-status market-status--open";
  if (value === "resolved" || value === "awaiting_resolution" || value === "trading_closed") {
    return "market-status market-status--resolving";
  }

  return "market-status market-status--closed";
}

export function formatPrice(value) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed.toFixed(2) : "-";
}

export function formatQuantity(value) {
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) return "-";
  return Number.isInteger(parsed) ? String(parsed) : parsed.toFixed(2);
}

export function formatMoneyCents(value) {
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) return "-";

  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(parsed / 100);
}

export function formatTimestamp(value) {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";

  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(date);
}

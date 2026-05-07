function buildContractLabel(contract) {
  const measurementRange = formatDateRange(contract.measurement_period);

  const parts = [
    contract.region_label,
    measurementRange !== "-" ? measurementRange : null,
    contract.metric ? `${contract.metric} Index` : null,
  ].filter(Boolean);

  return parts.join(" ");
}

function formatMeasurementUnit(unit) {
  if (unit === "deg_f") return "°F";
  if (unit === "deg_c") return "°C";
  return "";
}

function formatDateRange(period) {
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

  return `${startText} – ${endText}`;
}

function buildContractDescription(contract) {
  if (contract.description) return contract.description;

  return "No contract description available.";
}

function renderStatusBadge(contract, statusBadge) {
  if (!statusBadge) return;

  const status = (contract.status ?? "").toLowerCase();

  statusBadge.classList.remove(
    "market-status--open",
    "market-status--resolving",
    "market-status--closed",
  );

  if (status === "open") {
    statusBadge.textContent = "Open";
    statusBadge.classList.add("market-status--open");
  } else if (status === "resolving") {
    statusBadge.textContent = "Resolving";
    statusBadge.classList.add("market-status--resolving");
  } else {
    statusBadge.textContent = contract.status ?? "Closed";
    statusBadge.classList.add("market-status--closed");
  }
}

export function renderCurrentContract({ state, elements }) {
  if (!state.contract || !elements.contract.label) return;

  elements.contract.label.textContent = buildContractLabel(state.contract);
}

export function renderContractInfo({ state, elements }) {
  const contract = state.contract;
  if (!contract) return;

  if (elements.contract.description) {
    elements.contract.description.textContent = buildContractDescription(contract);
  }

  if (elements.contract.dataProvider) {
    elements.contract.dataProvider.textContent =
      contract.data_provider?.name && contract.data_provider?.station_mode
        ? `${contract.data_provider.name} ${contract.data_provider.station_mode === "single_station" ? "Single Station" : contract.data_provider.station_mode}`
        : "-";
  }

  if (elements.contract.threshold) {
    const unit = formatMeasurementUnit(contract.measurement_unit);
    elements.contract.threshold.textContent =
      contract.threshold != null
        ? `Threshold ${Number(contract.threshold)}${unit}`
        : "Threshold -";
  }

  renderStatusBadge(contract, elements.contract.statusBadge);

  if (elements.contract.tradingPeriod) {
    elements.contract.tradingPeriod.textContent =
      `Trading ${formatDateRange(contract.trading_period)}`;
  }

  if (elements.contract.measurementPeriod) {
    elements.contract.measurementPeriod.textContent =
      `Measurement ${formatDateRange(contract.measurement_period)}`;
  }
}
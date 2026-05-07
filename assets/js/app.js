import { renderChart } from "./chart";
import { renderContractInfo, renderCurrentContract } from "./contract_info";
import { fetchContracts } from "./contracts";
import { fetchMarketChart, fetchMarketState, submitOrder } from "./exchange_api";
import { connectMarketChannel } from "./market_channel";
import { renderOrderBook, renderOrderBookToggle } from "./orderbook";
import { initRegionMap, syncRegionMap } from "./region_map";
import { renderTradePanel } from "./trade_panel";

/*--------------STATE--------------*/
const state = {
  contract: null,
  market: null,
  chartNow: new Date(),
  ui: {
    belief: "above",
    action: "bid",
    mode: "trade",
    orderBookView: "above",
    chartLookbackSeconds: 24 * 60 * 60,
    chartBucketSeconds: 5 * 60,
  },
  contracts: [],
};

function buildContractState(contract) {
  if (!contract) return null;

  return {
    id: contract.id,
    name: contract.name,
    region_label: contract.region,
    metric: contract.metric,
    measurement_unit: contract.measurement_unit,
    status: contract.status,
    threshold: contract.threshold,
    multiplier: contract.multiplier,
    trading_period: {
      start: contract.trading_period_start,
      end: contract.trading_period_end,
    },
    measurement_period: {
      start: contract.measurement_period_start,
      end: contract.measurement_period_end,
    },
    data_provider: {
      name: contract.data_provider_name,
      station_mode: contract.data_provider_station_mode,
    },
    description: contract.description,
  };
}

function buildMarketState({ marketStatePayload, chartData }) {
  const chart = {
    config: chartData?.config ?? null,
    points: chartData?.points ?? [],
  };

  return {
    ...(marketStatePayload ?? {}),
    chart,
  };
}

function parseBucketTimestamp(bucketStart) {
  if (!bucketStart) return NaN;

  const normalized = bucketStart.endsWith("Z") ? bucketStart : `${bucketStart}Z`;
  return Date.parse(normalized);
}

export function patchChartSeries(existingPoints, incomingChartPoint, chartConfig) {
  if (!incomingChartPoint?.bucket_start) return existingPoints ?? [];

  const pointsByBucket = new Map(
    (existingPoints ?? [])
      .filter((point) => point?.bucket_start)
      .map((point) => [point.bucket_start, point]),
  );

  pointsByBucket.set(incomingChartPoint.bucket_start, incomingChartPoint);

  const sortedPoints = Array.from(pointsByBucket.values()).sort((left, right) =>
    left.bucket_start.localeCompare(right.bucket_start),
  );

  const lookbackSeconds = Number(chartConfig?.lookback_seconds);
  if (!Number.isFinite(lookbackSeconds) || lookbackSeconds <= 0) {
    return sortedPoints;
  }

  const latestTimestamp = parseBucketTimestamp(
    sortedPoints[sortedPoints.length - 1]?.bucket_start,
  );
  if (!Number.isFinite(latestTimestamp)) return sortedPoints;

  const windowStart = latestTimestamp - lookbackSeconds * 1000;

  return sortedPoints.filter((point) => {
    const bucketTimestamp = parseBucketTimestamp(point.bucket_start);
    return Number.isFinite(bucketTimestamp) && bucketTimestamp >= windowStart;
  });
}

function patchMarketChart(chart, incomingChartPoint) {
  if (!chart || !incomingChartPoint) return chart;

  return {
    ...chart,
    points: patchChartSeries(chart.points, incomingChartPoint, chart.config),
  };
}

function chartRequestConfig() {
  return {
    lookbackSeconds: state.ui.chartLookbackSeconds,
    bucketSeconds: state.ui.chartBucketSeconds,
  };
}

/*--------------UTIL--------------*/
function findContractById(contracts, contractId) {
  return contracts.find(
    (contract) => String(contract.id) === String(contractId),
  );
}

function setContract(contract) {
  state.contract = buildContractState(contract);
}

function startChartViewportClock(elements, interval = 5000) {
  return setInterval(() => {
    state.chartNow = new Date();
    renderChartFromState(elements);
  }, interval);
}


function getElements() {
  const actionSelector = document.getElementById("action-selector");

  return {
    chart: {
      canvas: document.getElementById("chart-canvas"),
      shell: document.querySelector(".chart-shell"),
    },
    regionMap: document.getElementById("region-map"),
    contract: {
      label: document.getElementById("contract-title"),
      statusBadge: document.getElementById("market-status-badge"),
      tradingPeriod: document.getElementById("market-trading-period"),
      measurementPeriod: document.getElementById("market-measurement-period"),
      description: document.getElementById("contract-description"),
      threshold: document.getElementById("contract-threshold"),
      dataProvider: document.getElementById("contract-data-provider"),
    },
    orderBook: {
      list: document.getElementById("order-book-list"),
      viewSelect: document.getElementById("order-book-view"),
    },
    trade: {
      actionSelector,
      actionOptions:
        actionSelector?.querySelectorAll(".action-selector__option") ?? [],
      beliefSelect: document.getElementById("order-belief"),
      submitButton: document.getElementById("order-submit"),
      priceInput: document.getElementById("order-price"),
      quantityInput: document.getElementById("order-quantity"),
      previewSide: document.getElementById("order-preview-side"),
      previewQuote: document.getElementById("order-preview-quote"),
      previewPrice: document.getElementById("order-preview-price"),
      previewQuantity: document.getElementById("order-preview-quantity"),
    },
  };
}

/*--------------API--------------*/
async function fetchChartData(contractId) {
  return fetchMarketChart(contractId, chartRequestConfig());
}

async function refreshChartData(elements) {
  if (!state.contract) return;

  const chartData = await fetchChartData(state.contract.id);

  state.market = {
    ...(state.market ?? {}),
    chart: {
      config: chartData?.config ?? null,
      points: chartData?.points ?? [],
    },
  };

  renderChartFromState(elements);
}

async function refreshMarketState(elements) {
  if (!state.contract) return;

  const currentContractId = state.contract.id;
  const [marketStatePayload, chartData] = await Promise.all([
    fetchMarketState(currentContractId),
    fetchChartData(currentContractId),
  ]);

  state.market = buildMarketState({
    marketStatePayload,
    chartData,
  });

  renderMarketState(elements);
  renderChartFromState(elements);
}

async function refreshContractState(elements, contractId) {
  state.contracts = await fetchContracts();

  const contractPayload =
    findContractById(state.contracts, contractId) ?? state.contracts[0] ?? null;

  setContract(contractPayload);

  if (!state.contract) return;

  const [marketStatePayload, chartData] = await Promise.all([
    fetchMarketState(state.contract.id),
    fetchChartData(state.contract.id),
  ]);

  state.market = buildMarketState({
    marketStatePayload,
    chartData,
  });

  renderFullState(elements);
}

/*--------------RENDER--------------*/
function renderFullState(elements) {
  renderOrderBookToggle({ state, elements });
  renderOrderBook({ state, elements, centerOnSpread: true });
  renderCurrentContract({ state, elements });
  renderContractInfo({ state, elements });
  renderTradePanel({ state, elements });
  renderChartFromState(elements);
  syncRegionMap(elements.regionMap, state.contract);
}

function renderMarketState(elements) {
  renderOrderBook({ state, elements });
  renderTradePanel({ state, elements });
}

function renderChartFromState(elements) {
  renderChart(
    elements.chart.canvas,
    state.contract,
    state.market?.chart ?? null,
    state.chartNow
  );
}

/*--------------BINDS--------------*/
function bindOrderBook(elements) {
  const { viewSelect } = elements.orderBook;

  if (viewSelect) {
    viewSelect.addEventListener("change", () => {
      const nextView = viewSelect.value;
      if (nextView === state.ui.orderBookView) return;
      if (!["above", "below"].includes(nextView)) return;

      state.ui.orderBookView = nextView;
      renderOrderBookToggle({ state, elements });
      renderOrderBook({ state, elements, centerOnSpread: true });
    });
  }
}
function bindTradePanel(elements) {
  const {
    actionSelector,
    beliefSelect,
    priceInput,
    quantityInput,
    submitButton,
  } = elements.trade;

  if (actionSelector) {
    actionSelector.addEventListener("click", (event) => {
      const option = event.target.closest(".action-selector__option");
      if (!option) return;

      const nextAction = option.dataset.action;
      if (nextAction === state.ui.action) return;

      state.ui.action = nextAction;
      renderTradePanel({ state, elements });
    });
  }

  if (beliefSelect) {
    beliefSelect.addEventListener("change", () => {
      const nextBelief = beliefSelect.value;
      if (nextBelief === state.ui.belief) return;

      state.ui.belief = nextBelief;
      renderTradePanel({ state, elements });
    });
  }

  if (priceInput) {
    priceInput.addEventListener("input", () => {
      renderTradePanel({ state, elements });
    });
  }

  if (quantityInput) {
    quantityInput.addEventListener("input", () => {
      renderTradePanel({ state, elements });
    });
  }

  if (submitButton && priceInput && quantityInput && beliefSelect) {
    submitButton.addEventListener("click", async () => {
      if (!state.contract) return;

      const price = Number(priceInput.value);
      const quantity = Number(quantityInput.value);

      if (!price || !quantity) {
        console.error("Missing price or quantity");
        return;
      }

      try {
        await submitOrder({
          contractId: state.contract.id,
          tokenType: beliefSelect.value,
          orderSide: state.ui.action,
          price,
          quantity,
        });
      } catch (err) {
        console.error("Order failed", err);
      }
    });
  }
}

function bindEvents(elements) {
  bindOrderBook(elements);
  bindTradePanel(elements);
  bindChartControls(elements);
}

function bindChartControls(elements) {
  const shell = elements.chart.shell;
  if (!shell) return;

  shell.addEventListener("chart-config-change", async (event) => {
    const { lookbackSeconds, bucketSeconds } = event.detail ?? {};
    const nextLookbackSeconds = Number(lookbackSeconds);
    const nextBucketSeconds = Number(bucketSeconds);

    if (
      !Number.isFinite(nextLookbackSeconds) ||
      !Number.isFinite(nextBucketSeconds) ||
      (nextLookbackSeconds === state.ui.chartLookbackSeconds &&
        nextBucketSeconds === state.ui.chartBucketSeconds)
    ) {
      return;
    }

    state.ui.chartLookbackSeconds = nextLookbackSeconds;
    state.ui.chartBucketSeconds = nextBucketSeconds;

    try {
      await refreshChartData(elements);
    } catch (error) {
      console.error("Failed to refresh chart", error);
    }
  });
}

document.addEventListener("DOMContentLoaded", async () => {
  const elements = getElements();

  try {
    state.contracts = await fetchContracts();
  } catch (error) {
    console.error("Failed to load contracts", error);
    return;
  }

  if (state.contracts.length === 0) return;
  setContract(state.contracts[0]);
  initRegionMap(elements.regionMap, state.contract);
  await refreshContractState(elements, state.contract.id);
  
  if (elements.trade.beliefSelect) {
    state.ui.belief = elements.trade.beliefSelect.value;
  }

  connectMarketChannel({
    contractId: state.contract.id,
    getCurrentMarket: () => state.market,
    onMarketState: (payload) => {
      state.market = buildMarketState({
        marketStatePayload: payload.market_state,
        chartData: state.market?.chart,
      });

      if (state.ui.chartBucketSeconds === 5) {
        state.market.chart = patchMarketChart(
          state.market.chart,
          payload.chart_point,
        );
      } else {
        refreshChartData(elements).catch((error) => {
          console.error("Failed to refresh chart", error);
        });
      }

      renderMarketState(elements);
      renderChartFromState(elements);
    },
    onReconnect: async () => {
      await refreshMarketState(elements);
    },
  });

  bindEvents(elements);
  startChartViewportClock(elements);
});

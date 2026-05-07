import Chart from "chart.js/auto";

let chartInstance = null;
const MISSING_VALUE = "\u2014";
const DAY_SECONDS = 24 * 60 * 60;
const MONTH_SECONDS = 30 * DAY_SECONDS;
const LOOKBACK_PRESETS = [
  { label: "1month", seconds: MONTH_SECONDS },
  { label: "5day", seconds: 5 * DAY_SECONDS },
  { label: "1day", seconds: DAY_SECONDS }
];
const RESOLUTION_PRESETS = [
  { label: "5min", seconds: 5 * 60 },
  { label: "1hour", seconds: 60 * 60 },
  { label: "1day", seconds: DAY_SECONDS }
];
const DEFAULT_LOOKBACK_SECONDS = DAY_SECONDS;
const DEFAULT_BUCKET_SECONDS = 5 * 60;

function cssToken(styles, name, fallback = "") {
  return styles.getPropertyValue(name).trim() || fallback;
}

function cssNumberToken(styles, name, fallback) {
  const value = Number.parseFloat(cssToken(styles, name));
  return Number.isFinite(value) ? value : fallback;
}

function latestDatasetPoint(dataset) {
  const data = dataset?.data ?? [];

  for (let index = data.length - 1; index >= 0; index -= 1) {
    const point = data[index];
    if (point?.y != null && !Number.isNaN(point.y)) return point;
  }

  return null;
}

function withAlpha(color, alpha) {
  if (!color) return color;

  if (color.startsWith("#")) {
    let hex = color.slice(1);

    if (hex.length === 3) {
      hex = hex.split("").map((char) => char + char).join("");
    }

    const int = Number.parseInt(hex, 16);
    const r = (int >> 16) & 255;
    const g = (int >> 8) & 255;
    const b = int & 255;

    return `rgba(${r}, ${g}, ${b}, ${alpha})`;
  }

  if (color.startsWith("rgb(")) {
    return color.replace("rgb(", "rgba(").replace(")", `, ${alpha})`);
  }

  return color;
}

function splitHistoricalAndLiveSeries(dataset) {
  const data = dataset ?? [];
  if (data.length <= 1) return { historical: data, liveTail: [] };

  return {
    historical: data.slice(0, -1),
    liveTail: data.slice(-2),
  };
}

const hoverGuidePlugin = {
  id: "hoverGuide",
  afterDatasetsDraw(chart) {
    drawLiveAnchors(chart);

    const index = chart.$hoverIndex;
    if (index == null) return;

    const xScale = chart.scales.x;
    const { left, right, top, bottom } = chart.chartArea;

    const xValue = chart.$xValues?.[index];
    const x = xScale.getPixelForValue(xValue);
    const y = chart.$hoverYPixel;

    if (!Number.isFinite(x)) return;

    const { ctx } = chart;
    const guideColor = chart.$theme?.chartGuide ?? getChartTheme().chartGuide;

    ctx.save();
    ctx.beginPath();
    ctx.setLineDash([4, 5]);
    ctx.lineWidth = 1;
    ctx.strokeStyle = guideColor;
    ctx.moveTo(x, top);
    ctx.lineTo(x, bottom);
    ctx.stroke();
    ctx.restore();

    if (Number.isFinite(y) && y >= top && y <= bottom) {
      ctx.save();
      ctx.beginPath();
      ctx.setLineDash([2, 4]);
      ctx.lineWidth = 1;
      ctx.strokeStyle = guideColor;
      ctx.globalAlpha = 0.78;
      ctx.moveTo(left, y);
      ctx.lineTo(right, y);
      ctx.stroke();
      ctx.restore();
    }
  },
  afterEvent(chart, args) {
    const event = args.event;
    const area = chart.chartArea;
    const isHoverEvent = event.type === "mousemove" || event.type === "mouseout";

    if (!isHoverEvent) return;

    const nextIndex =
      event.type === "mousemove" && args.inChartArea
        ? getNearestDataIndex(chart, event.x, area)
        : null;
    const nextYPixel =
      event.type === "mousemove" && args.inChartArea
        ? Math.min(Math.max(event.y, area.top), area.bottom)
        : null;

    if (chart.$hoverIndex === nextIndex && chart.$hoverYPixel === nextYPixel) return;

    chart.$hoverIndex = nextIndex;
    chart.$hoverYPixel = nextYPixel;
    updateChartChrome(chart);
    args.changed = true;
  }
};

Chart.register(hoverGuidePlugin);

function parseSnapshotTimestamp(timestamp) {
  if (!timestamp) return null;
  const normalized = timestamp.endsWith("Z") ? timestamp : `${timestamp}Z`;
  const date = new Date(normalized);

  return Number.isNaN(date.getTime()) ? null : date;
}

function getChartTheme() {
  const styles = getComputedStyle(document.documentElement);

  return {
    chartAbove: cssToken(styles, "--color-chart-above"),
    chartBelow: cssToken(styles, "--color-chart-below"),
    chartGrid: cssToken(styles, "--color-chart-grid"),
    chartGuide: cssToken(styles, "--color-chart-guide"),
    chartAxisBorder: cssToken(styles, "--color-border-chart-axis"),
    chartAnchorStroke: cssToken(styles, "--color-chart-anchor-stroke"),
    chartMarkerText: cssToken(styles, "--color-chart-marker-text"),
    chartTransparent: cssToken(styles, "--color-surface-transparent", "transparent"),
    textColor: cssToken(styles, "--color-text-primary"),
    textMuted: cssToken(styles, "--color-text-muted"),
    fontFamilySans: cssToken(styles, "--font-family-sans"),
    fontSizeCaption: cssNumberToken(styles, "--font-size-caption", 11),
    fontWeightNormal: cssNumberToken(styles, "--font-weight-normal", 400),
    minWidthChartMarker: cssNumberToken(styles, "--min-width-chart-marker", 44),
    heightChartMarker: cssNumberToken(styles, "--height-chart-marker", 20),
    radiusChartControl: cssNumberToken(styles, "--radius-sm", 4),
    spaceSm: cssNumberToken(styles, "--space-sm", 4),
    spaceMd: cssNumberToken(styles, "--space-md", 8),
    spaceLg: cssNumberToken(styles, "--space-lg", 16)
  };
}

function formatTickLabel(date, lookbackSeconds) {
  if (!date) return "-";

  if (lookbackSeconds === DAY_SECONDS) {
    return new Intl.DateTimeFormat("en-US", {
      hour: "numeric",
      hour12: true
    }).format(date);
  }

  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric"
  }).format(date);
}

function xTickCount(lookbackSeconds) {
  if (lookbackSeconds === MONTH_SECONDS) return 10;
  if (lookbackSeconds === 5 * DAY_SECONDS) return 5;
  if (lookbackSeconds === DAY_SECONDS) return 12;
  return 6;
}

function buildDeterministicXTicks(min, max, lookbackSeconds) {
  if (!Number.isFinite(min) || !Number.isFinite(max) || min >= max) return [];

  const tickCount = xTickCount(lookbackSeconds);
  const interval = (max - min) / (tickCount - 1);

  return Array.from({ length: tickCount }, (_value, index) => ({
    value: index === tickCount - 1 ? max : min + interval * index
  }));
}

function formatPrice(value) {
  return value == null || Number.isNaN(value) ? MISSING_VALUE : value.toFixed(2);
}

function formatAbsoluteChange(currentValue, baselineValue) {
  if (
    currentValue == null ||
    Number.isNaN(currentValue) ||
    baselineValue == null ||
    Number.isNaN(baselineValue)
  ) {
    return { text: MISSING_VALUE, tone: "neutral" };
  }

  const change = currentValue - baselineValue;
  const prefix = change > 0 ? "+" : "";
  const tone = change > 0 ? "positive" : change < 0 ? "negative" : "neutral";

  return { text: `${prefix}${change.toFixed(2)}`, tone };
}

function firstNonNullValue(values) {
  return values.find((value) => value != null && !Number.isNaN(value)) ?? null;
}

function formatPercentChange(currentValue, baselineValue) {
  if (
    currentValue == null ||
    Number.isNaN(currentValue) ||
    baselineValue == null ||
    Number.isNaN(baselineValue) ||
    baselineValue === 0
  ) {
    return { text: MISSING_VALUE, tone: "neutral" };
  }

  const percentChange = ((currentValue - baselineValue) / baselineValue) * 100;

  if (!Number.isFinite(percentChange)) {
    return { text: MISSING_VALUE, tone: "neutral" };
  }

  const prefix = percentChange > 0 ? "+" : "";
  const tone = percentChange > 0 ? "positive" : percentChange < 0 ? "negative" : "neutral";

  return { text: `${prefix}${percentChange.toFixed(1)}%`, tone };
}

function formatChangeParts(currentValue, baselineValue) {
  const absolute = formatAbsoluteChange(currentValue, baselineValue);
  const percent = formatPercentChange(currentValue, baselineValue);
  const tone = absolute.tone !== "neutral" ? absolute.tone : percent.tone;

  if (absolute.text === MISSING_VALUE && percent.text === MISSING_VALUE) {
    return { text: MISSING_VALUE, tone: "neutral" };
  }

  return {
    text: `${absolute.text} / ${percent.text}`,
    tone
  };
}

function formatDateTime(date) {
  if (!date) return MISSING_VALUE;

  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    second: "2-digit",
    hour12: true
  }).format(date);
}

function formatTimezone(date = new Date()) {
  const offsetMinutes = -date.getTimezoneOffset();
  const sign = offsetMinutes >= 0 ? "+" : "-";
  const absMinutes = Math.abs(offsetMinutes);
  const hours = Math.floor(absMinutes / 60);
  const minutes = absMinutes % 60;

  return minutes === 0
    ? `UTC${sign}${hours}`
    : `UTC${sign}${hours}:${String(minutes).padStart(2, "0")}`;
}

function formatDuration(seconds) {
  if (seconds >= 30 * 24 * 3600 && seconds % (30 * 24 * 3600) === 0) {
    return `${seconds / (30 * 24 * 3600)}mo`;
  }
  if (seconds >= 24 * 3600 && seconds % (24 * 3600) === 0) {
    return `${seconds / (24 * 3600)}d`;
  }
  if (seconds >= 3600) return `${seconds / 3600}h`;
  if (seconds >= 60) return `${seconds / 60}m`;
  return `${seconds}s`;
}

function chartLayoutPadding() {
  const width = window.innerWidth;

  return {
    top: width <= 560 ? 86 : 54,
    right: 58,
    bottom: width <= 560 ? 54 : 30,
    left: 4
  };
}

function getNearestDataIndex(chart, eventX, area) {
  const xValues = chart.$xValues ?? [];
  if (xValues.length === 0) return null;

  const boundedX = Math.min(Math.max(eventX, area.left), area.right);
  const hoveredX = chart.scales.x.getValueForPixel(boundedX);

  let nearestIndex = 0;
  let nearestDistance = Math.abs(xValues[0] - hoveredX);

  for (let index = 1; index < xValues.length; index += 1) {
    const distance = Math.abs(xValues[index] - hoveredX);
    if (distance < nearestDistance) {
      nearestDistance = distance;
      nearestIndex = index;
    }
  }

  return nearestIndex;
}

function drawLiveAnchors(chart) {
  const yScale = chart.scales.y;
  const area = chart.chartArea;
  if (!yScale || !area) return;

  const theme = chart.$theme ?? getChartTheme();
  const anchors = [
    {
      label: "Above",
      value:
        latestDatasetPoint(chart.data.datasets[1])?.y ??
        latestDatasetPoint(chart.data.datasets[0])?.y,
      color: theme.chartAbove
    },
    {
      label: "Below",
      value:
        latestDatasetPoint(chart.data.datasets[3])?.y ??
        latestDatasetPoint(chart.data.datasets[2])?.y,
      color: theme.chartBelow
    }
  ].filter((anchor) => anchor.value != null && !Number.isNaN(anchor.value));

  const { ctx } = chart;

  anchors.forEach((anchor) => {
    const y = yScale.getPixelForValue(anchor.value);
    if (!Number.isFinite(y) || y < area.top || y > area.bottom) return;

    ctx.save();
    ctx.beginPath();
    ctx.setLineDash([3, 4]);
    ctx.lineWidth = 1;
    ctx.strokeStyle = anchor.color;
    ctx.globalAlpha = 0.18;
    ctx.moveTo(area.left, y);
    ctx.lineTo(area.right, y);
    ctx.stroke();
    ctx.restore();

    const text = formatPrice(anchor.value);
    ctx.save();
    ctx.font = `${theme.fontSizeCaption}px ${theme.fontFamilySans}`;
    const textWidth = ctx.measureText(text).width;
    const markerWidth = Math.max(theme.minWidthChartMarker, textWidth + theme.spaceLg);
    const markerHeight = theme.heightChartMarker;
    const x = area.right + theme.spaceSm;
    const markerY = Math.min(
      Math.max(y - markerHeight / 2, area.top + 2),
      area.bottom - markerHeight - 2
    );

    ctx.fillStyle = anchor.color;
    ctx.strokeStyle = theme.chartAnchorStroke;
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.roundRect(x, markerY, markerWidth, markerHeight, theme.radiusChartControl);
    ctx.fill();
    ctx.stroke();
    ctx.fillStyle = theme.chartMarkerText;
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText(text, x + markerWidth / 2, markerY + markerHeight / 2 + 0.5);
    ctx.restore();
  });
}

function ensureChartChrome(canvas, theme, config) {
  const shell = canvas.parentElement;
  if (!shell) return;

  shell.classList.add("chart-shell--enhanced");
  shell.dataset.chartLookbackSeconds = String(
    Number(config?.lookback_seconds) || DEFAULT_LOOKBACK_SECONDS
  );
  shell.dataset.chartBucketSeconds = String(Number(config?.bucket_seconds) || DEFAULT_BUCKET_SECONDS);

  shell.querySelector("[data-chart-hover-panel]")?.remove();
  shell.querySelector("[data-chart-legend]")?.remove();

  let topChrome = shell.querySelector("[data-chart-top-chrome]");
  if (!topChrome) {
    topChrome = document.createElement("div");
    topChrome.className = "chart-top-chrome";
    topChrome.dataset.chartTopChrome = "";
    topChrome.innerHTML = `
      <div class="chart-market-readout" aria-live="polite">
        <div class="chart-market-readout__row" data-chart-readout-above>
          <span class="chart-market-readout__series chart-market-readout__series--above">Above</span>
          <span class="chart-market-readout__value" data-chart-readout-above-value></span>
          <span class="chart-market-readout__change" data-chart-readout-above-change></span>
        </div>
        <div class="chart-market-readout__row" data-chart-readout-below>
          <span class="chart-market-readout__series chart-market-readout__series--below">Below</span>
          <span class="chart-market-readout__value" data-chart-readout-below-value></span>
          <span class="chart-market-readout__change" data-chart-readout-below-change></span>
        </div>
      </div>
      <div class="chart-controls" aria-label="Chart controls">
        <div class="chart-control-group" aria-label="Lookback window" data-chart-lookbacks></div>
        <div class="chart-control-group" aria-label="Resolution" data-chart-resolutions></div>
      </div>
    `;
    shell.appendChild(topChrome);
  }

  let bottomStrip = shell.querySelector("[data-chart-status-strip]");
  if (!bottomStrip) {
    bottomStrip = document.createElement("div");
    bottomStrip.className = "chart-status-strip";
    bottomStrip.dataset.chartStatusStrip = "";
    bottomStrip.innerHTML = `
      <div class="chart-status-strip__time" data-chart-status-time></div>
      <div class="chart-status-strip__meta">
        <span data-chart-status-timezone></span>
        <span data-chart-status-context></span>
      </div>
    `;
    shell.appendChild(bottomStrip);
  }

  renderControlGroup(
    topChrome.querySelector("[data-chart-lookbacks]"),
    LOOKBACK_PRESETS,
    Number(config?.lookback_seconds) || DEFAULT_LOOKBACK_SECONDS,
    "lookback"
  );
  renderControlGroup(
    topChrome.querySelector("[data-chart-resolutions]"),
    RESOLUTION_PRESETS,
    Number(config?.bucket_seconds) || DEFAULT_BUCKET_SECONDS,
    "resolution"
  );

  if (!shell.dataset.chartControlsBound) {
    shell.addEventListener("click", (event) => {
      const button = event.target.closest("[data-chart-control]");
      if (!button) return;

      const nextLookback = Number(
        button.dataset.lookbackSeconds ?? shell.dataset.chartLookbackSeconds ?? DEFAULT_LOOKBACK_SECONDS
      );
      const nextResolution = Number(
        button.dataset.bucketSeconds ?? shell.dataset.chartBucketSeconds ?? DEFAULT_BUCKET_SECONDS
      );

      shell.dispatchEvent(
        new CustomEvent("chart-config-change", {
          bubbles: true,
          detail: {
            lookbackSeconds: nextLookback,
            bucketSeconds: nextResolution
          }
        })
      );
    });
    shell.dataset.chartControlsBound = "true";
  }

  shell.style.setProperty("--color-chart-above-active", theme.chartAbove);
  shell.style.setProperty("--color-chart-below-active", theme.chartBelow);
}

function renderControlGroup(container, presets, activeSeconds, type) {
  if (!container) return;

  container.innerHTML = presets
    .map((preset) => {
      const activeClass = preset.seconds === activeSeconds ? " chart-control-button--active" : "";
      const dataAttribute =
        type === "lookback"
          ? `data-lookback-seconds="${preset.seconds}"`
          : `data-bucket-seconds="${preset.seconds}"`;

      return `
        <button
          class="chart-control-button${activeClass}"
          type="button"
          data-chart-control
          ${dataAttribute}
        >
          ${preset.label}
        </button>
      `;
    })
    .join("");
}

function setReadoutMetric(shell, key, value, baseline) {
  const valueElement = shell.querySelector(`[data-chart-readout-${key}-value]`);
  const changeElement = shell.querySelector(`[data-chart-readout-${key}-change]`);
  const change = formatChangeParts(value, baseline);

  if (!valueElement || !changeElement) return;

  valueElement.textContent = formatPrice(value);
  changeElement.textContent = change.text;
  changeElement.className = `chart-market-readout__change chart-market-readout__change--${change.tone}`;
}

function updateChartChrome(chart) {
  const shell = chart.canvas.parentElement;
  if (!shell) return;

  const index = chart.$hoverIndex;
  const isHovering = index != null && index < (chart.$xValues?.length ?? 0);

  const aboveValue = isHovering
    ? chart.$projectedPoints?.[index]?.mid_above != null
      ? Number(chart.$projectedPoints[index].mid_above)
      : null
    : latestDatasetPoint(chart.data.datasets[1])?.y ?? latestDatasetPoint(chart.data.datasets[0])?.y;

  const belowValue = isHovering
    ? chart.$projectedPoints?.[index]?.mid_below != null
      ? Number(chart.$projectedPoints[index].mid_below)
      : null
    : latestDatasetPoint(chart.data.datasets[3])?.y ?? latestDatasetPoint(chart.data.datasets[2])?.y;

  const statusDate = isHovering ? chart.$dates?.[index] : chart.$liveDate;
  const config = parseChartConfig(chart.$config);

  setReadoutMetric(shell, "above", aboveValue, chart.$baselines?.above);
  setReadoutMetric(shell, "below", belowValue, chart.$baselines?.below);
  shell.querySelector("[data-chart-status-time]").textContent = formatDateTime(statusDate);
  shell.querySelector("[data-chart-status-timezone]").textContent = formatTimezone(statusDate);
  shell.querySelector("[data-chart-status-context]").textContent =
    `${formatDuration(config.bucketSeconds)} resolution / ${formatDuration(config.lookbackSeconds)} window`;
}

function buildChartData(points) {
  const dates = points.map((point) =>
    parseSnapshotTimestamp(point.bucket_start ?? point.inserted_at),
  );

  const xValues = dates.map((date) => date?.getTime() ?? NaN);

  const aboveValues = points.map((point) =>
    point.mid_above != null ? Number(point.mid_above) : null
  );
  const belowValues = points.map((point) =>
    point.mid_below != null ? Number(point.mid_below) : null
  );

  return {
    xValues,
    above: xValues.map((x, index) => ({
      x,
      y: aboveValues[index],
    })),
    below: xValues.map((x, index) => ({
      x,
      y: belowValues[index],
    })),
    baselines: {
      above: firstNonNullValue(aboveValues),
      below: firstNonNullValue(belowValues),
    },
    dates,
  };
}

function parseChartConfig(config) {
  return {
    lookbackSeconds: Number(config?.lookback_seconds) || DEFAULT_LOOKBACK_SECONDS,
    bucketSeconds: Number(config?.bucket_seconds) || DEFAULT_BUCKET_SECONDS,
  };
}

function projectChartPoints(points, config, now) {
  const { lookbackSeconds } = parseChartConfig(config);
  const nowDate = now instanceof Date ? now : new Date(now);
  const nowMs = nowDate.getTime();

  if (!Number.isFinite(nowMs)) return points ?? [];

  const windowStartMs = nowMs - lookbackSeconds * 1000;

  const canonicalPoints = (points ?? []).filter((point) => {
    const pointMs = parseSnapshotTimestamp(point.bucket_start ?? point.inserted_at)?.getTime();
    return Number.isFinite(pointMs) && pointMs >= windowStartMs;
  });

  const lastPoint = canonicalPoints[canonicalPoints.length - 1];
  if (!lastPoint) return canonicalPoints;

  const lastPointMs =
    parseSnapshotTimestamp(lastPoint.bucket_start ?? lastPoint.inserted_at)?.getTime();

  if (!Number.isFinite(lastPointMs) || lastPointMs >= nowMs) {
    return canonicalPoints;
  }

  return [
    ...canonicalPoints,
    {
      ...lastPoint,
      bucket_start: nowDate.toISOString(),
      inserted_at: nowDate.toISOString(),
    },
  ];
}

export function renderChart(canvas, contract, chart, now = new Date()) {
  if (!canvas || !contract || !chart?.points || chart.points.length === 0) return;
  const theme = getChartTheme();

  const projectedPoints = projectChartPoints(chart.points, chart.config, now);
  const { xValues, above, below, dates, baselines } = buildChartData(projectedPoints);
  const aboveSeries = splitHistoricalAndLiveSeries(above);
  const belowSeries = splitHistoricalAndLiveSeries(below);
  const chartConfig = parseChartConfig(chart.config);

  ensureChartChrome(canvas, theme, chart.config);

  if (!chartInstance) {
    chartInstance = new Chart(canvas, {
      type: "line",
      data: {
        datasets: [
          {
            label: "Above",
            data: aboveSeries.historical,
            borderColor: withAlpha(theme.chartAbove, 0.92),
            backgroundColor: theme.chartTransparent,
            pointRadius: 0,
            pointHoverRadius: 0,
            pointHitRadius: 12,
            pointHoverBorderWidth: 0,
            borderWidth: 2,
            borderCapStyle: "round",
            borderJoinStyle: "round",
            tension: 0.22,
            spanGaps: true
          },
          {
            label: "Above Live Tail",
            data: aboveSeries.liveTail,
            borderColor: withAlpha(theme.chartAbove, 0.55),
            backgroundColor: theme.chartTransparent,
            pointRadius: 0,
            pointHoverRadius: 0,
            pointHitRadius: 0,
            pointHoverBorderWidth: 0,
            borderWidth: 2,
            borderCapStyle: "round",
            borderJoinStyle: "round",
            tension: 0.22,
            spanGaps: true
          },
          {
            label: "Below",
            data: belowSeries.historical,
            borderColor: withAlpha(theme.chartBelow, 0.92),
            backgroundColor: theme.chartTransparent,
            pointRadius: 0,
            pointHoverRadius: 0,
            pointHitRadius: 12,
            pointHoverBorderWidth: 0,
            borderWidth: 2,
            borderCapStyle: "round",
            borderJoinStyle: "round",
            tension: 0.22,
            spanGaps: true
          },
          {
            label: "Below Live Tail",
            data: belowSeries.liveTail,
            borderColor: withAlpha(theme.chartBelow, 0.55),
            backgroundColor: theme.chartTransparent,
            pointRadius: 0,
            pointHoverRadius: 0,
            pointHitRadius: 0,
            pointHoverBorderWidth: 0,
            borderWidth: 2,
            borderCapStyle: "round",
            borderJoinStyle: "round",
            tension: 0.22,
            spanGaps: true
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          mode: "nearest",
          intersect: false
        },
        plugins: {
          title: {
            display: false
          },
          legend: {
            display: false
          },
          tooltip: {
            enabled: false
          }
        },
        layout: {
          padding: chartLayoutPadding()
        },
        scales: {
          x: {
            type: "linear",
            min: xValues[0],
            max: xValues[xValues.length - 1],
            afterBuildTicks(scale) {
              scale.ticks = buildDeterministicXTicks(
                scale.min,
                scale.max,
                chartConfig.lookbackSeconds
              );
            },
            border: {
              display: false
            },
            grid: {
              color: theme.chartGrid,
              display: true,
              drawOnChartArea: true,
              drawTicks: true,
              tickLength: theme.spaceSm
            },
            ticks: {
              color: theme.textMuted,
              autoSkip: false,
              maxTicksLimit: xTickCount(chartConfig.lookbackSeconds),
              minRotation: 0,
              maxRotation: 0,
              callback(value) {
                return formatTickLabel(new Date(Number(value)), chartConfig.lookbackSeconds);
              },
              font: {
                size: theme.fontSizeCaption
              }
            }
          },
          y: {
            position: "right",
            grace: "6%",
            border: {
              display: true,
              color: theme.chartAxisBorder
            },
            grid: {
              color: theme.chartGrid,
              lineWidth: 1
            },
            ticks: {
              color: theme.textMuted,
              mirror: false,
              padding: theme.spaceMd,
              callback(value) {
                return formatPrice(Number(value));
              },
              font: {
                size: theme.fontSizeCaption,
                weight: theme.fontWeightNormal
              }
            }
          }
        }
      }
    });
  } else {
    chartInstance.data.datasets[0].data = aboveSeries.historical;
    chartInstance.data.datasets[1].data = aboveSeries.liveTail;
    chartInstance.data.datasets[2].data = belowSeries.historical;
    chartInstance.data.datasets[3].data = belowSeries.liveTail;

    chartInstance.options.scales.x.min = xValues[0];
    chartInstance.options.scales.x.max = xValues[xValues.length - 1];
    chartInstance.options.layout.padding = chartLayoutPadding();

    chartInstance.data.datasets[0].borderColor = withAlpha(theme.chartAbove, 0.92);
    chartInstance.data.datasets[0].backgroundColor = theme.chartTransparent;
    chartInstance.data.datasets[0].borderWidth = 2;
    chartInstance.data.datasets[0].borderCapStyle = "round";
    chartInstance.data.datasets[0].borderJoinStyle = "round";
    chartInstance.data.datasets[0].tension = 0.22;
    chartInstance.data.datasets[0].pointRadius = 0;
    chartInstance.data.datasets[0].pointHoverRadius = 0;
    chartInstance.data.datasets[0].pointHitRadius = 12;
    chartInstance.data.datasets[0].pointHoverBorderWidth = 0;

    chartInstance.data.datasets[1].borderColor = withAlpha(theme.chartAbove, 0.55);
    chartInstance.data.datasets[1].backgroundColor = theme.chartTransparent;
    chartInstance.data.datasets[1].borderWidth = 2;
    chartInstance.data.datasets[1].borderCapStyle = "round";
    chartInstance.data.datasets[1].borderJoinStyle = "round";
    chartInstance.data.datasets[1].tension = 0.22;
    chartInstance.data.datasets[1].pointRadius = 0;
    chartInstance.data.datasets[1].pointHoverRadius = 0;
    chartInstance.data.datasets[1].pointHitRadius = 0;
    chartInstance.data.datasets[1].pointHoverBorderWidth = 0;

    chartInstance.data.datasets[2].borderColor = withAlpha(theme.chartBelow, 0.92);
    chartInstance.data.datasets[2].backgroundColor = theme.chartTransparent;
    chartInstance.data.datasets[2].borderWidth = 2;
    chartInstance.data.datasets[2].borderCapStyle = "round";
    chartInstance.data.datasets[2].borderJoinStyle = "round";
    chartInstance.data.datasets[2].tension = 0.22;
    chartInstance.data.datasets[2].pointRadius = 0;
    chartInstance.data.datasets[2].pointHoverRadius = 0;
    chartInstance.data.datasets[2].pointHitRadius = 12;
    chartInstance.data.datasets[2].pointHoverBorderWidth = 0;

    chartInstance.data.datasets[3].borderColor = withAlpha(theme.chartBelow, 0.55);
    chartInstance.data.datasets[3].backgroundColor = theme.chartTransparent;
    chartInstance.data.datasets[3].borderWidth = 2;
    chartInstance.data.datasets[3].borderCapStyle = "round";
    chartInstance.data.datasets[3].borderJoinStyle = "round";
    chartInstance.data.datasets[3].tension = 0.22;
    chartInstance.data.datasets[3].pointRadius = 0;
    chartInstance.data.datasets[3].pointHoverRadius = 0;
    chartInstance.data.datasets[3].pointHitRadius = 0;
    chartInstance.data.datasets[3].pointHoverBorderWidth = 0;

    chartInstance.options.scales.x.afterBuildTicks = function (scale) {
      scale.ticks = buildDeterministicXTicks(
        scale.min,
        scale.max,
        chartConfig.lookbackSeconds
      );
    };
    chartInstance.options.scales.x.grid.color = theme.chartGrid;
    chartInstance.options.scales.x.ticks.maxTicksLimit = xTickCount(chartConfig.lookbackSeconds);
    chartInstance.options.scales.x.ticks.callback = function (value) {
      return formatTickLabel(new Date(Number(value)), chartConfig.lookbackSeconds);
    };
    chartInstance.options.scales.y.grid.color = theme.chartGrid;
    chartInstance.options.scales.y.border.color = theme.chartAxisBorder;
    chartInstance.options.scales.x.ticks.color = theme.textMuted;
    chartInstance.options.scales.y.ticks.color = theme.textMuted;

    chartInstance.$theme = theme;
    chartInstance.$dates = dates;
    chartInstance.$baselines = baselines;
    chartInstance.$xValues = xValues;
    chartInstance.$projectedPoints = projectedPoints;
    chartInstance.$config = chart.config;
    chartInstance.$liveDate = now instanceof Date ? now : new Date(now);

    updateChartChrome(chartInstance);
    chartInstance.update("none");
  }

  chartInstance.$theme = theme;
  chartInstance.$dates = dates;
  chartInstance.$baselines = baselines;
  chartInstance.$xValues = xValues;
  chartInstance.$projectedPoints = projectedPoints;
  chartInstance.$config = chart.config;
  chartInstance.$liveDate = now instanceof Date ? now : new Date(now);

  updateChartChrome(chartInstance);
}

import maplibregl from "maplibre-gl";

const REGION_MAP_STYLE =
  "https://basemaps.cartocdn.com/gl/positron-gl-style/style.json";

const DEFAULT_VIEW = {
  center: [-80.1918, 25.7617],
  zoom: 11,
};

const REGION_VIEWS = new Map([
  ["miami", DEFAULT_VIEW],
]);

function regionView(contract) {
  const key = contract?.region_label?.trim().toLowerCase();
  return REGION_VIEWS.get(key) ?? DEFAULT_VIEW;
}

function cssToken(styles, name) {
  return styles.getPropertyValue(name).trim();
}

function getRegionMapTheme() {
  const styles = getComputedStyle(document.documentElement);

  return {
    water: cssToken(styles, "--color-map-water"),
    land: cssToken(styles, "--color-map-land"),
    road: cssToken(styles, "--color-map-road"),
    label: cssToken(styles, "--color-map-label"),
    labelHalo: cssToken(styles, "--color-map-label-halo"),
    durationMapEase: Number.parseFloat(cssToken(styles, "--duration-map-ease")) || 420,
  };
}

function applyMinimalMapTheme(map) {
  const theme = getRegionMapTheme();

  const paintPropertyApplies = (layerType, prop) => {
    if (prop.startsWith("background-")) return layerType === "background";
    if (prop.startsWith("fill-")) return layerType === "fill";
    if (prop.startsWith("line-")) return layerType === "line";
    if (prop.startsWith("text-")) return layerType === "symbol";

    return true;
  };

  const setPaint = (layerId, prop, value) => {
    const layer = map.getLayer(layerId);

    if (layer && paintPropertyApplies(layer.type, prop)) {
      map.setPaintProperty(layerId, prop, value);
    }
  };

  const setLayout = (layerId, prop, value) => {
    if (map.getLayer(layerId)) {
      map.setLayoutProperty(layerId, prop, value);
    }
  };

  [
    "water",
    "water-fill",
    "water_shadow",
    "waterway",
    "waterway-shadow"
  ].forEach((id) => {
    setPaint(id, "fill-color", theme.water);
    setPaint(id, "line-color", theme.water);
  });

  [
    "background",
    "landcover",
    "landuse",
    "park",
    "land"
  ].forEach((id) => {
    setPaint(id, "background-color", theme.land);
    setPaint(id, "fill-color", theme.land);
  });

  [
    "road",
    "road_minor",
    "road_major",
    "road_trunk_primary",
    "road_secondary_tertiary",
    "road_street",
    "road_service_track",
    "bridge_main",
    "bridge_minor"
  ].forEach((id) => {
    setPaint(id, "line-color", theme.road);
  });

  // Hide nearly all labels
  [
    "road_name",
    "poi_label",
    "water_name",
    "country_label",
    "marine_label",
    "state_label",
    "place_label_other"
  ].forEach((id) => {
    setLayout(id, "visibility", "none");
  });

  // Keep only city/place label
  ["place_label"].forEach((id) => {
    setPaint(id, "text-color", theme.label);
    setPaint(id, "text-halo-color", theme.labelHalo);
  });
}


function createResizeScheduler(map) {
  let queued = false;

  return function scheduleResize() {
    if (queued) return;
    queued = true;

    requestAnimationFrame(() => {
      queued = false;
      map.resize();
    });
  };
}

function runInitialResizeSequence(scheduleResize) {
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      scheduleResize();

      [120, 300, 700].forEach((delay) => {
        window.setTimeout(scheduleResize, delay);
      });
    });
  });
}

export function initRegionMap(container, contract = null) {
  if (!container || container.dataset.mapInitialized === "true") return null;

  const view = regionView(contract);

  const map = new maplibregl.Map({
    container,
    style: REGION_MAP_STYLE,
    center: view.center,
    zoom: view.zoom,
    attributionControl: false,
    pitch: 36,
    bearing: -10,
    pitchWithRotate: false,
    dragRotate: false,
    interactive: false,
  });

  const scheduleResize = createResizeScheduler(map);

  map.addControl(
    new maplibregl.AttributionControl({
      compact: true,
    }),
    "bottom-right",
  );

  map.touchZoomRotate.disableRotation();
  map.boxZoom.disable();

  container.dataset.mapInitialized = "true";
  container.regionMap = map;
  container.regionMapScheduleResize = scheduleResize;

  if ("ResizeObserver" in window) {
    const observer = new ResizeObserver(() => {
      scheduleResize();
    });

    observer.observe(container);

    const panel = container.closest(".panel--region-map");
    if (panel) observer.observe(panel);

    const marketMain = container.closest(".market-main");
    if (marketMain) observer.observe(marketMain);

    container.regionMapResizeObserver = observer;
  }

  const handleWindowResize = () => scheduleResize();
  window.addEventListener("resize", handleWindowResize, { passive: true });
  container.regionMapWindowResizeHandler = handleWindowResize;

  map.once("load", () => {
    applyMinimalMapTheme(map);
    runInitialResizeSequence(scheduleResize);
  });
  runInitialResizeSequence(scheduleResize);

  return map;
}

export function syncRegionMap(container, contract) {
  const map = container?.regionMap;
  if (!map || !contract) return;

  const view = regionView(contract);
  map.easeTo({
    center: view.center,
    zoom: view.zoom,
    duration: getRegionMapTheme().durationMapEase,
    essential: true,
  });

  container.regionMapScheduleResize?.();
}

import { useEffect, useRef } from "react";
import { destroyRegionMap, initRegionMap, syncRegionMap } from "../lib/regionMap";

export function RegionMapPanel({ contract }) {
  const mapRef = useRef(null);

  useEffect(() => {
    if (!mapRef.current) return undefined;
    initRegionMap(mapRef.current, contract);
    return () => destroyRegionMap(mapRef.current);
  }, []);

  useEffect(() => {
    if (!mapRef.current || !contract) return;
    syncRegionMap(mapRef.current, contract);
  }, [contract]);

  return (
    <section className="panel panel--region-map">
      <div className="region-map__header">
        <h2 className="region-map__title">Region Map</h2>
      </div>

      <div className="region-map__canvas">
        <div id="region-map" ref={mapRef}></div>
      </div>
    </section>
  );
}

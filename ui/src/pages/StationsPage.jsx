import { useDeferredValue, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { listStations } from "../lib/api";

function metricLabel(metric) {
  return String(metric || "")
    .split(/[_\s]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function stationSearchText(station) {
  return [
    station.display_name,
    station.station_id,
    station.region,
    station.provider_name,
    ...(station.supported_metrics || []),
  ]
    .filter(Boolean)
    .join(" ")
    .toLowerCase();
}

export function StationsPage() {
  const [search, setSearch] = useState("");
  const deferredSearch = useDeferredValue(search);
  const stationsQuery = useQuery({
    queryKey: ["stations"],
    queryFn: () => listStations(true),
  });

  const filteredStations = useMemo(() => {
    const stations = stationsQuery.data || [];
    const query = deferredSearch.trim().toLowerCase();

    if (!query) {
      return stations;
    }

    return stations.filter((station) => stationSearchText(station).includes(query));
  }, [deferredSearch, stationsQuery.data]);

  return (
    <div className="station-directory-page">
      <header className="station-directory-page__hero">
        <div>
          <p className="station-directory-page__eyebrow">Station Directory</p>
          <h1 className="station-directory-page__title">Search weather stations before you create a market.</h1>
        </div>
        <p className="station-directory-page__subtitle">
          Pick a station to review its active markets and create a new contract tied to that location.
        </p>
      </header>

      <section className="station-directory-page__search-shell panel" aria-label="Search stations">
        <label className="station-form__field">
          <span className="station-form__label">Search by name, station code, region, provider, or metric</span>
          <input
            className="station-form__input"
            type="search"
            placeholder="Halifax, NOAA, temperature, station code..."
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </label>
      </section>

      <section className="station-directory-page__results" aria-label="Station results">
        {stationsQuery.isLoading ? (
          <div className="station-empty panel">Loading stations...</div>
        ) : null}

        {stationsQuery.isError ? (
          <div className="station-empty panel">
            {stationsQuery.error.payload?.error || stationsQuery.error.message}
          </div>
        ) : null}

        {!stationsQuery.isLoading && !stationsQuery.isError && filteredStations.length === 0 ? (
          <div className="station-empty panel">No stations matched your search.</div>
        ) : null}

        {!stationsQuery.isLoading && !stationsQuery.isError
          ? filteredStations.map((station) => (
              <Link
                key={station.id}
                className="station-card panel"
                to={`/station/${station.id}`}
              >
                <div className="station-card__header">
                  <div>
                    <h2 className="station-card__title">{station.display_name}</h2>
                    <p className="station-card__meta">
                      {station.provider_name} · {station.station_id}
                    </p>
                  </div>
                  <span className={`station-card__status${station.active ? " station-card__status--active" : ""}`}>
                    {station.active ? "Active" : "Inactive"}
                  </span>
                </div>

                <p className="station-card__region">{station.region || "Region unavailable"}</p>

                <div className="station-card__metrics">
                  {(station.supported_metrics || []).length === 0 ? (
                    <span className="station-card__metric">Metrics unavailable</span>
                  ) : (
                    station.supported_metrics.map((metric) => (
                      <span key={metric} className="station-card__metric">
                        {metricLabel(metric)}
                      </span>
                    ))
                  )}
                </div>
              </Link>
            ))
          : null}
      </section>
    </div>
  );
}

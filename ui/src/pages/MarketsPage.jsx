import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { listContracts } from "../lib/api";

function titleCase(value) {
  return String(value || "")
    .split(/[_\s]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function contractSortKey(contract) {
  return contract.measurement_period_end || contract.trading_period_end || "";
}

export function MarketsPage() {
  const contractsQuery = useQuery({
    queryKey: ["contracts"],
    queryFn: listContracts,
  });

  const contracts = useMemo(
    () => [...(contractsQuery.data || [])].sort((left, right) => contractSortKey(right).localeCompare(contractSortKey(left))),
    [contractsQuery.data],
  );

  return (
    <section className="markets-page">
      <header className="markets-page__header panel">
        <div>
          <p className="station-directory-page__eyebrow">Markets</p>
          <h1 className="markets-page__title">Browse open and draft weather markets.</h1>
        </div>
        <p className="markets-page__subtitle">
          Open a market to trade, or start from stations if you want to create a new one.
        </p>
      </header>

      {contractsQuery.isLoading ? <div className="panel portfolio-panel__empty">Loading markets...</div> : null}
      {contractsQuery.isError ? (
        <div className="panel portfolio-panel__empty">
          {contractsQuery.error.payload?.error || contractsQuery.error.message}
        </div>
      ) : null}

      {!contractsQuery.isLoading && !contractsQuery.isError ? (
        <div className="markets-page__grid">
          {contracts.length === 0 ? (
            <div className="panel portfolio-panel__empty">No markets available.</div>
          ) : (
            contracts.map((contract) => (
              <Link key={contract.id} className="market-card panel" to={`/contracts/${contract.id}`}>
                <div className="market-card__header">
                  <h2 className="market-card__title">{contract.name}</h2>
                  <span className="market-card__status">{contract.status}</span>
                </div>
                <p className="market-card__meta">
                  {contract.region || "Region unavailable"} | {titleCase(contract.metric)}
                </p>
                <div className="market-card__grid">
                  <span>Threshold: {contract.threshold ?? "-"}</span>
                  <span>Payout: ${(Number(contract.multiplier || 100) / 100).toFixed(2)}</span>
                  <span>Trading: {contract.trading_period_start || "-"} to {contract.trading_period_end || "-"}</span>
                  <span>Measurement: {contract.measurement_period_start || "-"} to {contract.measurement_period_end || "-"}</span>
                </div>
              </Link>
            ))
          )}
        </div>
      ) : null}
    </section>
  );
}

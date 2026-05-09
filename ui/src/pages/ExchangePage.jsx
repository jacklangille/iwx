import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams } from "react-router-dom";
import { MarketChart } from "../components/MarketChart";
import { TradeRail } from "../components/TradeRail";
import {
  getMarketSnapshots,
  getMarketState,
  getOrderCommand,
  getPortfolio,
  listContracts,
  submitOrder,
} from "../lib/api";
import { useAuth } from "../lib/auth";
import {
  buildContractDescription,
  buildContractLabel,
  contractStatusClassName,
  formatDateRange,
  formatMeasurementUnit,
} from "../lib/formatters";
import { useChrome } from "../lib/chrome";

export function ExchangePage() {
  const queryClient = useQueryClient();
  const { token, isAuthenticated } = useAuth();
  const { openLogin } = useChrome();
  const navigate = useNavigate();
  const { contractId: routeContractId } = useParams();
  const [chartConfig, setChartConfig] = useState({
    lookbackSeconds: 24 * 60 * 60,
    bucketSeconds: 5 * 60,
  });
  const [chartNow, setChartNow] = useState(() => new Date());
  const [orderBookView, setOrderBookView] = useState("above");
  const [orderCommandId, setOrderCommandId] = useState(null);
  const [orderError, setOrderError] = useState("");

  useEffect(() => {
    const intervalId = window.setInterval(() => setChartNow(new Date()), 5_000);
    return () => window.clearInterval(intervalId);
  }, []);

  const contractsQuery = useQuery({
    queryKey: ["contracts"],
    queryFn: listContracts,
  });

  const selectedContract = useMemo(() => {
    const contracts = contractsQuery.data || [];
    if (!routeContractId) return contracts[0] || null;

    return contracts.find((contract) => String(contract.id) === String(routeContractId)) || null;
  }, [contractsQuery.data, routeContractId]);

  useEffect(() => {
    if (routeContractId || !selectedContract) return;
    navigate(`/contracts/${selectedContract.id}`, { replace: true });
  }, [navigate, routeContractId, selectedContract]);

  const contractId = selectedContract?.id || null;

  const marketStateQuery = useQuery({
    queryKey: ["market-state", contractId],
    queryFn: () => getMarketState(contractId),
    enabled: Boolean(contractId),
    refetchInterval: contractId ? 2_000 : false,
  });

  const chartQuery = useQuery({
    queryKey: ["market-chart", contractId, chartConfig.lookbackSeconds, chartConfig.bucketSeconds],
    queryFn: () => getMarketSnapshots(contractId, chartConfig),
    enabled: Boolean(contractId),
    refetchInterval: contractId ? 5_000 : false,
  });

  const portfolioQuery = useQuery({
    queryKey: ["portfolio", token],
    queryFn: () => getPortfolio(token),
    enabled: isAuthenticated,
    refetchInterval: isAuthenticated ? 5_000 : false,
  });

  const orderCommandQuery = useQuery({
    queryKey: ["order-command", token, orderCommandId],
    queryFn: () => getOrderCommand(token, orderCommandId),
    enabled: Boolean(token && orderCommandId),
    refetchInterval(query) {
      const status = String(query.state.data?.status || "").toLowerCase();
      if (!status) return 1_500;
      if (["queued", "processing", "enqueued"].includes(status)) return 1_500;
      return false;
    },
  });

  useEffect(() => {
    const status = String(orderCommandQuery.data?.status || "").toLowerCase();
    if (!status || ["queued", "processing", "enqueued"].includes(status)) return;

    queryClient.invalidateQueries({ queryKey: ["market-state", contractId] });
    queryClient.invalidateQueries({ queryKey: ["market-chart", contractId] });
    queryClient.invalidateQueries({ queryKey: ["portfolio", token] });
  }, [contractId, orderCommandQuery.data, queryClient, token]);

  const orderMutation = useMutation({
    mutationFn: (payload) => submitOrder(token, payload),
    onSuccess(payload) {
      setOrderError("");
      setOrderCommandId(payload.command_id);
    },
    onError(error) {
      const nextError =
        error.payload?.errors
          ? Object.values(error.payload.errors).flat().join(", ")
          : error.payload?.error || error.message;
      setOrderError(nextError);
    },
  });

  const contractView = selectedContract
    ? {
        ...selectedContract,
        trading_period: {
          start: selectedContract.trading_period_start,
          end: selectedContract.trading_period_end,
        },
        measurement_period: {
          start: selectedContract.measurement_period_start,
          end: selectedContract.measurement_period_end,
        },
        data_provider: {
          name: selectedContract.data_provider_name,
          station_mode: selectedContract.data_provider_station_mode,
        },
      }
    : null;

  return (
    <section className="exchange-grid">
      <div className="exchange-grid__main">
        <section className="panel--market-header">
          <div className="market-header">
            <h1 className="market-header__title" id="contract-title">
              {contractView ? buildContractLabel(contractView) : "Loading market..."}
            </h1>

            <div className="market-meta-rail" aria-label="Contract metadata">
              <span
                id="market-status-badge"
                className={contractStatusClassName(contractView?.status)}
              >
                {contractView?.status || "Loading"}
              </span>
              <span className="market-meta-rail__item" id="contract-threshold">
                {contractView?.threshold != null
                  ? `Threshold ${Number(contractView.threshold)}${formatMeasurementUnit(contractView.measurement_unit)}`
                  : "Threshold -"}
              </span>
              <span className="market-meta-rail__item" id="market-trading-period">
                Trading {formatDateRange(contractView?.trading_period)}
              </span>
              <span className="market-meta-rail__item" id="market-measurement-period">
                Measurement {formatDateRange(contractView?.measurement_period)}
              </span>
              <span className="market-meta-rail__item" id="contract-data-provider">
                {contractView?.data_provider?.name
                  ? `${contractView.data_provider.name} ${
                      contractView.data_provider.station_mode === "single_station"
                        ? "Single Station"
                        : contractView.data_provider.station_mode || ""
                    }`.trim()
                  : "-"}
              </span>
            </div>

            <div className="market-header__description" id="contract-description">
              {buildContractDescription(contractView)}
            </div>
          </div>
        </section>

        <div className="market-main">
          <MarketChart
            contract={contractView}
            chartData={chartQuery.data}
            now={chartNow}
            onConfigChange={(nextConfig) =>
              setChartConfig({
                lookbackSeconds: nextConfig.lookbackSeconds,
                bucketSeconds: nextConfig.bucketSeconds,
              })
            }
          />
        </div>
      </div>

      <TradeRail
        contract={contractView}
        marketState={marketStateQuery.data}
        portfolio={portfolioQuery.data}
        token={token}
        onRequireLogin={openLogin}
        onSubmitOrder={(payload) => orderMutation.mutate(payload)}
        orderPending={orderMutation.isPending}
        orderCommand={orderCommandQuery.data}
        orderError={orderError}
        orderBookView={orderBookView}
        onOrderBookViewChange={setOrderBookView}
      />
    </section>
  );
}

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Alert,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from "@mui/material";
import { useState } from "react";
import { depositFunds, getPortfolio, getUserSettlements } from "../lib/api";
import { useAuth } from "../lib/auth";
import { formatMoneyCents, formatQuantity, formatTimestamp } from "../lib/formatters";

function EmptyAuthState() {
  return (
    <section className="portfolio-page">
      <div className="portfolio-page__grid">
        <section className="panel portfolio-panel">
          <h1 className="portfolio-panel__title">Portfolio</h1>
          <p className="portfolio-panel__subtitle">
            Log in to view balances, positions, reservations, and settlement history.
          </p>
        </section>
      </div>
    </section>
  );
}

export function PortfolioPage() {
  const queryClient = useQueryClient();
  const { token, isAuthenticated } = useAuth();
  const [depositAmount, setDepositAmount] = useState("1000");
  const [depositError, setDepositError] = useState("");
  const [depositSuccess, setDepositSuccess] = useState("");

  const portfolioQuery = useQuery({
    queryKey: ["portfolio", token],
    queryFn: () => getPortfolio(token),
    enabled: isAuthenticated,
  });
  const settlementsQuery = useQuery({
    queryKey: ["portfolio-settlements", token],
    queryFn: () => getUserSettlements(token),
    enabled: isAuthenticated,
  });

  const depositMutation = useMutation({
    mutationFn: async (amountDollars) => {
      const amountNumber = Number.parseFloat(String(amountDollars || ""));
      if (!Number.isFinite(amountNumber) || amountNumber <= 0) {
        throw new Error("Enter a valid deposit amount");
      }

      return depositFunds(token, {
        currency: "USD",
        amount_cents: Math.round(amountNumber * 100),
        reference_id: `ui-deposit-${Date.now()}`,
        description: "UI dev funding",
      });
    },
    onSuccess(payload) {
      setDepositError("");
      setDepositSuccess(
        `Deposited ${formatMoneyCents(payload?.ledger_entry?.amount_cents || 0)} into your USD account.`,
      );
      queryClient.invalidateQueries({ queryKey: ["portfolio", token] });
      queryClient.invalidateQueries({ queryKey: ["portfolio-settlements", token] });
    },
    onError(error) {
      setDepositSuccess("");
      setDepositError(error?.payload?.error || error?.message || "Deposit failed");
    },
  });

  if (!isAuthenticated) return <EmptyAuthState />;

  const portfolio = portfolioQuery.data || {};
  const settlements = settlementsQuery.data || [];

  const handleDepositSubmit = (event) => {
    event.preventDefault();
    setDepositError("");
    setDepositSuccess("");
    depositMutation.mutate(depositAmount);
  };

  return (
    <section className="portfolio-page">
      <div className="portfolio-page__grid">
        <section className="panel portfolio-panel">
          <h1 className="portfolio-panel__title">Portfolio</h1>
          <p className="portfolio-panel__subtitle">
            Read-model snapshot from the Go services.
          </p>
        </section>

        <section className="panel portfolio-panel">
          <div className="portfolio-panel__header">
            <h2 className="portfolio-panel__section-title">Fund account</h2>
          </div>
          <p className="portfolio-panel__subtitle">
            Local dev funding only. This calls the exchange-core deposit endpoint directly and credits your internal USD balance.
          </p>
          <form className="portfolio-funding-form" onSubmit={handleDepositSubmit}>
            <label className="portfolio-funding-form__field">
              <span className="portfolio-funding-form__label">Deposit amount (USD)</span>
              <input
                className="portfolio-funding-form__input"
                type="number"
                min="1"
                step="1"
                value={depositAmount}
                onChange={(event) => setDepositAmount(event.target.value)}
                placeholder="1000"
              />
            </label>
            <div className="portfolio-funding-form__actions">
              <button className="topbar__signup" type="submit" disabled={depositMutation.isPending}>
                {depositMutation.isPending ? "Funding..." : "Fund account"}
              </button>
            </div>
          </form>
          {depositError ? <div className="station-form__error">{depositError}</div> : null}
          {depositSuccess ? <div className="station-form__success">{depositSuccess}</div> : null}
        </section>

        <DataPanel
          title="Cash Accounts"
          loading={portfolioQuery.isLoading}
          rows={portfolio.accounts || []}
          columns={[
            ["Currency", (row) => row.currency],
            ["Available", (row) => formatMoneyCents(row.available_cents)],
            ["Locked", (row) => formatMoneyCents(row.locked_cents)],
            ["Total", (row) => formatMoneyCents(row.total_cents)],
          ]}
        />

        <DataPanel
          title="Positions"
          loading={portfolioQuery.isLoading}
          rows={portfolio.positions || []}
          columns={[
            ["Contract", (row) => row.contract_id],
            ["Side", (row) => row.side],
            ["Available", (row) => formatQuantity(row.available_quantity)],
            ["Locked", (row) => formatQuantity(row.locked_quantity)],
            ["Total", (row) => formatQuantity(row.total_quantity)],
          ]}
        />

        <DataPanel
          title="Position Locks"
          loading={portfolioQuery.isLoading}
          rows={portfolio.position_locks || []}
          columns={[
            ["Contract", (row) => row.contract_id],
            ["Side", (row) => row.side],
            ["Quantity", (row) => formatQuantity(row.quantity)],
            ["Status", (row) => row.status],
            ["Updated", (row) => formatTimestamp(row.updated_at)],
          ]}
        />

        <DataPanel
          title="Collateral Locks"
          loading={portfolioQuery.isLoading}
          rows={portfolio.collateral_locks || []}
          columns={[
            ["Contract", (row) => row.contract_id],
            ["Amount", (row) => formatMoneyCents(row.amount_cents)],
            ["Status", (row) => row.status],
            ["Updated", (row) => formatTimestamp(row.updated_at)],
          ]}
        />

        <DataPanel
          title="Cash Reservations"
          loading={portfolioQuery.isLoading}
          rows={portfolio.cash_reservations || []}
          columns={[
            ["Contract", (row) => row.contract_id],
            ["Amount", (row) => formatMoneyCents(row.amount_cents)],
            ["Status", (row) => row.status],
            ["Updated", (row) => formatTimestamp(row.updated_at)],
          ]}
        />

        <DataPanel
          title="Settlements"
          loading={settlementsQuery.isLoading}
          rows={settlements}
          columns={[
            ["Contract", (row) => row.contract_id],
            ["Type", (row) => row.entry_type],
            ["Outcome", (row) => row.outcome],
            ["Amount", (row) => formatMoneyCents(row.amount_cents)],
            ["Created", (row) => formatTimestamp(row.created_at)],
          ]}
        />
      </div>
    </section>
  );
}

function DataPanel({ title, rows, columns, loading }) {
  return (
    <section className="panel portfolio-panel">
      <div className="portfolio-panel__header">
        <h2 className="portfolio-panel__section-title">{title}</h2>
      </div>

      {loading ? (
        <div className="portfolio-panel__empty">Loading...</div>
      ) : rows.length === 0 ? (
        <Alert severity="info" variant="outlined">
          No rows
        </Alert>
      ) : (
        <TableContainer component={Paper} variant="outlined" className="data-table-shell">
          <Table size="small" stickyHeader>
            <TableHead>
              <TableRow>
                {columns.map(([label]) => (
                  <TableCell key={label}>{label}</TableCell>
                ))}
              </TableRow>
            </TableHead>
            <TableBody>
              {rows.map((row, index) => (
                <TableRow hover key={`${title}-${index}`}>
                  {columns.map(([label, render]) => (
                    <TableCell key={label}>{render(row)}</TableCell>
                  ))}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </section>
  );
}

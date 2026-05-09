import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useParams } from "react-router-dom";
import {
  activateContract,
  approveContract,
  createContract,
  createIssuanceBatch,
  getCollateralRequirement,
  getContractCommand,
  getPortfolio,
  listContracts,
  listIssuanceBatches,
  listStations,
  lockContractCollateral,
  submitContractForApproval,
} from "../lib/api";
import { useAuth } from "../lib/auth";
import { useChrome } from "../lib/chrome";
import { formatMoneyCents } from "../lib/formatters";

const OPEN_CONTRACT_STATUSES = new Set([
  "pending_approval",
  "pending_collateral",
  "active",
  "trading_closed",
  "awaiting_resolution",
]);

const DEFAULT_METRIC_UNITS = {
  temperature_max: "F",
  temperature_min: "F",
  temperature_avg: "F",
  precipitation_total: "mm",
  snowfall_total: "cm",
  wind_speed_max: "mph",
};

function titleCase(value) {
  return String(value || "")
    .split(/[_\s]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function defaultMetric(station) {
  const metrics = station?.supported_metrics || [];
  return metrics[0] || "";
}

function buildInitialForm(station) {
  const metric = defaultMetric(station);

  return {
    name: "",
    metric,
    threshold: "",
    multiplier: "100",
    pairedQuantity: "",
    measurementUnit: DEFAULT_METRIC_UNITS[metric] || "",
    tradingPeriodStart: "",
    tradingPeriodEnd: "",
    measurementPeriodStart: "",
    measurementPeriodEnd: "",
    description: "",
  };
}

function validationErrorText(error) {
  if (error?.payload?.errors) {
    return Object.entries(error.payload.errors)
      .flatMap(([field, messages]) => messages.map((message) => `${field} ${message}`))
      .join(", ");
  }

  return error?.payload?.error || error?.message || "Request failed";
}

function contractSortKey(contract) {
  return contract.measurement_period_end || contract.trading_period_end || "";
}

function parsePositiveInt(value) {
  const parsed = Number.parseInt(String(value || ""), 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 0;
}

function normalizeStatus(value) {
  return String(value || "").trim().toLowerCase();
}

function collateralForDraft(payoutCents, pairedQuantity) {
  return Math.max(payoutCents * pairedQuantity, 100000);
}

export function StationDetailPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { stationId } = useParams();
  const { session, token, isAuthenticated } = useAuth();
  const { openLogin } = useChrome();
  const stationsQuery = useQuery({
    queryKey: ["stations"],
    queryFn: () => listStations(true),
  });
  const contractsQuery = useQuery({
    queryKey: ["contracts"],
    queryFn: listContracts,
  });
  const portfolioQuery = useQuery({
    queryKey: ["portfolio", token],
    queryFn: () => getPortfolio(token),
    enabled: isAuthenticated,
  });

  const station = useMemo(
    () => (stationsQuery.data || []).find((item) => String(item.id) === String(stationId)) || null,
    [stationId, stationsQuery.data],
  );

  const [formState, setFormState] = useState(() => buildInitialForm(null));
  const [formError, setFormError] = useState("");
  const [formSuccess, setFormSuccess] = useState("");
  const [activationDrafts, setActivationDrafts] = useState({});

  const sessionUserId = Number(session?.userId || 0);

  useEffect(() => {
    if (!station) return;
    setFormState(buildInitialForm(station));
  }, [station]);

  const openContracts = useMemo(() => {
    const contracts = contractsQuery.data || [];
    if (!station) return [];

    return contracts
      .filter(
        (contract) =>
          contract.station_id === station.station_id &&
          OPEN_CONTRACT_STATUSES.has(String(contract.status || "").toLowerCase()),
      )
      .sort((left, right) => contractSortKey(right).localeCompare(contractSortKey(left)));
  }, [contractsQuery.data, station]);

  const myStationContracts = useMemo(() => {
    const contracts = contractsQuery.data || [];
    if (!station || !sessionUserId) return [];

    return contracts
      .filter(
        (contract) =>
          contract.station_id === station.station_id &&
          Number(contract.creator_user_id || 0) === sessionUserId,
      )
      .sort((left, right) => contractSortKey(right).localeCompare(contractSortKey(left)));
  }, [contractsQuery.data, sessionUserId, station]);

  const activeLifecycleContracts = useMemo(
    () =>
      myStationContracts.filter((contract) =>
        ["draft", "pending_approval", "pending_collateral", "active"].includes(
          normalizeStatus(contract.status),
        ),
      ),
    [myStationContracts],
  );

  const createContractMutation = useMutation({
    mutationFn: async (draft) => {
      const accepted = await createContract(token, {
        name: draft.name,
        region: draft.region,
        metric: draft.metric,
        threshold: draft.threshold,
        multiplier: draft.multiplier,
        measurement_unit: draft.measurement_unit,
        trading_period_start: draft.trading_period_start,
        trading_period_end: draft.trading_period_end,
        measurement_period_start: draft.measurement_period_start,
        measurement_period_end: draft.measurement_period_end,
        data_provider_name: draft.data_provider_name,
        station_id: draft.station_id,
        data_provider_station_mode: draft.data_provider_station_mode,
        description: draft.description,
      });
      const command = await getContractCommand(token, accepted.command_id);
      return command;
    },
    onSuccess(command, draft) {
      const payoutCents = parsePositiveInt(draft.multiplier) || 100;
      const pairs = parsePositiveInt(draft.paired_quantity);
      const collateral = collateralForDraft(payoutCents, pairs);
      setFormError("");
      setFormSuccess(
        `Draft contract #${command.result_contract_id} created. Planned issuance: ${pairs || 0} pairs. Required collateral before activation: $${(collateral / 100).toFixed(2)}.`,
      );
      queryClient.invalidateQueries({ queryKey: ["contracts"] });
      if (command.result_contract_id) {
        setActivationDrafts((current) => ({
          ...current,
          [command.result_contract_id]: {
            pairedQuantity: pairs ? String(pairs) : "",
          },
        }));
      }
      setFormState(buildInitialForm(station));
    },
    onError(error) {
      setFormSuccess("");
      setFormError(validationErrorText(error));
    },
  });

  const lifecycleMutation = useMutation({
    mutationFn: async ({ contract, pairedQuantity }) => {
      let current = contract;
      const quantity = parsePositiveInt(pairedQuantity);
      if (!quantity) {
        throw new Error("paired quantity must be greater than 0");
      }

      const status = () => normalizeStatus(current?.status);

      if (status() === "draft") {
        current = await submitContractForApproval(token, current.id);
      }
      if (status() === "pending_approval") {
        current = await approveContract(token, current.id);
      }
      if (status() === "pending_collateral") {
        const issuanceRows = await listIssuanceBatches(token, current.id);
        const issuanceBatches = issuanceRows?.issuance_batches || issuanceRows || [];
        const hasIssuedSupply = issuanceBatches.some(
          (batch) => normalizeStatus(batch.status) === "issued",
        );

        if (!hasIssuedSupply) {
          const requirement = await getCollateralRequirement(token, current.id, quantity, "USD");
          const existingLock = (portfolioQuery.data?.collateral_locks || []).find(
            (lock) =>
              Number(lock.contract_id) === Number(current.id) &&
              normalizeStatus(lock.status) === "locked" &&
              Number(lock.amount_cents || 0) >= Number(requirement.required_amount_cents || 0) &&
              !lock.reference_issuance_id,
          );

          let lockId = existingLock?.id || null;
          if (!lockId) {
            const collateralLock = await lockContractCollateral(token, current.id, {
              paired_quantity: quantity,
              currency: "USD",
              correlation_id: `ui-activate-${current.id}-${Date.now()}`,
              description: "UI activation collateral",
            });
            lockId = collateralLock?.collateral_lock?.id;
          }

          if (!lockId) {
            throw new Error("failed to lock collateral");
          }

          await createIssuanceBatch(token, current.id, {
            collateral_lock_id: lockId,
            paired_quantity: quantity,
          });
        }

        current = await activateContract(token, current.id);
      }

      return current;
    },
    onSuccess(contract) {
      queryClient.invalidateQueries({ queryKey: ["contracts"] });
      queryClient.invalidateQueries({ queryKey: ["portfolio", token] });
      setFormSuccess(`Market #${contract.id} is now ${contract.status}.`);
      setFormError("");
    },
    onError(error) {
      setFormSuccess("");
      setFormError(validationErrorText(error));
    },
  });

  if (!station && stationsQuery.isLoading) {
    return (
      <div className="station-detail-page">
        <div className="station-empty panel">Loading station...</div>
      </div>
    );
  }

  if (!station && stationsQuery.isError) {
    return (
      <div className="station-detail-page">
        <div className="station-empty panel">{validationErrorText(stationsQuery.error)}</div>
      </div>
    );
  }

  if (!station) {
    return (
      <div className="station-detail-page">
        <div className="station-empty panel">Station not found.</div>
      </div>
    );
  }

  const handleFieldChange = (field, value) => {
    setFormState((current) => {
      const next = { ...current, [field]: value };
      if (field === "metric" && !current.measurementUnit) {
        next.measurementUnit = DEFAULT_METRIC_UNITS[value] || "";
      }
      return next;
    });
  };

  const handleSubmit = (event) => {
    event.preventDefault();
    setFormError("");
    setFormSuccess("");

    if (!isAuthenticated) {
      openLogin();
      return;
    }

    createContractMutation.mutate({
      name: formState.name,
      region: station.region || station.display_name,
      metric: formState.metric,
      threshold: formState.threshold ? Number(formState.threshold) : undefined,
      multiplier: formState.multiplier ? Number(formState.multiplier) : undefined,
      paired_quantity: parsePositiveInt(formState.pairedQuantity),
      measurement_unit: formState.measurementUnit,
      trading_period_start: formState.tradingPeriodStart,
      trading_period_end: formState.tradingPeriodEnd,
      measurement_period_start: formState.measurementPeriodStart,
      measurement_period_end: formState.measurementPeriodEnd,
      data_provider_name: station.provider_name,
      station_id: station.station_id,
      data_provider_station_mode: "single_station",
      description: formState.description,
    });
  };

  const payoutCents = parsePositiveInt(formState.multiplier) || 100;
  const pairedQuantity = parsePositiveInt(formState.pairedQuantity);
  const thresholdText = formState.threshold
    ? `${formState.threshold}${formState.measurementUnit || ""}`
    : "the threshold";
  const metricText = titleCase(formState.metric || "metric");
  const payoutText = `$${(payoutCents / 100).toFixed(2)}`;
  const requiredCollateralCents = collateralForDraft(payoutCents, pairedQuantity);
  const maxPayoutText = `$${(requiredCollateralCents / 100).toFixed(2)}`;
  const availableUsdBalance = portfolioQuery.data?.accounts?.find((account) => account.currency === "USD");

  return (
    <div className="station-detail-page">
      <div className="station-detail-page__backlink">
        <Link className="station-detail-page__backlink-anchor" to="/">
          All stations
        </Link>
      </div>

      <header className="station-detail-page__header panel">
        <div className="station-detail-page__header-main">
          <p className="station-directory-page__eyebrow">Station</p>
          <h1 className="station-detail-page__title">{station.display_name}</h1>
          <p className="station-detail-page__meta">
            {station.provider_name} | {station.station_id} | {station.region || "Region unavailable"}
          </p>
        </div>
        <div className="station-detail-page__chips">
          {(station.supported_metrics || []).map((metric) => (
            <span key={metric} className="station-card__metric">
              {titleCase(metric)}
            </span>
          ))}
        </div>
      </header>

      <section className="station-detail-page__content">
        <div className="station-detail-page__contracts panel">
          <div className="station-section__header">
            <div>
              <h2 className="station-section__title">Open contracts</h2>
              <p className="station-section__copy">
                Markets already tied to this station and still in an open lifecycle state.
              </p>
            </div>
          </div>

          {contractsQuery.isLoading ? <div className="station-empty">Loading contracts...</div> : null}
          {contractsQuery.isError ? <div className="station-empty">{validationErrorText(contractsQuery.error)}</div> : null}
          {!contractsQuery.isLoading && !contractsQuery.isError && openContracts.length === 0 ? (
            <div className="station-empty">No open contracts are linked to this station yet.</div>
          ) : null}

          <div className="station-contract-list">
            {openContracts.map((contract) => (
              <div key={contract.id} className="station-contract-card">
                <div className="station-contract-card__header">
                  <div>
                    <h3 className="station-contract-card__title">{contract.name}</h3>
                    <p className="station-contract-card__meta">
                      {titleCase(contract.metric)} | {contract.status}
                    </p>
                  </div>
                  <button
                    className="topbar__signup"
                    type="button"
                    onClick={() => navigate(`/contracts/${contract.id}`)}
                  >
                    Open market
                  </button>
                </div>
                <p className="station-contract-card__description">
                  {contract.description || "No description provided."}
                </p>
                <div className="station-contract-card__grid">
                  <span>Threshold: {contract.threshold ?? "-"}</span>
                  <span>Unit: {contract.measurement_unit || "-"}</span>
                  <span>Trading: {contract.trading_period_start || "-"} to {contract.trading_period_end || "-"}</span>
                  <span>Measurement: {contract.measurement_period_start || "-"} to {contract.measurement_period_end || "-"}</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="station-detail-page__form panel">
          <div className="station-section__header">
            <div>
              <h2 className="station-section__title">Create contract</h2>
              <p className="station-section__copy">
                Define the market terms for this specific station. Station, provider, and region are fixed from the selected station.
              </p>
            </div>
          </div>

          <form className="station-form" onSubmit={handleSubmit}>
            <div className="station-form__grid">
              <label className="station-form__field">
                <span className="station-form__label">Market name</span>
                <input
                  className="station-form__input"
                  type="text"
                  value={formState.name}
                  onChange={(event) => handleFieldChange("name", event.target.value)}
                  placeholder="Halifax daily max temperature above 75F"
                />
              </label>

              <label className="station-form__field">
                <span className="station-form__label">Metric</span>
                {(station.supported_metrics || []).length > 0 ? (
                  <select
                    className="station-form__input"
                    value={formState.metric}
                    onChange={(event) => handleFieldChange("metric", event.target.value)}
                  >
                    {(station.supported_metrics || []).map((metric) => (
                      <option key={metric} value={metric}>
                        {titleCase(metric)}
                      </option>
                    ))}
                  </select>
                ) : (
                  <input
                    className="station-form__input"
                    type="text"
                    value={formState.metric}
                    onChange={(event) => handleFieldChange("metric", event.target.value)}
                    placeholder="temperature_max"
                  />
                )}
              </label>

              <label className="station-form__field">
                <span className="station-form__label">Threshold</span>
                <input
                  className="station-form__input"
                  type="number"
                  step="1"
                  value={formState.threshold}
                  onChange={(event) => handleFieldChange("threshold", event.target.value)}
                  placeholder="75"
                />
              </label>

              <label className="station-form__field">
                <span className="station-form__label">Measurement unit</span>
                <input
                  className="station-form__input"
                  type="text"
                  value={formState.measurementUnit}
                  onChange={(event) => handleFieldChange("measurementUnit", event.target.value)}
                  placeholder="F"
                />
              </label>

              <label className="station-form__field">
                <span className="station-form__label">Payout per winning claim (cents)</span>
                <input
                  className="station-form__input"
                  type="number"
                  step="1"
                  value={formState.multiplier}
                  onChange={(event) => handleFieldChange("multiplier", event.target.value)}
                  placeholder="100"
                />
              </label>

              <label className="station-form__field">
                <span className="station-form__label">Planned pair issuance</span>
                <input
                  className="station-form__input"
                  type="number"
                  step="1"
                  min="1"
                  value={formState.pairedQuantity}
                  onChange={(event) => handleFieldChange("pairedQuantity", event.target.value)}
                  placeholder="1000"
                />
              </label>

              <label className="station-form__field">
                <span className="station-form__label">Provider</span>
                <input className="station-form__input station-form__input--readonly" type="text" value={station.provider_name} readOnly />
              </label>

              <label className="station-form__field">
                <span className="station-form__label">Station code</span>
                <input className="station-form__input station-form__input--readonly" type="text" value={station.station_id} readOnly />
              </label>

              <label className="station-form__field">
                <span className="station-form__label">Region</span>
                <input className="station-form__input station-form__input--readonly" type="text" value={station.region || station.display_name} readOnly />
              </label>

              <label className="station-form__field">
                <span className="station-form__label">Trading period start</span>
                <input
                  className="station-form__input"
                  type="date"
                  value={formState.tradingPeriodStart}
                  onChange={(event) => handleFieldChange("tradingPeriodStart", event.target.value)}
                />
              </label>

              <label className="station-form__field">
                <span className="station-form__label">Trading period end</span>
                <input
                  className="station-form__input"
                  type="date"
                  value={formState.tradingPeriodEnd}
                  onChange={(event) => handleFieldChange("tradingPeriodEnd", event.target.value)}
                />
              </label>

              <label className="station-form__field">
                <span className="station-form__label">Measurement period start</span>
                <input
                  className="station-form__input"
                  type="date"
                  value={formState.measurementPeriodStart}
                  onChange={(event) => handleFieldChange("measurementPeriodStart", event.target.value)}
                />
              </label>

              <label className="station-form__field">
                <span className="station-form__label">Measurement period end</span>
                <input
                  className="station-form__input"
                  type="date"
                  value={formState.measurementPeriodEnd}
                  onChange={(event) => handleFieldChange("measurementPeriodEnd", event.target.value)}
                />
              </label>
            </div>

            <label className="station-form__field">
              <span className="station-form__label">Description</span>
              <textarea
                className="station-form__textarea"
                value={formState.description}
                onChange={(event) => handleFieldChange("description", event.target.value)}
                placeholder="Describe what the contract resolves on and why this market exists."
              />
            </label>

            <section className="station-definition panel" aria-label="Contract definition preview">
              <div className="station-section__header">
                <div>
                  <h3 className="station-section__title">Resolution and payout preview</h3>
                  <p className="station-section__copy">
                    This makes the settlement rule explicit before the draft is created.
                  </p>
                </div>
              </div>

              <div className="station-definition__grid">
                <div className="station-definition__card">
                  <div className="station-definition__label">If outcome resolves above threshold</div>
                  <div className="station-definition__value">
                    ABOVE pays {payoutText} per claim when {metricText} is above {thresholdText}.
                  </div>
                </div>

                <div className="station-definition__card">
                  <div className="station-definition__label">If outcome resolves at or below threshold</div>
                  <div className="station-definition__value">
                    BELOW pays {payoutText} per claim when {metricText} is at or below {thresholdText}.
                  </div>
                </div>

                <div className="station-definition__card">
                  <div className="station-definition__label">Planned issuance</div>
                  <div className="station-definition__value">
                    {pairedQuantity || 0} above claims and {pairedQuantity || 0} below claims.
                  </div>
                </div>

              <div className="station-definition__card">
                <div className="station-definition__label">Required collateral</div>
                <div className="station-definition__value">
                    Creators must post at least $1,000.00 in collateral before activation. Current requirement for this setup: {maxPayoutText}.
                  </div>
                </div>
              </div>

              <p className="station-definition__note">
                The backend creates the draft first. You can then progress it to active below by posting collateral, issuing supply, and activating the market.
              </p>
            </section>

            {formError ? <div className="station-form__error">{formError}</div> : null}
            {formSuccess ? <div className="station-form__success">{formSuccess}</div> : null}

            <div className="station-form__actions">
              <button className="topbar__signup" type="submit" disabled={createContractMutation.isPending}>
                {createContractMutation.isPending ? "Creating contract..." : "Create contract"}
              </button>
            </div>
          </form>
        </div>

        <div className="station-detail-page__creator panel">
          <div className="station-section__header">
            <div>
              <h2 className="station-section__title">My market drafts</h2>
              <p className="station-section__copy">
                Only the creator can move a market from draft to active. Activation will post at least $1,000.00 in collateral, issue paired inventory, and then open the market.
              </p>
            </div>
          </div>

          {isAuthenticated ? (
            <div className="station-creator__balance">
              Available USD balance: {availableUsdBalance ? formatMoneyCents(availableUsdBalance.available_cents) : "$0.00"}
            </div>
          ) : (
            <div className="station-empty">Log in to manage your station-backed markets.</div>
          )}

          {isAuthenticated && activeLifecycleContracts.length === 0 ? (
            <div className="station-empty">You have no draft or activatable markets for this station yet.</div>
          ) : null}

          <div className="station-contract-list">
            {activeLifecycleContracts.map((contract) => {
              const activationDraft = activationDrafts[contract.id] || { pairedQuantity: "" };
              const draftPairs = parsePositiveInt(activationDraft.pairedQuantity);
              const draftPayoutCents = Number(contract.multiplier || 100);
              const draftRequiredCollateral = collateralForDraft(draftPayoutCents, draftPairs || 1);
              const contractStatus = normalizeStatus(contract.status);
              const canRunLifecycle = ["draft", "pending_approval", "pending_collateral"].includes(contractStatus);

              return (
                <div key={`creator-${contract.id}`} className="station-contract-card station-contract-card--creator">
                  <div className="station-contract-card__header">
                    <div>
                      <h3 className="station-contract-card__title">{contract.name}</h3>
                      <p className="station-contract-card__meta">
                        {titleCase(contract.metric)} | {contract.status}
                      </p>
                    </div>
                    <button
                      className="topbar__signup"
                      type="button"
                      onClick={() => navigate(`/contracts/${contract.id}`)}
                    >
                      Open market
                    </button>
                  </div>

                  <div className="station-contract-card__grid">
                    <span>Threshold: {contract.threshold ?? "-"}</span>
                    <span>Unit: {contract.measurement_unit || "-"}</span>
                    <span>Provider: {contract.data_provider_name || "-"}</span>
                    <span>Station: {contract.station_id || "-"}</span>
                  </div>

                  {canRunLifecycle ? (
                    <div className="station-creator__workflow">
                      <label className="station-form__field">
                        <span className="station-form__label">Pair issuance for activation</span>
                        <input
                          className="station-form__input"
                          type="number"
                          min="1"
                          step="1"
                          value={activationDraft.pairedQuantity}
                          onChange={(event) =>
                            setActivationDrafts((current) => ({
                              ...current,
                              [contract.id]: {
                                ...current[contract.id],
                                pairedQuantity: event.target.value,
                              },
                            }))
                          }
                          placeholder="1000"
                        />
                      </label>

                      <div className="station-definition__card">
                        <div className="station-definition__label">Activation requirement</div>
                        <div className="station-definition__value">
                          Required collateral: {formatMoneyCents(draftRequiredCollateral)}.
                          {draftRequiredCollateral === 100000
                            ? " The $1,000.00 minimum floor applies here."
                            : " This exceeds the minimum floor."}
                        </div>
                      </div>

                      <div className="station-form__actions station-form__actions--stacked">
                        <button
                          className="topbar__signup"
                          type="button"
                          disabled={lifecycleMutation.isPending || !draftPairs}
                          onClick={() =>
                            lifecycleMutation.mutate({
                              contract,
                              pairedQuantity: draftPairs,
                            })
                          }
                        >
                          {lifecycleMutation.isPending ? "Activating..." : "Confirm and activate market"}
                        </button>
                      </div>
                    </div>
                  ) : contractStatus === "active" ? (
                    <div className="station-form__success">This market is already active.</div>
                  ) : null}
                </div>
              );
            })}
          </div>
        </div>
      </section>
    </div>
  );
}

import { OrderBookPanel } from "./OrderBookPanel";
import { useMemo, useState } from "react";
import { formatMoneyCents, formatPrice, formatQuantity } from "../lib/formatters";
import {
  Box,
  Button,
  FormControl,
  InputAdornment,
  InputLabel,
  MenuItem,
  Select,
  Tab,
  Tabs,
  TextField,
} from "@mui/material";

function actionLabel(action) {
  return action === "bid" ? "Buy" : "Sell";
}

function commandTone(command) {
  const status = String(command?.status || "").toLowerCase();
  if (status === "failed") return "negative";
  if (status === "succeeded" || status === "completed") return "positive";
  return "muted";
}

export function TradeRail({
  contract,
  marketState,
  portfolio,
  token,
  onRequireLogin,
  onSubmitOrder,
  orderCommand,
  orderPending,
  orderError,
  orderBookView,
  onOrderBookViewChange,
}) {
  const [action, setAction] = useState("bid");
  const [belief, setBelief] = useState("above");
  const [price, setPrice] = useState("");
  const [quantity, setQuantity] = useState("");
  const [mintQuantity, setMintQuantity] = useState("");

  const summary = marketState?.summary;
  const bestQuote =
    action === "bid"
      ? summary?.best?.[belief]?.ask
      : summary?.best?.[belief]?.bid;

  const account = portfolio?.accounts?.[0] || null;
  const position = (portfolio?.positions || []).find(
    (row) =>
      Number(row.contract_id) === Number(contract?.id) &&
      String(row.side || "").toLowerCase() === belief,
  ) || null;
  const availableInventory = Number(position?.available_quantity || 0);
  const requestedQuantity = Number(quantity || 0);
  const hasSession = Boolean(token);
  const submitLabel = `Submit ${actionLabel(action)} ${belief === "above" ? "Above" : "Below"}`;
  const canSellRequestedQuantity =
    action !== "ask" || (requestedQuantity > 0 && availableInventory >= requestedQuantity);
  const clientError =
    action === "ask" && !position
      ? `You do not hold any ${belief} inventory for this market.`
      : action === "ask" && requestedQuantity > availableInventory
        ? `You only have ${formatQuantity(availableInventory)} ${belief} claims available to sell.`
        : "";

  const commandMessage = useMemo(() => {
    if (!orderCommand) return null;
    if (orderCommand.error_message) return orderCommand.error_message;
    if (orderCommand.status === "succeeded") {
      return `Processed order ${orderCommand.result_order_id || ""}`.trim();
    }
    return `Status: ${orderCommand.status}`;
  }, [orderCommand]);

  return (
    <aside className="panel--right-rail" aria-label="Market order entry">
      <section
        className="right-rail-section right-rail-section--order"
        aria-labelledby="place-order-heading"
      >
        <div className="right-rail-section__header">
          <h2 className="right-rail-section__title" id="place-order-heading">
            Trade
          </h2>
        </div>

        <div className="right-rail-section__body">
          <div className="trade-panel__body">
            <div className="trade-panel__top-row">
              <Tabs
                value={action}
                onChange={(_event, value) => setAction(value)}
                aria-label="Order action"
                variant="fullWidth"
                className="trade-tabs"
              >
                <Tab label="Buy" value="bid" />
                <Tab label="Sell" value="ask" />
              </Tabs>
            </div>

            <div className="order-entry-form">
              <div className="order-entry-form__inputs">
                <Box className="trade-form-row">
                  <div className="trade-form-row__label">Price</div>
                  <TextField
                    id="order-price"
                    type="number"
                    placeholder="0.00"
                    value={price}
                    onChange={(event) => setPrice(event.target.value)}
                    fullWidth
                    variant="outlined"
                    size="small"
                    InputProps={{
                      startAdornment: <InputAdornment position="start">$</InputAdornment>,
                    }}
                    inputProps={{ step: "0.01", min: "0", inputMode: "decimal" }}
                    className="trade-mui-field"
                  />
                </Box>

                <Box className="trade-form-row">
                  <div className="trade-form-row__label">Quantity</div>
                  <Box className="trade-mui-row">
                    <TextField
                      id="order-quantity"
                      type="number"
                      placeholder="0"
                      value={quantity}
                      onChange={(event) => setQuantity(event.target.value)}
                      fullWidth
                      variant="outlined"
                      size="small"
                      inputProps={{ step: "1", min: "0", inputMode: "numeric" }}
                      className="trade-mui-field"
                    />

                    <FormControl size="small" className="trade-mui-select">
                      <InputLabel id="order-belief-label">Side</InputLabel>
                      <Select
                        labelId="order-belief-label"
                        id="order-belief"
                        value={belief}
                        label="Side"
                        onChange={(event) => setBelief(event.target.value)}
                      >
                        <MenuItem value="above">Above</MenuItem>
                        <MenuItem value="below">Below</MenuItem>
                      </Select>
                    </FormControl>
                  </Box>
                </Box>
              </div>

              <Button
                id="order-submit"
                type="button"
                variant="contained"
                disabled={orderPending || !contract || (action === "ask" && !canSellRequestedQuantity)}
                onClick={() => {
                  if (!contract) {
                    return;
                  }

                  if (!hasSession) {
                    onRequireLogin();
                    return;
                  }

                  if (action === "ask" && !canSellRequestedQuantity) {
                    return;
                  }

                  onSubmitOrder({
                    contract_id: contract.id,
                    token_type: belief,
                    order_side: action,
                    price: Number(price),
                    quantity: Number(quantity),
                  });
                }}
              >
                {orderPending ? "Submitting..." : submitLabel}
              </Button>

              <div className="order-preview">
                <p className="order-preview__title">Order Preview</p>
                <div className="order-preview__rows">
                  <div className="order-preview__row">
                    <span className="order-preview__label">Action</span>
                    <span className="order-preview__value">
                      {actionLabel(action)} {belief === "above" ? "Above" : "Below"}
                    </span>
                  </div>
                  <div className="order-preview__row">
                    <span className="order-preview__label">Best quote</span>
                    <span className="order-preview__value">{formatPrice(bestQuote)}</span>
                  </div>
                  <div className="order-preview__row">
                    <span className="order-preview__label">Price</span>
                    <span className="order-preview__value">
                      {price ? formatPrice(Number(price)) : "-"}
                    </span>
                  </div>
                  <div className="order-preview__row">
                    <span className="order-preview__label">Quantity</span>
                    <span className="order-preview__value">
                      {quantity ? formatQuantity(Number(quantity)) : "-"}
                    </span>
                  </div>
                  {account ? (
                    <div className="order-preview__row">
                      <span className="order-preview__label">Available cash</span>
                      <span className="order-preview__value">
                        {formatMoneyCents(account.available_cents)}
                      </span>
                    </div>
                  ) : null}
                  {action === "ask" ? (
                    <div className="order-preview__row">
                      <span className="order-preview__label">
                        Available {belief === "above" ? "above" : "below"}
                      </span>
                      <span className="order-preview__value">
                        {formatQuantity(availableInventory)}
                      </span>
                    </div>
                  ) : null}
                  {commandMessage ? (
                    <div className="order-preview__row">
                      <span className="order-preview__label">Command</span>
                      <span className={`order-preview__value order-preview__value--${commandTone(orderCommand)}`}>
                        {commandMessage}
                      </span>
                    </div>
                  ) : null}
                  {clientError ? (
                    <div className="order-preview__row">
                      <span className="order-preview__label">Inventory</span>
                      <span className="order-preview__value order-preview__value--negative">
                        {clientError}
                      </span>
                    </div>
                  ) : null}
                  {orderError ? (
                    <div className="order-preview__row">
                      <span className="order-preview__label">Error</span>
                      <span className="order-preview__value order-preview__value--negative">
                        {orderError}
                      </span>
                    </div>
                  ) : null}
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section
        className="right-rail-section right-rail-section--mint"
        aria-labelledby="mint-heading"
      >
        <div className="right-rail-section__header">
          <h2 className="right-rail-section__title" id="mint-heading">
            Provide Liquidity
          </h2>
        </div>

        <div className="right-rail-section__body">
          <div className="mint-panel">
            <div className="mint-panel__form">
              <label className="mint-panel__field" htmlFor="mint-quantity">
                <span className="mint-panel__label">Quantity</span>
                <input
                  id="mint-quantity"
                  type="number"
                  placeholder="0"
                  value={mintQuantity}
                  onChange={(event) => setMintQuantity(event.target.value)}
                />
              </label>
              <button id="mint-submit" type="button" disabled>
                Issue {mintQuantity || 0} Pairs
              </button>
              <p className="rail-note">
                Issuance now follows the exchange-core approval and collateral flow.
              </p>
            </div>
          </div>
        </div>
      </section>

      <OrderBookPanel
        marketState={marketState}
        view={orderBookView}
        onViewChange={onOrderBookViewChange}
      />
    </aside>
  );
}

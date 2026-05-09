import { useMemo } from "react";
import { FormControl, MenuItem, Select } from "@mui/material";
import { buildOrderBookLadder } from "../lib/orderbook";
import { formatPrice, formatQuantity } from "../lib/formatters";

function formatSpreadPercent(value) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? `${parsed.toFixed(2)}%` : "-";
}

function renderTotal(value) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed.toFixed(2) : "-";
}

export function OrderBookPanel({ marketState, view, onViewChange }) {
  const ladder = useMemo(
    () => buildOrderBookLadder(marketState?.order_book, view),
    [marketState, view],
  );

  return (
    <section
      className="right-rail-section right-rail-section--book"
      aria-labelledby="order-book-heading"
    >
      <div className="right-rail-section__header right-rail-section__header--with-control">
        <h2 className="right-rail-section__title" id="order-book-heading">
          Order Book
        </h2>

        <FormControl size="small" className="order-book-view-select">
          <Select
            id="order-book-view"
            aria-label="Order book view"
            value={view}
            onChange={(event) => onViewChange(event.target.value)}
          >
            <MenuItem value="above">Above</MenuItem>
            <MenuItem value="below">Below</MenuItem>
          </Select>
        </FormControl>
      </div>

      <div className="right-rail-section__body">
        <div className="order-book">
          <div className="order-book__header">
            <span>Price</span>
            <span>Size</span>
            <span>Total</span>
          </div>

          <div id="order-book-list">
            {ladder.asks.map((level) => (
              <div
                key={level.key}
                className={`order-book__row order-book__row--${level.side}`}
                style={{ "--bar-width": `${level.barWidth}%` }}
              >
                <span className="order-book__cell order-book__cell--price">
                  {formatPrice(level.price)}
                </span>
                <span className="order-book__cell order-book__cell--size">
                  {formatQuantity(level.quantity)}
                </span>
                <span className="order-book__cell order-book__cell--total">
                  {renderTotal(level.total)}
                </span>
              </div>
            ))}

            <div className="order-book__spread" data-orderbook-spread-row="">
              <span className="order-book__spread-label">Spread</span>
              <span className="order-book__spread-value">
                {formatPrice(ladder.spread.spread)}
              </span>
              <span className="order-book__spread-percent">
                {formatSpreadPercent(ladder.spread.percent)}
              </span>
            </div>

            {ladder.bids.map((level) => (
              <div
                key={level.key}
                className={`order-book__row order-book__row--${level.side}`}
                style={{ "--bar-width": `${level.barWidth}%` }}
              >
                <span className="order-book__cell order-book__cell--price">
                  {formatPrice(level.price)}
                </span>
                <span className="order-book__cell order-book__cell--size">
                  {formatQuantity(level.quantity)}
                </span>
                <span className="order-book__cell order-book__cell--total">
                  {renderTotal(level.total)}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

function formatActionLabel(action) {
  return action === "bid" ? "Buy" : "Sell";
}

function renderTradeControls({ state, elements }) {
  const { actionOptions, beliefSelect, submitButton } = elements.trade;

  actionOptions.forEach((el) => {
    el.classList.toggle(
      "action-selector__option--active",
      el.dataset.action === state.ui.action,
    );
  });

  if (beliefSelect) {
    beliefSelect.value = state.ui.belief;
  }

  if (submitButton) {
    const beliefLabel = state.ui.belief === "above" ? "Above" : "Below";
    const actionLabel = formatActionLabel(state.ui.action);
    submitButton.textContent = `Submit ${actionLabel} ${beliefLabel}`;
  }
}

function renderOrderPreview({ state, elements }) {
  const {
    previewSide,
    previewQuote,
    previewPrice,
    previewQuantity,
    priceInput,
    quantityInput,
  } = elements.trade;

  if (previewSide) {
    previewSide.textContent = state.ui.belief === "above" ? "Above" : "Below";
  }

  if (previewQuote) {
    previewQuote.textContent = formatActionLabel(state.ui.action);
  }

  if (previewPrice) {
    previewPrice.textContent = priceInput?.value ? priceInput.value : "-";
  }

  if (previewQuantity) {
    previewQuantity.textContent = quantityInput?.value ? quantityInput.value : "-";
  }
}

export function renderTradePanel({ state, elements }) {
  const { actionSelector, beliefSelect, submitButton } = elements.trade;

  if (!actionSelector || !beliefSelect || !submitButton) return;

  renderTradeControls({ state, elements });
  renderOrderPreview({ state, elements });
}
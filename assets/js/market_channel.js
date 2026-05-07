import { Socket } from "../../deps/phoenix/assets/js/phoenix";

const MARKET_STATE_UPDATED = "market_state_updated";

function validMarketStatePayload(payload) {
  return (
    payload &&
    payload.type === MARKET_STATE_UPDATED &&
    payload.market_state &&
    payload.contract_id != null &&
    payload.sequence !== undefined &&
    payload.as_of !== undefined
  );
}

export function shouldApplyMarketUpdate(currentMarket, payload) {
  const incomingSequence = payload?.sequence;
  const currentSequence = currentMarket?.sequence;

  if (incomingSequence == null || currentSequence == null) return true;

  return Number(incomingSequence) > Number(currentSequence);
}

export function connectMarketChannel({
  contractId,
  getCurrentMarket,
  onMarketState,
  onReconnect,
}) {
  const socket = new Socket("/socket");
  const channel = socket.channel(`contract:${contractId}`, {});
  let joinedOnce = false;
  let disconnectedAfterJoin = false;

  socket.onClose(() => {
    if (joinedOnce) disconnectedAfterJoin = true;
  });

  socket.connect();

  channel.on(MARKET_STATE_UPDATED, (payload) => {
    if (!validMarketStatePayload(payload)) return;
    if (!shouldApplyMarketUpdate(getCurrentMarket(), payload)) return;

    onMarketState(payload);
  });

  channel
    .join()
    .receive("ok", async () => {
      if (joinedOnce && disconnectedAfterJoin) {
        disconnectedAfterJoin = false;
        await onReconnect();
      }

      joinedOnce = true;
    })
    .receive("error", (response) => {
      console.error("Market channel join failed", response);
    });

  return {
    disconnect() {
      channel.leave();
      socket.disconnect();
    },
  };
}

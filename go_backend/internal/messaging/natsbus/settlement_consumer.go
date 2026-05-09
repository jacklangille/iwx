package natsbus

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/nats-io/nats.go"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/requestctx"
	"iwx/go_backend/internal/settlement"
	"iwx/go_backend/pkg/logging"
)

type SettlementConsumer struct {
	conn    *nats.Conn
	js      nats.JetStreamContext
	subject string
	stream  string
	durable string
	service *settlement.Service
}

func NewSettlementConsumer(cfg config.Config, service *settlement.Service) (*SettlementConsumer, error) {
	conn, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return nil, err
	}
	js, err := conn.JetStream()
	if err != nil {
		_ = conn.Drain()
		return nil, err
	}

	return &SettlementConsumer{
		conn:    conn,
		js:      js,
		subject: cfg.NATSContractResolvedSubject,
		stream:  cfg.NATSContractResolvedStream,
		durable: cfg.SettlementInstanceID + "-contract-resolved",
		service: service,
	}, nil
}

func (c *SettlementConsumer) Run(ctx context.Context) error {
	sub, err := c.js.PullSubscribe(
		c.subject,
		c.durable,
		nats.BindStream(c.stream),
		nats.ManualAck(),
	)
	if err != nil {
		return err
	}
	logging.Info(context.Background(), "settlement_subscribed", "subject", c.subject, "durable", c.durable)

	for {
		select {
		case <-ctx.Done():
			_ = sub.Drain()
			return nil
		default:
		}

		msgs, err := sub.Fetch(1, nats.MaxWait(1*time.Second))
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				continue
			}
			logging.Error(ctx, "settlement_fetch_failed", err)
			time.Sleep(250 * time.Millisecond)
			continue
		}

		for _, msg := range msgs {
			if c.processMessage(ctx, msg) {
				_ = msg.Ack()
			} else {
				_ = msg.Nak()
			}
		}
	}
}

func (c *SettlementConsumer) processMessage(ctx context.Context, msg *nats.Msg) bool {
	var event events.ContractResolved
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		logging.Error(ctx, "settlement_parse_failed", err)
		return true
	}
	msgCtx := requestctx.WithTraceID(ctx, event.TraceID)
	if event.ContractID <= 0 {
		logging.Info(msgCtx, "settlement_invalid_event", "contract_id", event.ContractID)
		return true
	}

	result, err := c.service.HandleContractResolved(msgCtx, event)
	if err != nil {
		logging.Error(msgCtx, "settlement_failed", err, "contract_id", event.ContractID, "outcome", event.Outcome)
		return false
	}

	logging.Info(msgCtx, "settlement_completed", "contract_id", event.ContractID, "outcome", event.Outcome, "affected_users", result.AffectedUsers, "entries", len(result.Entries), "status", result.Contract.Status)
	return true
}

func (c *SettlementConsumer) Close() {
	if c == nil || c.conn == nil {
		return
	}
	logging.Info(context.Background(), "settlement_consumer_draining")
	_ = c.conn.Drain()
}

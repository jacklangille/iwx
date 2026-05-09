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
	"iwx/go_backend/pkg/logging"
)

type ExecutionEventHandler interface {
	HandleExecutionCreated(ctx context.Context, event events.ExecutionCreated) error
}

type ExecutionConsumer struct {
	conn    *nats.Conn
	js      nats.JetStreamContext
	subject string
	stream  string
	durable string
	handler ExecutionEventHandler
}

func NewExecutionConsumer(cfg config.Config, durable string, handler ExecutionEventHandler) (*ExecutionConsumer, error) {
	conn, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return nil, err
	}
	js, err := conn.JetStream()
	if err != nil {
		_ = conn.Drain()
		return nil, err
	}

	return &ExecutionConsumer{
		conn:    conn,
		js:      js,
		subject: cfg.NATSExecutionCreatedSubject,
		stream:  cfg.NATSExecutionCreatedStream,
		durable: durable,
		handler: handler,
	}, nil
}

func (c *ExecutionConsumer) Run(ctx context.Context) error {
	sub, err := c.js.PullSubscribe(c.subject, c.durable, nats.BindStream(c.stream), nats.ManualAck())
	if err != nil {
		return err
	}
	logging.Info(context.Background(), "execution_subscribed", "subject", c.subject, "durable", c.durable)

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
			logging.Error(ctx, "execution_fetch_failed", err, "durable", c.durable)
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

func (c *ExecutionConsumer) processMessage(ctx context.Context, msg *nats.Msg) bool {
	var event events.ExecutionCreated
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		logging.Error(ctx, "execution_parse_failed", err)
		return true
	}
	msgCtx := requestctx.WithTraceID(ctx, event.TraceID)

	if err := c.handler.HandleExecutionCreated(msgCtx, event); err != nil {
		logging.Error(msgCtx, "execution_handler_failed", err, "execution_id", event.ExecutionID, "contract_id", event.ContractID)
		return false
	}

	logging.Info(msgCtx, "execution_handler_succeeded", "execution_id", event.ExecutionID, "contract_id", event.ContractID)
	return true
}

func (c *ExecutionConsumer) Close() {
	if c == nil || c.conn == nil {
		return
	}
	logging.Info(context.Background(), "execution_consumer_draining")
	_ = c.conn.Drain()
}

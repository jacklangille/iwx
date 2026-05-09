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

type ProjectionChangeHandler interface {
	HandleProjectionChange(ctx context.Context, event events.ProjectionChange) error
}

type ProjectionChangeConsumer struct {
	conn    *nats.Conn
	js      nats.JetStreamContext
	subject string
	stream  string
	durable string
	handler ProjectionChangeHandler
}

func NewProjectionChangeConsumer(cfg config.Config, durable string, handler ProjectionChangeHandler) (*ProjectionChangeConsumer, error) {
	conn, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return nil, err
	}
	js, err := conn.JetStream()
	if err != nil {
		_ = conn.Drain()
		return nil, err
	}

	return &ProjectionChangeConsumer{
		conn:    conn,
		js:      js,
		subject: cfg.NATSProjectionChangeSubject,
		stream:  cfg.NATSProjectionChangeStream,
		durable: durable,
		handler: handler,
	}, nil
}

func (c *ProjectionChangeConsumer) Run(ctx context.Context) error {
	sub, err := c.js.PullSubscribe(c.subject, c.durable, nats.BindStream(c.stream), nats.ManualAck())
	if err != nil {
		return err
	}
	logging.Info(context.Background(), "projection_change_subscribed", "subject", c.subject, "durable", c.durable)

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
			logging.Error(ctx, "projection_change_fetch_failed", err, "durable", c.durable)
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

func (c *ProjectionChangeConsumer) processMessage(ctx context.Context, msg *nats.Msg) bool {
	var event events.ProjectionChange
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		logging.Error(ctx, "projection_change_decode_failed", err)
		return true
	}

	msgCtx := requestctx.WithTraceID(ctx, event.TraceID)
	if err := c.handler.HandleProjectionChange(msgCtx, event); err != nil {
		logging.Error(msgCtx, "projection_change_handle_failed", err, "event_id", event.EventID, "kind", event.Kind)
		return false
	}

	logging.Info(msgCtx, "projection_change_handle_succeeded", "event_id", event.EventID, "kind", event.Kind)
	return true
}

func (c *ProjectionChangeConsumer) Close() {
	if c == nil || c.conn == nil {
		return
	}
	logging.Info(context.Background(), "projection_change_consumer_draining")
	_ = c.conn.Drain()
}

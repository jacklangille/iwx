package natsbus

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/events"
	"iwx/go_backend/pkg/logging"
)

type ExecutionPublisher struct {
	conn    *nats.Conn
	js      nats.JetStreamContext
	subject string
}

func NewExecutionPublisher(cfg config.Config) (*ExecutionPublisher, error) {
	conn, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return nil, err
	}

	js, err := conn.JetStream()
	if err != nil {
		_ = conn.Drain()
		return nil, err
	}

	if err := ensureExecutionCreatedStream(js, cfg.NATSExecutionCreatedStream, cfg.NATSExecutionCreatedSubject); err != nil {
		_ = conn.Drain()
		return nil, err
	}

	logging.Info(context.Background(), "execution_publisher_ready", "stream", cfg.NATSExecutionCreatedStream, "subject", cfg.NATSExecutionCreatedSubject)
	return &ExecutionPublisher{
		conn:    conn,
		js:      js,
		subject: cfg.NATSExecutionCreatedSubject,
	}, nil
}

func (p *ExecutionPublisher) PublishExecutionCreated(ctx context.Context, event events.ExecutionCreated) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = p.js.Publish(p.subject, body, nats.MsgId(event.ExecutionID))
	if err == nil {
		logging.Info(ctx, "execution_created_published", "execution_id", event.ExecutionID, "contract_id", event.ContractID, "subject", p.subject)
	}
	return err
}

func (p *ExecutionPublisher) Close() {
	if p == nil || p.conn == nil {
		return
	}
	logging.Info(context.Background(), "execution_publisher_draining")
	_ = p.conn.Drain()
}

func ensureExecutionCreatedStream(js nats.JetStreamContext, streamName, subject string) error {
	_, err := js.StreamInfo(streamName)
	if err == nil {
		return nil
	}
	if err != nats.ErrStreamNotFound {
		return err
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:      streamName,
		Subjects:  []string{subject},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
		Discard:   nats.DiscardOld,
		MaxAge:    30 * 24 * time.Hour,
	})
	return err
}

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

type ProjectionChangePublisher struct {
	conn    *nats.Conn
	js      nats.JetStreamContext
	subject string
	stream  string
}

func NewProjectionChangePublisher(cfg config.Config) (*ProjectionChangePublisher, error) {
	conn, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return nil, err
	}

	js, err := conn.JetStream()
	if err != nil {
		_ = conn.Drain()
		return nil, err
	}

	if err := ensureProjectionChangeStream(js, cfg.NATSProjectionChangeStream, cfg.NATSProjectionChangeSubject); err != nil {
		_ = conn.Drain()
		return nil, err
	}

	logging.Info(context.Background(), "projection_change_publisher_ready", "stream", cfg.NATSProjectionChangeStream, "subject", cfg.NATSProjectionChangeSubject)
	return &ProjectionChangePublisher{
		conn:    conn,
		js:      js,
		subject: cfg.NATSProjectionChangeSubject,
		stream:  cfg.NATSProjectionChangeStream,
	}, nil
}

func (p *ProjectionChangePublisher) PublishProjectionChange(ctx context.Context, event events.ProjectionChange) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = p.js.Publish(p.subject, body, nats.MsgId(event.EventID))
	if err == nil {
		logging.Info(ctx, "projection_change_published", "event_id", event.EventID, "kind", event.Kind, "contract_id", event.ContractID, "user_id", event.UserID, "command_id", event.CommandID, "subject", p.subject)
	}
	return err
}

func (p *ProjectionChangePublisher) Close() {
	if p == nil || p.conn == nil {
		return
	}
	logging.Info(context.Background(), "projection_change_publisher_draining")
	_ = p.conn.Drain()
}

func ensureProjectionChangeStream(js nats.JetStreamContext, streamName, subject string) error {
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

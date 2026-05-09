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

type SettlementPublisher struct {
	conn    *nats.Conn
	js      nats.JetStreamContext
	subject string
}

func NewSettlementPublisher(cfg config.Config) (*SettlementPublisher, error) {
	conn, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return nil, err
	}

	js, err := conn.JetStream()
	if err != nil {
		_ = conn.Drain()
		return nil, err
	}

	if err := ensureSettlementCompletedStream(js, cfg.NATSSettlementCompletedStream, cfg.NATSSettlementCompletedSubject); err != nil {
		_ = conn.Drain()
		return nil, err
	}

	logging.Info(context.Background(), "settlement_publisher_ready", "stream", cfg.NATSSettlementCompletedStream, "subject", cfg.NATSSettlementCompletedSubject)

	return &SettlementPublisher{
		conn:    conn,
		js:      js,
		subject: cfg.NATSSettlementCompletedSubject,
	}, nil
}

func (p *SettlementPublisher) PublishSettlementCompleted(ctx context.Context, event events.SettlementCompleted) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = p.js.Publish(p.subject, body, nats.MsgId(event.EventID))
	if err == nil {
		logging.Info(ctx, "settlement_completed_published", "contract_id", event.ContractID, "subject", p.subject)
	}
	return err
}

func (p *SettlementPublisher) Close() {
	if p == nil || p.conn == nil {
		return
	}
	logging.Info(context.Background(), "settlement_publisher_draining")
	_ = p.conn.Drain()
}

func ensureSettlementCompletedStream(js nats.JetStreamContext, streamName, subject string) error {
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

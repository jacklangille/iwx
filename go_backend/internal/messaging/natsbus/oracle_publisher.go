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

type OraclePublisher struct {
	conn    *nats.Conn
	js      nats.JetStreamContext
	subject string
	stream  string
}

func NewOraclePublisher(cfg config.Config) (*OraclePublisher, error) {
	conn, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return nil, err
	}

	js, err := conn.JetStream()
	if err != nil {
		_ = conn.Drain()
		return nil, err
	}

	if err := ensureOracleResolvedStream(js, cfg.NATSContractResolvedStream, cfg.NATSContractResolvedSubject); err != nil {
		_ = conn.Drain()
		return nil, err
	}

	logging.Info(context.Background(), "oracle_publisher_ready", "stream", cfg.NATSContractResolvedStream, "subject", cfg.NATSContractResolvedSubject)

	return &OraclePublisher{
		conn:    conn,
		js:      js,
		subject: cfg.NATSContractResolvedSubject,
		stream:  cfg.NATSContractResolvedStream,
	}, nil
}

func (p *OraclePublisher) PublishContractResolved(ctx context.Context, event events.ContractResolved) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = p.js.Publish(p.subject, body)
	if err == nil {
		logging.Info(ctx, "oracle_contract_resolved_published", "contract_id", event.ContractID, "outcome", event.Outcome, "subject", p.subject)
	}
	return err
}

func (p *OraclePublisher) Close() {
	if p == nil || p.conn == nil {
		return
	}

	logging.Info(context.Background(), "oracle_publisher_draining")
	_ = p.conn.Drain()
}

func ensureOracleResolvedStream(js nats.JetStreamContext, streamName, subject string) error {
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

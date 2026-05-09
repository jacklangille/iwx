package natsbus

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/messaging"
)

type Client struct {
	conn             *nats.Conn
	js               nats.JetStreamContext
	subjects         messaging.Subjects
	timeout          time.Duration
	placeOrderStream string
}

func newClient(cfg config.Config) (*Client, error) {
	conn, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return nil, err
	}
	log.Printf("nats connected url=%s", cfg.NATSURL)

	js, err := conn.JetStream()
	if err != nil {
		_ = conn.Drain()
		return nil, err
	}

	subjects := messaging.Subjects{
		PlaceOrderBase: cfg.NATSPlaceOrderSubject,
		PartitionCount: cfg.NATSPartitionCount,
	}
	if err := ensurePlaceOrderStream(js, cfg.NATSPlaceOrderStream, subjects.PlaceOrderForPartitionWildcard()); err != nil {
		_ = conn.Drain()
		return nil, err
	}
	log.Printf(
		"nats stream ready stream=%s subject=%s partitions=%d",
		cfg.NATSPlaceOrderStream,
		subjects.PlaceOrderForPartitionWildcard(),
		subjects.PartitionCount,
	)

	return &Client{
		conn:             conn,
		js:               js,
		subjects:         subjects,
		timeout:          5 * time.Second,
		placeOrderStream: cfg.NATSPlaceOrderStream,
	}, nil
}

func (c *Client) Close() {
	if c == nil || c.conn == nil {
		return
	}

	log.Printf("nats draining connection")
	c.conn.Drain()
}

func ensurePlaceOrderStream(js nats.JetStreamContext, streamName, subject string) error {
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
		Retention: nats.WorkQueuePolicy,
		Discard:   nats.DiscardOld,
		MaxAge:    7 * 24 * time.Hour,
	})
	return err
}

package natsbus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/matching"
	"iwx/go_backend/internal/messaging"
	"iwx/go_backend/internal/requestctx"
	"iwx/go_backend/pkg/logging"
)

type MatcherConsumer struct {
	*Client
	instanceID      string
	ownedPartitions []int
	handler         matching.Handler
	subscriptions   []*nats.Subscription
}

func NewMatcherConsumer(cfg config.Config, handler matching.Handler) (*MatcherConsumer, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}

	ownedPartitions := append([]int(nil), cfg.MatcherOwnedPartitions...)
	if len(ownedPartitions) == 0 {
		return nil, fmt.Errorf("matcher must own at least one partition")
	}

	return &MatcherConsumer{
		Client:          client,
		instanceID:      cfg.MatcherInstanceID,
		ownedPartitions: ownedPartitions,
		handler:         handler,
		subscriptions:   make([]*nats.Subscription, 0, len(ownedPartitions)),
	}, nil
}

func (c *MatcherConsumer) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	for _, partition := range c.ownedPartitions {
		if err := c.startPlaceOrderConsumer(ctx, &wg, partition); err != nil {
			return err
		}
	}

	<-ctx.Done()
	wg.Wait()

	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}

	return ctx.Err()
}

func (c *MatcherConsumer) startPlaceOrderConsumer(ctx context.Context, wg *sync.WaitGroup, partition int) error {
	subject := c.subjects.PlaceOrderForPartition(partition)
	durable := fmt.Sprintf("%s-place-order-partition-%d", c.instanceID, partition)
	subscription, err := c.js.PullSubscribe(
		subject,
		durable,
		nats.BindStream(c.placeOrderStream),
		nats.ManualAck(),
	)
	if err != nil {
		return err
	}

	logging.Info(context.Background(), "matcher_subscribed", "flow", "place_order", "partition", partition, "subject", subject, "durable", durable)
	c.subscriptions = append(c.subscriptions, subscription)
	wg.Add(1)

	go func() {
		defer wg.Done()
		c.runPlaceOrderConsumer(ctx, partition, subscription)
	}()

	return nil
}

func (c *MatcherConsumer) runPlaceOrderConsumer(ctx context.Context, partition int, sub *nats.Subscription) {
	for {
		select {
		case <-ctx.Done():
			_ = sub.Drain()
			return
		default:
		}

		msgs, err := sub.Fetch(1, nats.MaxWait(1*time.Second))
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				continue
			}
			logging.Error(ctx, "matcher_fetch_failed", err, "partition", partition)
			time.Sleep(250 * time.Millisecond)
			continue
		}

		for _, msg := range msgs {
			ack := c.processPlaceOrderMessage(ctx, partition, msg)
			if ack {
				_ = msg.Ack()
			} else {
				_ = msg.Nak()
			}
		}
	}
}

func (c *MatcherConsumer) processPlaceOrderMessage(ctx context.Context, partition int, msg *nats.Msg) bool {
	request, ack, err := c.parsePlaceOrderMessage(partition, msg.Data)
	if err != nil {
		logging.Error(ctx, "matcher_message_parse_failed", err, "partition", partition, "ack", ack)
		return ack
	}
	msgCtx := requestctx.WithTraceID(ctx, request.Envelope.TraceID)

	logging.Info(msgCtx, "matcher_command_received", "command_id", request.Envelope.CommandID, "contract_id", request.Envelope.Command.ContractID, "partition", partition)
	_, err = c.handler.HandlePlaceOrder(msgCtx, request.Envelope)
	if err != nil {
		var validationErr *matching.ValidationError
		if errors.As(err, &validationErr) {
			logging.Error(msgCtx, "matcher_command_validation_failed", err, "command_id", request.Envelope.CommandID, "contract_id", request.Envelope.Command.ContractID, "errors", validationErr.Errors)
			return true
		}

		logging.Error(msgCtx, "matcher_command_failed", err, "command_id", request.Envelope.CommandID, "contract_id", request.Envelope.Command.ContractID)
		return false
	}

	logging.Info(msgCtx, "matcher_command_succeeded", "command_id", request.Envelope.CommandID, "contract_id", request.Envelope.Command.ContractID)
	return true
}

func (c *MatcherConsumer) parsePlaceOrderMessage(
	partition int,
	body []byte,
) (messaging.PlaceOrderRequest, bool, error) {
	var request messaging.PlaceOrderRequest
	if err := json.Unmarshal(body, &request); err != nil {
		return messaging.PlaceOrderRequest{}, true, err
	}

	if expected := c.subjects.PlaceOrderPartition(request.Envelope.Command.ContractID); expected != partition {
		return messaging.PlaceOrderRequest{}, true, fmt.Errorf(
			"contract %d routed to wrong partition: got %d want %d",
			request.Envelope.Command.ContractID,
			partition,
			expected,
		)
	}

	return request, true, nil
}

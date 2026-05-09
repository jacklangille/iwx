package natsbus

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/messaging"
	"iwx/go_backend/pkg/logging"
)

type MatcherClient struct {
	*Client
}

func NewMatcherClient(cfg config.Config) (*MatcherClient, error) {
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}

	return &MatcherClient{Client: client}, nil
}

func (c *MatcherClient) SubmitPlaceOrder(ctx context.Context, envelope commands.PlaceOrderEnvelope) (commands.PlaceOrderAccepted, error) {
	requestBody, err := json.Marshal(messaging.PlaceOrderRequest{Envelope: envelope})
	if err != nil {
		return commands.PlaceOrderAccepted{}, err
	}

	partition := c.subjects.PlaceOrderPartition(envelope.Command.ContractID)
	subject := c.subjects.PlaceOrderForPartition(partition)
	if _, err := c.js.Publish(subject, requestBody, nats.MsgId(envelope.CommandID)); err != nil {
		return commands.PlaceOrderAccepted{}, err
	}

	logging.Info(ctx, "nats_publish_place_order", "command_id", envelope.CommandID, "contract_id", envelope.Command.ContractID, "partition", partition, "subject", subject)

	return commands.PlaceOrderAccepted{
		CommandID:  envelope.CommandID,
		ContractID: envelope.Command.ContractID,
		Partition:  partition,
		Status:     "queued",
		EnqueuedAt: envelope.EnqueuedAt,
	}, nil
}

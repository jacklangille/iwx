package settlement

import (
	"context"
	"strings"

	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/store"
)

type EventPublisher interface {
	PublishSettlementCompleted(ctx context.Context, event events.SettlementCompleted) error
	Close()
}

type Service struct {
	settler   Settler
	publisher EventPublisher
}

type Settler interface {
	SettleContract(ctx context.Context, input store.SettleContractInput) (*store.SettlementResult, error)
}

func NewService(settler Settler, publisher EventPublisher) *Service {
	return &Service{
		settler:   settler,
		publisher: publisher,
	}
}

func (s *Service) HandleContractResolved(ctx context.Context, event events.ContractResolved) (*store.SettlementResult, error) {
	correlationID := strings.TrimSpace(event.TraceID)
	if correlationID == "" {
		correlationID = strings.TrimSpace(event.EventID)
	}
	if correlationID == "" {
		correlationID = "contract-resolved-" + strings.TrimSpace(event.ResolvedAt.UTC().Format("20060102150405")) + "-" + strings.TrimSpace(event.Outcome)
	}

	result, err := s.settler.SettleContract(ctx, store.SettleContractInput{
		ContractID:    event.ContractID,
		EventID:       event.EventID,
		Outcome:       strings.ToLower(strings.TrimSpace(event.Outcome)),
		ResolvedAt:    event.ResolvedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		CorrelationID: correlationID,
	})
	if err != nil {
		return nil, err
	}

	if s.publisher != nil {
		if err := s.publisher.PublishSettlementCompleted(ctx, events.SettlementCompleted{
			EventID:    "settlement-completed:" + event.EventID,
			ContractID: event.ContractID,
			TraceID:    event.TraceID,
			SettledAt:  result.SettledAt,
		}); err != nil {
			return nil, err
		}
	}

	return result, nil
}

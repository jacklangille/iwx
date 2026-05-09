package settlement

import (
	"context"
	"strings"

	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/readprojection"
	"iwx/go_backend/internal/store"
)

type EventPublisher interface {
	PublishSettlementCompleted(ctx context.Context, event events.SettlementCompleted) error
	Close()
}

type Service struct {
	repo      store.ExchangeCoreRepository
	projector *readprojection.Projector
	publisher EventPublisher
}

func NewService(repo store.ExchangeCoreRepository, projector *readprojection.Projector, publisher EventPublisher) *Service {
	return &Service{
		repo:      repo,
		projector: projector,
		publisher: publisher,
	}
}

func (s *Service) HandleContractResolved(ctx context.Context, event events.ContractResolved) (*store.SettlementResult, error) {
	correlationID := strings.TrimSpace(event.TraceID)
	if correlationID == "" {
		correlationID = "contract-resolved-" + strings.TrimSpace(event.ResolvedAt.UTC().Format("20060102150405")) + "-" + strings.TrimSpace(event.Outcome)
	}

	result, err := s.repo.SettleContract(ctx, store.SettleContractInput{
		ContractID:    event.ContractID,
		Outcome:       strings.ToLower(strings.TrimSpace(event.Outcome)),
		ResolvedAt:    event.ResolvedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		CorrelationID: correlationID,
	})
	if err != nil {
		return nil, err
	}

	if s.projector != nil {
		if err := s.projector.ProjectContract(ctx, event.ContractID); err != nil {
			return nil, err
		}
		if err := s.projector.ProjectSettlementState(ctx, event.ContractID); err != nil {
			return nil, err
		}
		for _, userID := range result.AffectedUsers {
			if err := s.projector.ProjectUserState(ctx, userID); err != nil {
				return nil, err
			}
		}
	}

	if s.publisher != nil {
		if err := s.publisher.PublishSettlementCompleted(ctx, events.SettlementCompleted{
			ContractID: event.ContractID,
			TraceID:    event.TraceID,
			SettledAt:  result.SettledAt,
		}); err != nil {
			return nil, err
		}
	}

	return result, nil
}

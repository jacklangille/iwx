package exchangecore

import (
	"context"

	"iwx/go_backend/internal/events"
)

func (s *Service) ApplyExecution(ctx context.Context, event events.ExecutionCreated) error {
	result, err := s.repo.ApplyExecution(ctx, event)
	if err != nil {
		return err
	}

	for _, userID := range result.AffectedUsers {
		if s.emitter != nil {
			if err := s.emitter.EmitUserStateChanged(ctx, userID, result.AppliedAt); err != nil {
				return err
			}
			continue
		}
		if err := s.projectUser(ctx, userID); err != nil {
			return err
		}
	}

	return nil
}

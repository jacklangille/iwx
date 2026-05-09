package app

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/matcherhttp"
	"iwx/go_backend/internal/matching"
	"iwx/go_backend/internal/messaging/natsbus"
	"iwx/go_backend/internal/outbox"
	"iwx/go_backend/internal/projectionchange"
	"iwx/go_backend/internal/store/postgres"
)

func RunMatcherService() error {
	config.LoadDotEnv()
	cfg := config.FromEnv()
	if err := cfg.ValidateForMatcher(); err != nil {
		return err
	}
	log.Printf(
		"matcher config loaded nats=%s db=%s instance=%s partitions=%v",
		cfg.NATSURL,
		cfg.MatcherDatabaseURL,
		cfg.MatcherInstanceID,
		cfg.MatcherOwnedPartitions,
	)

	if err := runStartupMigrations(context.Background(), "matcher", cfg, "matcher"); err != nil {
		return err
	}

	matchingRepo := postgres.NewMatchingRepository(cfg.MatcherDatabaseURL)
	orderRepo := postgres.NewOrderRepository(cfg.MatcherDatabaseURL)
	executionRepo := postgres.NewExecutionRepository(cfg.MatcherDatabaseURL)
	snapshotRepo := postgres.NewSnapshotRepository(cfg.MatcherDatabaseURL)
	outboxRepo := postgres.NewOutboxRepository(cfg.MatcherDatabaseURL)
	service := matching.NewService(matchingRepo)
	publisher, err := natsbus.NewExecutionPublisher(cfg)
	if err != nil {
		return err
	}
	defer publisher.Close()

	changeBusPublisher, err := natsbus.NewProjectionChangePublisher(cfg)
	if err != nil {
		return err
	}
	defer changeBusPublisher.Close()

	emitter := projectionchange.NewEmitter(outbox.NewProjectionChangePublisher(outboxRepo))
	handler := matching.NewProjectionChangePublishingHandler(service, emitter, matchingRepo)
	handler = matching.NewExecutionEventPublishingHandler(handler, outbox.NewExecutionCreatedPublisher(outboxRepo))

	consumer, err := natsbus.NewMatcherConsumer(cfg, handler)
	if err != nil {
		return err
	}
	defer func() {
		log.Printf("matcher shutting down nats consumer")
		consumer.Close()
	}()

	dispatcher := outbox.NewDispatcher(outboxRepo, 500*time.Millisecond, func(ctx context.Context, pending outbox.Event) error {
		switch pending.EventType {
		case outbox.EventTypeExecutionCreated:
			var event events.ExecutionCreated
			if err := json.Unmarshal(pending.Payload, &event); err != nil {
				return err
			}
			return publisher.PublishExecutionCreated(ctx, event)
		case outbox.EventTypeProjectionChange:
			var event events.ProjectionChange
			if err := json.Unmarshal(pending.Payload, &event); err != nil {
				return err
			}
			return changeBusPublisher.PublishProjectionChange(ctx, event)
		default:
			return nil
		}
	})
	go func() {
		if err := dispatcher.Run(context.Background()); err != nil {
			log.Printf("matcher outbox dispatcher stopped err=%v", err)
		}
	}()

	httpServer := matcherhttp.NewServer(cfg, orderRepo, executionRepo, snapshotRepo, matchingRepo)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Printf("matcher http server stopped err=%v", err)
		}
	}()

	log.Printf("matcher service ready instance=%s", cfg.MatcherInstanceID)

	return consumer.Run(context.Background())
}

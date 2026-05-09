package app

import (
	"context"
	"log"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/matching"
	"iwx/go_backend/internal/messaging/natsbus"
	"iwx/go_backend/internal/readprojection"
	"iwx/go_backend/internal/store/postgres"
)

func RunMatcherService() error {
	config.LoadDotEnv()
	cfg := config.FromEnv()
	if err := cfg.ValidateForMatcher(); err != nil {
		return err
	}
	log.Printf(
		"matcher config loaded nats=%s db=%s read_db=%s instance=%s partitions=%v",
		cfg.NATSURL,
		cfg.MatcherDatabaseURL,
		cfg.ReadDatabaseURL,
		cfg.MatcherInstanceID,
		cfg.MatcherOwnedPartitions,
	)

	if err := runStartupMigrations(context.Background(), "matcher", cfg, "matcher", "read"); err != nil {
		return err
	}

	matchingRepo := postgres.NewMatchingRepository(cfg.MatcherDatabaseURL)
	service := matching.NewService(matchingRepo)
	sourceOrderRepo := postgres.NewOrderRepository(cfg.MatcherDatabaseURL)
	sourceExecutionRepo := postgres.NewExecutionRepository(cfg.MatcherDatabaseURL)
	sourceSnapshotRepo := postgres.NewSnapshotRepository(cfg.MatcherDatabaseURL)
	readOrderRepo := postgres.NewOrderRepository(cfg.ReadDatabaseURL)
	readExecutionRepo := postgres.NewExecutionRepository(cfg.ReadDatabaseURL)
	readSnapshotRepo := postgres.NewSnapshotRepository(cfg.ReadDatabaseURL)
	readCommandRepo := postgres.NewMatchingRepository(cfg.ReadDatabaseURL)
	projector := readprojection.NewProjector(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		sourceOrderRepo,
		readOrderRepo,
		sourceExecutionRepo,
		readExecutionRepo,
		sourceSnapshotRepo,
		readSnapshotRepo,
		matchingRepo,
		readCommandRepo,
	)
	handler := matching.NewProjectingHandler(service, projector)

	publisher, err := natsbus.NewExecutionPublisher(cfg)
	if err != nil {
		return err
	}
	defer publisher.Close()
	handler = matching.NewExecutionEventPublishingHandler(handler, publisher)

	consumer, err := natsbus.NewMatcherConsumer(cfg, handler)
	if err != nil {
		return err
	}
	defer func() {
		log.Printf("matcher shutting down nats consumer")
		consumer.Close()
	}()

	log.Printf("matcher service ready instance=%s", cfg.MatcherInstanceID)

	return consumer.Run(context.Background())
}

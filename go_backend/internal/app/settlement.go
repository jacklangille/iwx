package app

import (
	"context"
	"log"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/messaging/natsbus"
	"iwx/go_backend/internal/readprojection"
	"iwx/go_backend/internal/settlement"
	"iwx/go_backend/internal/store/postgres"
)

func RunSettlementService() error {
	config.LoadDotEnv()
	cfg := config.FromEnv()
	if err := cfg.ValidateForSettlement(); err != nil {
		return err
	}
	log.Printf("settlement config loaded exchange_core_db=%s read_db=%s nats=%s instance=%s", cfg.ExchangeCoreDatabaseURL, cfg.ReadDatabaseURL, cfg.NATSURL, cfg.SettlementInstanceID)

	if err := runStartupMigrations(context.Background(), "settlement", cfg, "exchange-core", "read"); err != nil {
		return err
	}

	repo := postgres.NewContractRepository(cfg.ExchangeCoreDatabaseURL)
	readRepo := postgres.NewContractRepository(cfg.ReadDatabaseURL)
	projector := readprojection.NewProjector(
		repo,
		readRepo,
		repo,
		readRepo,
		nil,
		nil,
		repo,
		readRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	publisher, err := natsbus.NewSettlementPublisher(cfg)
	if err != nil {
		return err
	}
	defer publisher.Close()

	service := settlement.NewService(repo, projector, publisher)
	consumer, err := natsbus.NewSettlementConsumer(cfg, service)
	if err != nil {
		return err
	}
	defer consumer.Close()

	log.Printf("settlement service ready instance=%s", cfg.SettlementInstanceID)
	return consumer.Run(context.Background())
}

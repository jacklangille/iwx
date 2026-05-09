package app

import (
	"context"
	"log"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/exchangecorehttpclient"
	"iwx/go_backend/internal/matcherhttpclient"
	"iwx/go_backend/internal/messaging/natsbus"
	"iwx/go_backend/internal/oraclestationhttp"
	"iwx/go_backend/internal/projector"
	"iwx/go_backend/internal/readprojection"
	"iwx/go_backend/internal/store/postgres"
)

func RunProjectorService() error {
	config.LoadDotEnv()
	cfg := config.FromEnv()
	if err := cfg.ValidateForProjector(); err != nil {
		return err
	}
	log.Printf("projector config loaded read_db=%s exchange_core=%s matcher=%s oracle=%s nats=%s", cfg.ReadDatabaseURL, cfg.ExchangeCoreServiceURL, cfg.MatcherServiceURL, cfg.OracleServiceURL, cfg.NATSURL)

	if err := runStartupMigrations(context.Background(), "projector", cfg, "read"); err != nil {
		return err
	}

	contractRepo := postgres.NewContractRepository(cfg.ReadDatabaseURL)
	oracleRepo := postgres.NewOracleRepository(cfg.ReadDatabaseURL)
	orderRepo := postgres.NewOrderRepository(cfg.ReadDatabaseURL)
	executionRepo := postgres.NewExecutionRepository(cfg.ReadDatabaseURL)
	snapshotRepo := postgres.NewSnapshotRepository(cfg.ReadDatabaseURL)
	commandRepo := postgres.NewMatchingRepository(cfg.ReadDatabaseURL)
	checkpointRepo := postgres.NewProjectionCheckpointRepository(cfg.ReadDatabaseURL)

	applier := readprojection.NewApplier(checkpointRepo, contractRepo, contractRepo, oracleRepo, contractRepo, orderRepo, executionRepo, snapshotRepo, commandRepo)
	service := projector.NewService(
		applier,
		exchangecorehttpclient.NewClient(cfg.ExchangeCoreServiceURL),
		matcherhttpclient.NewClient(cfg.MatcherServiceURL),
		oraclestationhttp.NewClient(cfg.OracleServiceURL),
		nil,
	)

	consumer, err := natsbus.NewProjectionChangeConsumer(cfg, cfg.ProjectorInstanceID+"-projection-change", service)
	if err != nil {
		return err
	}
	defer consumer.Close()

	log.Printf("projector service ready instance=%s", cfg.ProjectorInstanceID)
	return consumer.Run(context.Background())
}

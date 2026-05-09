package app

import (
	"context"
	"log"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/httpapi"
	"iwx/go_backend/internal/readmodel"
	"iwx/go_backend/internal/store/postgres"
)

func RunReadAPIService() error {
	config.LoadDotEnv()
	cfg := config.FromEnv()
	if err := cfg.ValidateForReadAPI(); err != nil {
		return err
	}
	log.Printf("api config loaded http_addr=%s read_db=%s nats=%s", cfg.HTTPAddr, cfg.ReadDatabaseURL, cfg.NATSURL)

	if err := runStartupMigrations(context.Background(), "read-api", cfg, "read"); err != nil {
		return err
	}

	contractRepo := postgres.NewContractRepository(cfg.ReadDatabaseURL)
	oracleRepo := postgres.NewOracleRepository(cfg.ReadDatabaseURL)
	orderRepo := postgres.NewOrderRepository(cfg.ReadDatabaseURL)
	executionRepo := postgres.NewExecutionRepository(cfg.ReadDatabaseURL)
	snapshotRepo := postgres.NewSnapshotRepository(cfg.ReadDatabaseURL)
	commandRepo := postgres.NewMatchingRepository(cfg.ReadDatabaseURL)
	readService := readmodel.NewService(contractRepo, contractRepo, oracleRepo, contractRepo, orderRepo, executionRepo, snapshotRepo, commandRepo)

	server := httpapi.NewServer(cfg, readService)

	log.Printf("api service ready http_addr=%s", cfg.HTTPAddr)

	return server.ListenAndServe(context.Background())
}

func RunAPIServer() error {
	return RunReadAPIService()
}

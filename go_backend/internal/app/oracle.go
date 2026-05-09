package app

import (
	"context"
	"log"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/messaging/natsbus"
	"iwx/go_backend/internal/oracle"
	"iwx/go_backend/internal/oraclehttp"
	"iwx/go_backend/internal/readprojection"
	"iwx/go_backend/internal/store/postgres"
)

func RunOracleService() error {
	config.LoadDotEnv()
	cfg := config.FromEnv()
	if err := cfg.ValidateForOracle(); err != nil {
		return err
	}
	log.Printf("oracle config loaded http_addr=%s db=%s exchange_core_db=%s read_db=%s nats=%s", cfg.OracleHTTPAddr, cfg.OracleDatabaseURL, cfg.ExchangeCoreDatabaseURL, cfg.ReadDatabaseURL, cfg.NATSURL)

	if err := runStartupMigrations(context.Background(), "oracle", cfg, "oracle", "exchange-core", "read"); err != nil {
		return err
	}

	oracleRepo := postgres.NewOracleRepository(cfg.OracleDatabaseURL)
	contractRepo := postgres.NewContractRepository(cfg.ExchangeCoreDatabaseURL)
	readContractRepo := postgres.NewContractRepository(cfg.ReadDatabaseURL)
	readOracleRepo := postgres.NewOracleRepository(cfg.ReadDatabaseURL)

	projector := readprojection.NewProjector(
		contractRepo,
		readContractRepo,
		nil,
		nil,
		oracleRepo,
		readOracleRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	publisher, err := natsbus.NewOraclePublisher(cfg)
	if err != nil {
		return err
	}
	defer publisher.Close()

	service := oracle.NewService(oracleRepo, contractRepo, projector, publisher)
	server := oraclehttp.NewServer(cfg, service)

	log.Printf("oracle service ready http_addr=%s", cfg.OracleHTTPAddr)
	return server.ListenAndServe()
}

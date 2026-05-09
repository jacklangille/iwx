package app

import (
	"context"
	"log"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/exchangecore"
	"iwx/go_backend/internal/exchangecorehttp"
	"iwx/go_backend/internal/messaging/natsbus"
	"iwx/go_backend/internal/readprojection"
	"iwx/go_backend/internal/store/postgres"
)

func RunExchangeCoreService() error {
	config.LoadDotEnv()
	cfg := config.FromEnv()
	if err := cfg.ValidateForExchangeCore(); err != nil {
		return err
	}
	log.Printf("exchange-core config loaded http_addr=%s db=%s read_db=%s", cfg.ExchangeCoreHTTPAddr, cfg.ExchangeCoreDatabaseURL, cfg.ReadDatabaseURL)

	if err := runStartupMigrations(context.Background(), "exchange-core", cfg, "exchange-core", "read"); err != nil {
		return err
	}

	repo := postgres.NewContractRepository(cfg.ExchangeCoreDatabaseURL)
	readRepo := postgres.NewContractRepository(cfg.ReadDatabaseURL)
	orderCommandRepo := postgres.NewMatchingRepository(cfg.ReadDatabaseURL)
	stationCatalog := postgres.NewOracleRepository(cfg.ReadDatabaseURL)
	projector := readprojection.NewProjector(repo, readRepo, repo, readRepo, nil, nil, repo, readRepo, nil, nil, nil, nil, nil, nil, nil, nil)
	service := exchangecore.NewService(repo, stationCatalog, projector)
	matcherClient, err := natsbus.NewMatcherClient(cfg)
	if err != nil {
		return err
	}
	defer matcherClient.Close()
	server := exchangecorehttp.NewServer(cfg, service, matcherClient, orderCommandRepo)

	executionConsumer, err := natsbus.NewExecutionConsumer(cfg, cfg.ExchangeCoreInstanceID+"-execution-created", &exchangeCoreExecutionHandler{service: service})
	if err != nil {
		return err
	}
	defer executionConsumer.Close()

	go func() {
		if err := executionConsumer.Run(context.Background()); err != nil {
			log.Printf("exchange-core execution consumer stopped err=%v", err)
		}
	}()

	log.Printf("exchange-core service ready http_addr=%s", cfg.ExchangeCoreHTTPAddr)
	return server.ListenAndServe()
}

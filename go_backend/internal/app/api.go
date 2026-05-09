package app

import (
	"context"
	"log"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/exchangecore"
	"iwx/go_backend/internal/httpapi"
	"iwx/go_backend/internal/messaging/natsbus"
	"iwx/go_backend/internal/readmodel"
	"iwx/go_backend/internal/readprojection"
	"iwx/go_backend/internal/realtime"
	"iwx/go_backend/internal/store/postgres"
)

func RunReadAPIService() error {
	config.LoadDotEnv()
	cfg := config.FromEnv()
	if err := cfg.ValidateForReadAPI(); err != nil {
		return err
	}
	log.Printf("api config loaded http_addr=%s read_db=%s nats=%s", cfg.HTTPAddr, cfg.ReadDatabaseURL, cfg.NATSURL)

	if err := runStartupMigrations(context.Background(), "read-api", cfg, "read", "exchange-core", "matcher"); err != nil {
		return err
	}

	contractRepo := postgres.NewContractRepository(cfg.ReadDatabaseURL)
	oracleRepo := postgres.NewOracleRepository(cfg.ReadDatabaseURL)
	orderRepo := postgres.NewOrderRepository(cfg.ReadDatabaseURL)
	executionRepo := postgres.NewExecutionRepository(cfg.ReadDatabaseURL)
	snapshotRepo := postgres.NewSnapshotRepository(cfg.ReadDatabaseURL)
	commandRepo := postgres.NewMatchingRepository(cfg.ReadDatabaseURL)
	readService := readmodel.NewService(contractRepo, contractRepo, oracleRepo, contractRepo, orderRepo, executionRepo, snapshotRepo, commandRepo)
	exchangeCoreRepo := postgres.NewContractRepository(cfg.ExchangeCoreDatabaseURL)
	readProjectionRepo := postgres.NewContractRepository(cfg.ReadDatabaseURL)
	stationCatalog := postgres.NewOracleRepository(cfg.ReadDatabaseURL)
	projector := readprojection.NewProjector(
		exchangeCoreRepo,
		readProjectionRepo,
		exchangeCoreRepo,
		readProjectionRepo,
		nil,
		nil,
		exchangeCoreRepo,
		readProjectionRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	exchangeCoreService := exchangecore.NewService(exchangeCoreRepo, stationCatalog, projector)

	matcherClient, err := natsbus.NewMatcherClient(cfg)
	if err != nil {
		return err
	}
	defer func() {
		log.Printf("api shutting down matcher client")
		matcherClient.Close()
	}()

	hub := realtime.NewHub()
	server := httpapi.NewServer(cfg, readService, exchangeCoreService, hub, matcherClient)

	sourceOrderRepo := postgres.NewOrderRepository(cfg.MatcherDatabaseURL)
	sourceExecutionRepo := postgres.NewExecutionRepository(cfg.MatcherDatabaseURL)
	sourceSnapshotRepo := postgres.NewSnapshotRepository(cfg.MatcherDatabaseURL)
	marketProjector := readprojection.NewProjector(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		sourceOrderRepo,
		orderRepo,
		sourceExecutionRepo,
		executionRepo,
		sourceSnapshotRepo,
		snapshotRepo,
		nil,
		nil,
	)

	executionConsumer, err := natsbus.NewExecutionConsumer(cfg, cfg.ReadAPIInstanceID+"-execution-created", &readAPIExecutionHandler{
		reads:     readService,
		projector: marketProjector,
	})
	if err != nil {
		return err
	}
	defer executionConsumer.Close()

	go func() {
		if err := executionConsumer.Run(context.Background()); err != nil {
			log.Printf("read-api execution consumer stopped err=%v", err)
		}
	}()

	log.Printf("api service ready http_addr=%s", cfg.HTTPAddr)

	return server.ListenAndServe(context.Background())
}

func RunAPIServer() error {
	return RunReadAPIService()
}

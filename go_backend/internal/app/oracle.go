package app

import (
	"context"
	"log"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/exchangecorehttpclient"
	"iwx/go_backend/internal/messaging/natsbus"
	"iwx/go_backend/internal/oracle"
	"iwx/go_backend/internal/oraclehttp"
	"iwx/go_backend/internal/outbox"
	"iwx/go_backend/internal/projectionchange"
	"iwx/go_backend/internal/store/postgres"
)

func RunOracleService() error {
	config.LoadDotEnv()
	cfg := config.FromEnv()
	if err := cfg.ValidateForOracle(); err != nil {
		return err
	}
	log.Printf("oracle config loaded http_addr=%s db=%s exchange_core_service=%s nats=%s", cfg.OracleHTTPAddr, cfg.OracleDatabaseURL, cfg.ExchangeCoreServiceURL, cfg.NATSURL)

	if err := runStartupMigrations(context.Background(), "oracle", cfg, "oracle"); err != nil {
		return err
	}

	oracleRepo := postgres.NewOracleRepository(cfg.OracleDatabaseURL)
	contractRepo := exchangecorehttpclient.NewClient(cfg.ExchangeCoreServiceURL)
	emitter := projectionchange.NewEmitter(outbox.NewProjectionChangePublisher(oracleRepo))

	publisher, err := natsbus.NewOraclePublisher(cfg)
	if err != nil {
		return err
	}
	defer publisher.Close()
	changeBusPublisher, err := natsbus.NewProjectionChangePublisher(cfg)
	if err != nil {
		return err
	}
	defer changeBusPublisher.Close()

	service := oracle.NewService(oracleRepo, contractRepo, emitter, publisher)
	server := oraclehttp.NewServer(cfg, service)
	dispatcher := oracle.NewOutboxDispatcher(oracleRepo, publisher, changeBusPublisher)

	go func() {
		if err := dispatcher.Run(context.Background()); err != nil {
			log.Printf("oracle outbox dispatcher stopped err=%v", err)
		}
	}()

	log.Printf("oracle service ready http_addr=%s", cfg.OracleHTTPAddr)
	return server.ListenAndServe()
}

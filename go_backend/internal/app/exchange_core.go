package app

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/exchangecore"
	"iwx/go_backend/internal/exchangecorehttp"
	"iwx/go_backend/internal/messaging/natsbus"
	"iwx/go_backend/internal/oraclestationhttp"
	"iwx/go_backend/internal/outbox"
	"iwx/go_backend/internal/projectionchange"
	"iwx/go_backend/internal/store/postgres"
)

func RunExchangeCoreService() error {
	config.LoadDotEnv()
	cfg := config.FromEnv()
	if err := cfg.ValidateForExchangeCore(); err != nil {
		return err
	}
	log.Printf("exchange-core config loaded http_addr=%s db=%s oracle_service=%s", cfg.ExchangeCoreHTTPAddr, cfg.ExchangeCoreDatabaseURL, cfg.OracleServiceURL)

	if err := runStartupMigrations(context.Background(), "exchange-core", cfg, "exchange-core"); err != nil {
		return err
	}

	repo := postgres.NewContractRepository(cfg.ExchangeCoreDatabaseURL)
	outboxRepo := postgres.NewOutboxRepository(cfg.ExchangeCoreDatabaseURL)
	stationCatalog := oraclestationhttp.NewClient(cfg.OracleServiceURL)
	changeBusPublisher, err := natsbus.NewProjectionChangePublisher(cfg)
	if err != nil {
		return err
	}
	defer changeBusPublisher.Close()

	emitter := projectionchange.NewEmitter(outbox.NewProjectionChangePublisher(outboxRepo))
	service := exchangecore.NewService(repo, stationCatalog, emitter)
	matcherClient, err := natsbus.NewMatcherClient(cfg)
	if err != nil {
		return err
	}
	defer matcherClient.Close()
	server := exchangecorehttp.NewServer(cfg, service, matcherClient)

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

	dispatcher := outbox.NewDispatcher(outboxRepo, 500*time.Millisecond, func(ctx context.Context, pending outbox.Event) error {
		switch pending.EventType {
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
			log.Printf("exchange-core outbox dispatcher stopped err=%v", err)
		}
	}()

	log.Printf("exchange-core service ready http_addr=%s", cfg.ExchangeCoreHTTPAddr)
	return server.ListenAndServe()
}

package app

import (
	"context"
	"log"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/exchangecorehttpclient"
	"iwx/go_backend/internal/messaging/natsbus"
	"iwx/go_backend/internal/settlement"
)

func RunSettlementService() error {
	config.LoadDotEnv()
	cfg := config.FromEnv()
	if err := cfg.ValidateForSettlement(); err != nil {
		return err
	}
	log.Printf("settlement config loaded exchange_core_service=%s nats=%s instance=%s", cfg.ExchangeCoreServiceURL, cfg.NATSURL, cfg.SettlementInstanceID)

	if err := runStartupMigrations(context.Background(), "settlement", cfg); err != nil {
		return err
	}

	publisher, err := natsbus.NewSettlementPublisher(cfg)
	if err != nil {
		return err
	}
	defer publisher.Close()

	settler := exchangecorehttpclient.NewClient(cfg.ExchangeCoreServiceURL)
	service := settlement.NewService(settler, publisher)
	consumer, err := natsbus.NewSettlementConsumer(cfg, service)
	if err != nil {
		return err
	}
	defer consumer.Close()

	log.Printf("settlement service ready instance=%s", cfg.SettlementInstanceID)
	return consumer.Run(context.Background())
}

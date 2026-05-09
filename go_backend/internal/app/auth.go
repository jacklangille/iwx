package app

import (
	"context"
	"log"
	"strings"

	"iwx/go_backend/internal/auth"
	"iwx/go_backend/internal/authhttp"
	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/store/postgres"
)

func RunAuthServer() error {
	config.LoadDotEnv()
	cfg := config.FromEnv()
	if err := cfg.ValidateForAuth(); err != nil {
		return err
	}
	log.Printf("auth config loaded http_addr=%s db=%s issuer=%s", cfg.AuthHTTPAddr, cfg.AuthDatabaseURL, cfg.AuthJWTIssuer)

	if err := runStartupMigrations(context.Background(), "auth", cfg, "auth"); err != nil {
		return err
	}

	userRepo := postgres.NewAuthUserRepository(cfg.AuthDatabaseURL)

	if err := bootstrapAuthUser(context.Background(), userRepo, cfg); err != nil {
		return err
	}

	authService, err := auth.NewService(userRepo, cfg)
	if err != nil {
		return err
	}

	server := authhttp.NewServer(cfg, authService)

	log.Printf("auth service ready http_addr=%s", cfg.AuthHTTPAddr)

	return server.ListenAndServe(context.Background())
}

func bootstrapAuthUser(ctx context.Context, userRepo *postgres.AuthUserRepository, cfg config.Config) error {
	if strings.TrimSpace(cfg.AuthBootstrapUsername) == "" || strings.TrimSpace(cfg.AuthBootstrapPassword) == "" {
		return nil
	}

	passwordHash, err := auth.HashPassword(cfg.AuthBootstrapPassword)
	if err != nil {
		return err
	}

	return userRepo.UpsertUser(ctx, cfg.AuthBootstrapUsername, passwordHash, true)
}

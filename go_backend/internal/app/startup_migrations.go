package app

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/migrate"
)

func runStartupMigrations(ctx context.Context, serviceName string, cfg config.Config, targetNames ...string) error {
	targets := migrationTargets(cfg)

	for _, targetName := range targetNames {
		target, ok := targets[targetName]
		if !ok {
			return fmt.Errorf("unknown startup migration target %q", targetName)
		}

		log.Printf("%s startup migration target=%s dir=%s", serviceName, target.Name, target.Dir)
		if err := migrate.Run(ctx, target); err != nil {
			return err
		}
	}

	return nil
}

func migrationTargets(cfg config.Config) map[string]migrate.Target {
	return map[string]migrate.Target{
		"auth": {
			Name:        "auth",
			DatabaseURL: cfg.AuthDatabaseURL,
			Dir:         filepath.Join("migrations", "auth"),
		},
		"exchange-core": {
			Name:        "exchange-core",
			DatabaseURL: cfg.ExchangeCoreDatabaseURL,
			Dir:         filepath.Join("migrations", "exchange-core"),
		},
		"matcher": {
			Name:        "matcher",
			DatabaseURL: cfg.MatcherDatabaseURL,
			Dir:         filepath.Join("migrations", "matcher"),
		},
		"oracle": {
			Name:        "oracle",
			DatabaseURL: cfg.OracleDatabaseURL,
			Dir:         filepath.Join("migrations", "oracle"),
		},
		"read": {
			Name:        "read",
			DatabaseURL: cfg.ReadDatabaseURL,
			Dir:         filepath.Join("migrations", "read"),
		},
	}
}

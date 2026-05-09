package main

import (
	"context"
	"flag"
	"log"
	"path/filepath"

	"iwx/go_backend/internal/config"
	"iwx/go_backend/internal/migrate"
)

func main() {
	targetName := flag.String("target", "all", "migration target: auth|exchange-core|matcher|oracle|read|all")
	flag.Parse()

	config.LoadDotEnv()
	cfg := config.FromEnv()

	targets := []migrate.Target{
		{Name: "auth", DatabaseURL: cfg.AuthDatabaseURL, Dir: filepath.Join("migrations", "auth")},
		{Name: "exchange-core", DatabaseURL: cfg.ExchangeCoreDatabaseURL, Dir: filepath.Join("migrations", "exchange-core")},
		{Name: "matcher", DatabaseURL: cfg.MatcherDatabaseURL, Dir: filepath.Join("migrations", "matcher")},
		{Name: "oracle", DatabaseURL: cfg.OracleDatabaseURL, Dir: filepath.Join("migrations", "oracle")},
		{Name: "read", DatabaseURL: cfg.ReadDatabaseURL, Dir: filepath.Join("migrations", "read")},
	}

	for _, target := range targets {
		if *targetName != "all" && *targetName != target.Name {
			continue
		}

		log.Printf("migrating target=%s dir=%s", target.Name, target.Dir)
		if err := migrate.Run(context.Background(), target); err != nil {
			log.Fatal(err)
		}
		log.Printf("migration complete target=%s", target.Name)
	}
}

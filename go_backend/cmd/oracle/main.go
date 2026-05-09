package main

import (
	"log"

	"iwx/go_backend/internal/app"
	"iwx/go_backend/pkg/logging"
)

func main() {
	closeLogger, err := logging.Setup("oracle", "cmd/oracle")
	if err != nil {
		log.Fatal(err)
	}
	defer closeLogger()

	if err := app.RunOracleService(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"log"

	"iwx/go_backend/internal/app"
	"iwx/go_backend/pkg/logging"
)

func main() {
	closeLogger, err := logging.Setup("exchange-core", "cmd/exchange-core")
	if err != nil {
		log.Fatal(err)
	}
	defer closeLogger()

	if err := app.RunExchangeCoreService(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"log"

	"iwx/go_backend/internal/app"
	"iwx/go_backend/pkg/logging"
)

func main() {
	closeLogger, err := logging.Setup("settlement", "cmd/settlement")
	if err != nil {
		log.Fatal(err)
	}
	defer closeLogger()

	if err := app.RunSettlementService(); err != nil {
		log.Fatal(err)
	}
}

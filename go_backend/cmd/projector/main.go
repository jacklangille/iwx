package main

import (
	"log"

	"iwx/go_backend/internal/app"
	"iwx/go_backend/pkg/logging"
)

func main() {
	closeLogger, err := logging.Setup("projector", "cmd/projector")
	if err != nil {
		log.Fatal(err)
	}
	defer closeLogger()

	if err := app.RunProjectorService(); err != nil {
		log.Fatal(err)
	}
}

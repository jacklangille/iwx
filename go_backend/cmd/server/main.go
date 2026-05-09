package main

import (
	"log"

	"iwx/go_backend/internal/app"
	"iwx/go_backend/pkg/logging"
)

func main() {
	closeLogger, err := logging.Setup("server", "cmd/server")
	if err != nil {
		log.Fatal(err)
	}
	defer closeLogger()

	if err := app.RunAPIServer(); err != nil {
		log.Fatal(err)
	}
}

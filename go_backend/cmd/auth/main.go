package main

import (
	"log"

	"iwx/go_backend/internal/app"
	"iwx/go_backend/pkg/logging"
)

func main() {
	closeLogger, err := logging.Setup("auth", "cmd/auth")
	if err != nil {
		log.Fatal(err)
	}
	defer closeLogger()

	if err := app.RunAuthServer(); err != nil {
		log.Fatal(err)
	}
}

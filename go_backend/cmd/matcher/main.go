package main

import (
	"log"

	"iwx/go_backend/internal/app"
	"iwx/go_backend/pkg/logging"
)

func main() {
	closeLogger, err := logging.Setup("matcher", "cmd/matcher")
	if err != nil {
		log.Fatal(err)
	}
	defer closeLogger()

	if err := app.RunMatcherService(); err != nil {
		log.Fatal(err)
	}
}

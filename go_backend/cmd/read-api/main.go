package main

import (
	"log"

	"iwx/go_backend/internal/app"
	"iwx/go_backend/pkg/logging"
)

func main() {
	closeLogger, err := logging.Setup("read-api", "cmd/read-api")
	if err != nil {
		log.Fatal(err)
	}
	defer closeLogger()

	if err := app.RunReadAPIService(); err != nil {
		log.Fatal(err)
	}
}

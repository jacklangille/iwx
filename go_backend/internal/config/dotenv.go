package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadDotEnv() {
	candidates := []string{
		".env",
		filepath.Join("go_backend", ".env"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				log.Printf("warning: failed to load %s: %v", path, err)
			}
			return
		}
	}
}

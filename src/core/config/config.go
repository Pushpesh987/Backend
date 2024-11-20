package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func SetupEnv() {
	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

// Config returns the environment variable or defaults to empty string
func Config(key string) string {
	return os.Getenv(key)
}

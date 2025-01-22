package main

import (
	"log"

	"github.com/joho/godotenv"

	"tedalogger-logfetcher/internal/logfetcher"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading .env: %v", err)
	}

	logfetcher.StartManager()
}

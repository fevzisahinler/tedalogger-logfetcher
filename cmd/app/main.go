// cmd/app/main.go

package main

import (
	"log"
	"tedalogger-logfetcher/internal/logfetcher"

	"tedalogger-logfetcher/config"
	"tedalogger-logfetcher/internal/logexporter"
)

func main() {
	_, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config yüklenemedi: %v", err)
	}

	go logfetcher.StartManager()
	go logexporter.StartDailyJobScheduler()

	select {}
}

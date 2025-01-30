// config/config.go

package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RabbitMQURL string

	APIBaseURL  string
	APIUsername string
	APIPassword string

	ElasticURL  string
	ElasticUser string
	ElasticPass string
}

var cfg *Config

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg = &Config{
		RabbitMQURL: getEnv("RABBITMQ_URL", ""),

		APIBaseURL:  getEnv("API_BASE_URL", ""),
		APIUsername: getEnv("API_USERNAME", ""),
		APIPassword: getEnv("API_PASSWORD", ""),

		ElasticURL:  getEnv("ELASTIC_URL", ""),
		ElasticUser: getEnv("ELASTIC_USER", ""),
		ElasticPass: getEnv("ELASTIC_PASS", ""),
	}
	return cfg, nil
}

func GetConfig() *Config {
	return cfg
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

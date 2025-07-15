package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	PostgresURL  string
	KafkaBrokers []string
	KafkaTopic   string
	HTTPPort     string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, fmt.Errorf("failed to load env file")
	}
	cfg := &Config{
		PostgresURL:  os.Getenv("POSTGRES_URL"),
		KafkaBrokers: []string{os.Getenv("KAFKA_BROKER")},
		KafkaTopic:   os.Getenv("KAFKA_TOPIC"),
		HTTPPort:     os.Getenv("HTTP_PORT"),
	}
	// Значения по умолчанию, если нет .env
	if cfg.PostgresURL == "" {
		log.Println("Postgres_url not set, using default")
		cfg.PostgresURL = "postgres://order_user:securepassword@postgres:5432/orders_db?sslmode=disable"
	}
	if len(cfg.KafkaBrokers) == 1 && cfg.KafkaBrokers[0] == "" {
		log.Println("kafka_broker not set, using defualt")
		cfg.KafkaBrokers = []string{"kafka:9092"}
	}
	if cfg.KafkaTopic == "" {
		log.Println("kafka_topic not set, using default")
		cfg.KafkaTopic = "orders"
	}
	if cfg.HTTPPort == "" {
		log.Println("http_port not set, using default")
		cfg.HTTPPort = "8081"
	}

	return cfg, nil
}

package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	Token string
}

func LoadConfig() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	return &Config{
		Token: token,
	}
}

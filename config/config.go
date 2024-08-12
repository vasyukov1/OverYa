package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Config struct {
	Token           string
	AdminID         int64
	TelegramChannel int64
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

	adminIDStr := os.Getenv("ADMIN_ID")
	if adminIDStr == "" {
		log.Fatal("ADMIN_ID is not set")
	}
	adminID, err := strconv.ParseInt(adminIDStr, 10, 64)
	if err != nil {
		log.Fatal("Invalid ADMIN_ID")
	}

	telegramChannelStr := os.Getenv("TELEGRAM_CHANNEL")
	if telegramChannelStr == "" {
		log.Fatal("TELEGRAM_CHANNEL is not set")
	}
	telegramChannel, err := strconv.ParseInt(telegramChannelStr, 10, 64)
	if err != nil {
		log.Fatal("Invalid TELEGRAM_CHANNEL")
	}

	return &Config{
		Token:           token,
		AdminID:         adminID,
		TelegramChannel: telegramChannel,
	}
}

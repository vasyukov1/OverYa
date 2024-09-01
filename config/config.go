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
	Host            string
	Port            int64
	User            string
	Password        string
	Dbname          string
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

	host := os.Getenv("HOST")
	if host == "" {
		log.Fatal("HOST is not set")
	}

	portStr := os.Getenv("PORT")
	if portStr == "" {
		log.Fatal("PORT is not set")
	}
	port, err := strconv.ParseInt(portStr, 10, 64)
	if err != nil {
		log.Fatal("Invalid PORT")
	}

	user := os.Getenv("USER")
	if user == "" {
		log.Fatal("USER is not set")
	}

	password := os.Getenv("PASSWORD")
	if password == "" {
		log.Fatal("PASSWORD is not set")
	}

	dbname := os.Getenv("DBNAME")
	if dbname == "" {
		log.Fatal("DBNAME is not set")
	}

	return &Config{
		Token:           token,
		AdminID:         adminID,
		TelegramChannel: telegramChannel,
		Host:            host,
		Port:            port,
		User:            user,
		Password:        password,
		Dbname:          dbname,
	}
}

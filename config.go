package main

import (
	"os"
	"strconv"

	"github.com/ehsundar/social-content-dl/internal/tgapp"
)

type Config struct {
	Telegram tgapp.Config
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	// Load Telegram configuration
	config.Telegram.PhoneNumber = getEnv("TELEGRAM_PHONE", "")
	if config.Telegram.PhoneNumber == "" {
		config.Telegram.PhoneNumber = getEnv("PHONE_NUMBER", "")
	}

	appIDStr := getEnv("TELEGRAM_APP_ID", "17349")
	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		return nil, err
	}
	config.Telegram.AppID = appID

	config.Telegram.AppHash = getEnv("TELEGRAM_APP_HASH", "344583e45741c457fe1862106095a5eb")

	// Load Download configuration
	config.Telegram.DownloadPath = getEnv("DOWNLOAD_PATH", "./downloads")

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

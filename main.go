package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/ehsundar/social-content-dl/internal/tgapp"
)

func main() {
	fmt.Println("Social Content DL - Starting up...")

	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go telegram <channel_username> [limit]")
		fmt.Println("Example: go run main.go telegram musicchannel 50")
		fmt.Println("")
		fmt.Println("Environment variables:")
		fmt.Println("  TELEGRAM_PHONE - Your phone number (e.g., +1234567890)")
		fmt.Println("  TELEGRAM_APP_ID - Your app ID (optional, default: 17349)")
		fmt.Println("  TELEGRAM_APP_HASH - Your app hash (optional, default: 344583e45741c457fe1862106095a5eb)")
		fmt.Println("  DOWNLOAD_PATH - Download directory (default: ./downloads)")
		fmt.Println("")
		fmt.Println("Note: You'll need to enter the verification code sent to your phone on first run.")
		os.Exit(1)
	}

	platform := os.Args[1]
	channelUsername := os.Args[2]

	var limit int
	if len(os.Args) > 3 {
		var err error
		limit, err = strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatalf("Invalid limit: %v", err)
		}
	}

	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	switch platform {
	case "telegram":
		if err := downloadTelegramMusic(channelUsername, limit, config); err != nil {
			log.Fatalf("Error downloading from Telegram: %v", err)
		}
	default:
		log.Fatalf("Unsupported platform: %s", platform)
	}
}

func downloadTelegramMusic(channelUsername string, limit int, config *Config) error {
	if config.Telegram.PhoneNumber == "" {
		return fmt.Errorf("TELEGRAM_PHONE environment variable is required")
	}

	downloader, err := tgapp.NewTelegramDownloader(config.Telegram)
	if err != nil {
		return fmt.Errorf("failed to create downloader: %w", err)
	}

	ctx := context.Background()
	return downloader.DownloadChannelMusic(ctx, channelUsername, limit)
}

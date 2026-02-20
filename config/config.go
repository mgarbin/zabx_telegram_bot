// Package config loads bot configuration from environment variables.
package config

import (
	"errors"
	"os"
	"strconv"
)

// Config holds all runtime configuration values.
type Config struct {
	// TelegramToken is the bot token provided by BotFather.
	TelegramToken string

	// ChatID is the Telegram group chat ID the bot posts to.
	ChatID int64

	// ServerAddr is the address the HTTP server listens on (e.g. ":8080").
	ServerAddr string
}

// Load reads configuration from environment variables:
//   - TELEGRAM_BOT_TOKEN (required)
//   - TELEGRAM_CHAT_ID   (required, numeric)
//   - SERVER_ADDR        (optional, default ":8080")
func Load() (*Config, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, errors.New("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		return nil, errors.New("TELEGRAM_CHAT_ID environment variable is required")
	}
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return nil, errors.New("TELEGRAM_CHAT_ID must be a valid integer")
	}

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return &Config{
		TelegramToken: token,
		ChatID:        chatID,
		ServerAddr:    addr,
	}, nil
}

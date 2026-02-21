// Package config loads bot configuration from a YAML file and/or environment
// variables. Environment variables always take precedence over file values.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds all runtime configuration values.
type Config struct {
	// TelegramToken is the bot token provided by BotFather.
	TelegramToken string

	// ChatID is the Telegram group chat ID the bot posts to.
	ChatID int64

	// ServerAddr is the address the HTTP server listens on (e.g. ":8080").
	ServerAddr string

	// ServerSecret is an optional shared secret that must be present in the
	// JSON body of every incoming request. When empty, no secret check is done.
	ServerSecret string
}

// fileConfig mirrors the YAML structure of the optional config file.
type fileConfig struct {
	TelegramToken string `yaml:"telegram_bot_token"`
	ChatID        string `yaml:"telegram_chat_id"`
	ServerAddr    string `yaml:"server_addr"`
	ServerSecret  string `yaml:"server_secret"`
}

// Load reads configuration from an optional YAML file and environment variables.
// Environment variables always override values from the file.
//
// YAML file lookup:
//   - Path given by CONFIG_FILE env var (error if the file is missing).
//   - Falls back to "config.yaml" in the working directory (silently skipped if absent).
//
// Environment variables:
//   - TELEGRAM_BOT_TOKEN (required if not set in the file)
//   - TELEGRAM_CHAT_ID   (required if not set in the file, numeric)
//   - SERVER_ADDR        (optional, default ":8080")
//   - SERVER_SECRET      (optional, shared secret for incoming requests)
func Load() (*Config, error) {
	fc, err := loadFile()
	if err != nil {
		return nil, err
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		token = fc.TelegramToken
	}
	if token == "" {
		return nil, errors.New("TELEGRAM_BOT_TOKEN is required (env var or config file)")
	}

	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		chatIDStr = fc.ChatID
	}
	if chatIDStr == "" {
		return nil, errors.New("TELEGRAM_CHAT_ID is required (env var or config file)")
	}
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return nil, errors.New("TELEGRAM_CHAT_ID must be a valid integer")
	}

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = fc.ServerAddr
	}
	if addr == "" {
		addr = ":8080"
	}

	secret := os.Getenv("SERVER_SECRET")
	if secret == "" {
		secret = fc.ServerSecret
	}

	return &Config{
		TelegramToken: token,
		ChatID:        chatID,
		ServerAddr:    addr,
		ServerSecret:  secret,
	}, nil
}

// loadFile parses the YAML config file, if present.
func loadFile() (fileConfig, error) {
	path := os.Getenv("CONFIG_FILE")
	explicit := path != ""
	if !explicit {
		path = "config.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && !explicit {
			return fileConfig{}, nil // default file is optional
		}
		return fileConfig{}, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var fc fileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		return fileConfig{}, fmt.Errorf("parsing config file %q: %w", path, err)
	}
	return fc, nil
}

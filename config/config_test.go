package config_test

import (
	"os"
	"testing"

	"github.com/mgarbin/zabx_telegram_bot/config"
)

func TestLoadMissingToken(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("TELEGRAM_CHAT_ID")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when TELEGRAM_BOT_TOKEN is missing")
	}
}

func TestLoadMissingChatID(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Unsetenv("TELEGRAM_CHAT_ID")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when TELEGRAM_CHAT_ID is missing")
	}
}

func TestLoadInvalidChatID(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Setenv("TELEGRAM_CHAT_ID", "not-a-number")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when TELEGRAM_CHAT_ID is not numeric")
	}
}

func TestLoadSuccess(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "my-token")
	os.Setenv("TELEGRAM_CHAT_ID", "-100123456")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TelegramToken != "my-token" {
		t.Errorf("expected token 'my-token', got %q", cfg.TelegramToken)
	}
	if cfg.ChatID != -100123456 {
		t.Errorf("expected chat ID -100123456, got %d", cfg.ChatID)
	}
	if cfg.ServerAddr != ":8080" {
		t.Errorf("expected default addr ':8080', got %q", cfg.ServerAddr)
	}
}

func TestLoadCustomServerAddr(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "tok")
	os.Setenv("TELEGRAM_CHAT_ID", "1")
	os.Setenv("SERVER_ADDR", ":9090")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")
	defer os.Unsetenv("SERVER_ADDR")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ServerAddr != ":9090" {
		t.Errorf("expected addr ':9090', got %q", cfg.ServerAddr)
	}
}

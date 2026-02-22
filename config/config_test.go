package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mgarbin/zabbix-telegram-event-correlator/config"
)

// clearEnv unsets all env vars used by config.Load so tests are isolated.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"TELEGRAM_BOT_TOKEN", "TELEGRAM_CHAT_ID", "SERVER_ADDR", "SERVER_SECRET", "CONFIG_FILE"} {
		os.Unsetenv(key)
	}
}

func TestLoadMissingToken(t *testing.T) {
	clearEnv(t)

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when TELEGRAM_BOT_TOKEN is missing")
	}
}

func TestLoadMissingChatID(t *testing.T) {
	clearEnv(t)
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when TELEGRAM_CHAT_ID is missing")
	}
}

func TestLoadInvalidChatID(t *testing.T) {
	clearEnv(t)
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
	clearEnv(t)
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
	clearEnv(t)
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

// writeYAML writes content to a temp file and returns its path.
func writeYAML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoadFromYAMLFile(t *testing.T) {
	clearEnv(t)
	path := writeYAML(t, `
telegram_bot_token: "yaml-token"
telegram_chat_id: "-999888"
server_addr: ":7070"
`)
	os.Setenv("CONFIG_FILE", path)
	defer os.Unsetenv("CONFIG_FILE")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TelegramToken != "yaml-token" {
		t.Errorf("expected token 'yaml-token', got %q", cfg.TelegramToken)
	}
	if cfg.ChatID != -999888 {
		t.Errorf("expected chat ID -999888, got %d", cfg.ChatID)
	}
	if cfg.ServerAddr != ":7070" {
		t.Errorf("expected addr ':7070', got %q", cfg.ServerAddr)
	}
}

func TestEnvVarsOverrideYAMLFile(t *testing.T) {
	clearEnv(t)
	path := writeYAML(t, `
telegram_bot_token: "yaml-token"
telegram_chat_id: "-999888"
server_addr: ":7070"
`)
	os.Setenv("CONFIG_FILE", path)
	os.Setenv("TELEGRAM_BOT_TOKEN", "env-token")
	os.Setenv("TELEGRAM_CHAT_ID", "42")
	os.Setenv("SERVER_ADDR", ":5050")
	defer os.Unsetenv("CONFIG_FILE")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")
	defer os.Unsetenv("SERVER_ADDR")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TelegramToken != "env-token" {
		t.Errorf("expected env token 'env-token', got %q", cfg.TelegramToken)
	}
	if cfg.ChatID != 42 {
		t.Errorf("expected chat ID 42, got %d", cfg.ChatID)
	}
	if cfg.ServerAddr != ":5050" {
		t.Errorf("expected addr ':5050', got %q", cfg.ServerAddr)
	}
}

func TestLoadMissingExplicitConfigFile(t *testing.T) {
	clearEnv(t)
	os.Setenv("CONFIG_FILE", filepath.Join(t.TempDir(), "nonexistent.yaml"))
	defer os.Unsetenv("CONFIG_FILE")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when CONFIG_FILE points to a missing file")
	}
}

func TestLoadYAMLOnlyPartialOverridesDefault(t *testing.T) {
	clearEnv(t)
	// File only sets token and chat ID; server_addr should default to ":8080".
	path := writeYAML(t, `
telegram_bot_token: "partial-token"
telegram_chat_id: "1234"
`)
	os.Setenv("CONFIG_FILE", path)
	defer os.Unsetenv("CONFIG_FILE")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ServerAddr != ":8080" {
		t.Errorf("expected default addr ':8080', got %q", cfg.ServerAddr)
	}
}

func TestLoadServerSecretFromYAML(t *testing.T) {
	clearEnv(t)
	path := writeYAML(t, `
telegram_bot_token: "tok"
telegram_chat_id: "1"
server_secret: "yaml-secret"
`)
	os.Setenv("CONFIG_FILE", path)
	defer os.Unsetenv("CONFIG_FILE")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ServerSecret != "yaml-secret" {
		t.Errorf("expected server_secret 'yaml-secret', got %q", cfg.ServerSecret)
	}
}

func TestLoadServerSecretFromEnv(t *testing.T) {
	clearEnv(t)
	os.Setenv("TELEGRAM_BOT_TOKEN", "tok")
	os.Setenv("TELEGRAM_CHAT_ID", "1")
	os.Setenv("SERVER_SECRET", "env-secret")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")
	defer os.Unsetenv("SERVER_SECRET")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ServerSecret != "env-secret" {
		t.Errorf("expected server_secret 'env-secret', got %q", cfg.ServerSecret)
	}
}

func TestEnvSecretOverridesYAMLSecret(t *testing.T) {
	clearEnv(t)
	path := writeYAML(t, `
telegram_bot_token: "tok"
telegram_chat_id: "1"
server_secret: "yaml-secret"
`)
	os.Setenv("CONFIG_FILE", path)
	os.Setenv("SERVER_SECRET", "env-secret")
	defer os.Unsetenv("CONFIG_FILE")
	defer os.Unsetenv("SERVER_SECRET")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ServerSecret != "env-secret" {
		t.Errorf("expected env secret to override yaml, got %q", cfg.ServerSecret)
	}
}

func TestLoadNoSecretIsEmpty(t *testing.T) {
	clearEnv(t)
	os.Setenv("TELEGRAM_BOT_TOKEN", "tok")
	os.Setenv("TELEGRAM_CHAT_ID", "1")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ServerSecret != "" {
		t.Errorf("expected empty server_secret, got %q", cfg.ServerSecret)
	}
}

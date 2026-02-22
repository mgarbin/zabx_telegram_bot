// zabx_telegram_bot receives Zabbix trigger alerts over HTTP and forwards
// them to a Telegram group chat via the Bot API. When a trigger transitions
// from PROBLEM to RESOLVED the original Telegram message is edited in-place
// rather than posting a duplicate.
//
// Configuration is read from an optional YAML file (default: config.yaml,
// overridable via CONFIG_FILE) and/or environment variables. Environment
// variables always take precedence over the file.
//
// Required (env var or config file):
//
//	TELEGRAM_BOT_TOKEN – bot token from BotFather
//	TELEGRAM_CHAT_ID   – numeric ID of the target group chat
//
// Optional:
//
//	SERVER_ADDR  – listen address for the HTTP server (default ":8080")
//	CONFIG_FILE  – path to the YAML configuration file (default "config.yaml")
//
// Endpoint:
//
//	POST /zabbix/alert  – receive a Zabbix alert JSON payload
package main

import (
	"log"
	"net/http"

	"github.com/mgarbin/zabbix-telegram-event-correlator/config"
	"github.com/mgarbin/zabbix-telegram-event-correlator/internal/bot"
	"github.com/mgarbin/zabbix-telegram-event-correlator/internal/handler"
	"github.com/mgarbin/zabbix-telegram-event-correlator/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	tgBot, err := bot.New(cfg.TelegramToken, cfg.ChatID)
	if err != nil {
		log.Fatalf("failed to create Telegram bot: %v", err)
	}

	msgStore := store.New()
	alertHandler := handler.New(tgBot, msgStore, cfg.ServerSecret)

	mux := http.NewServeMux()
	mux.Handle("/zabbix/alert", alertHandler)

	log.Printf("zabbix-telegram-event-correlator listening on %s", cfg.ServerAddr)
	if err := http.ListenAndServe(cfg.ServerAddr, mux); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

// zabx_telegram_bot receives Zabbix trigger alerts over HTTP and forwards
// them to a Telegram group chat via the Bot API. When a trigger transitions
// from PROBLEM to RESOLVED the original Telegram message is edited in-place
// rather than posting a duplicate.
//
// Required environment variables:
//
//	TELEGRAM_BOT_TOKEN – bot token from BotFather
//	TELEGRAM_CHAT_ID   – numeric ID of the target group chat
//
// Optional environment variables:
//
//	SERVER_ADDR – listen address for the HTTP server (default ":8080")
//
// Endpoint:
//
//	POST /zabbix/alert  – receive a Zabbix alert JSON payload
package main

import (
	"log"
	"net/http"

	"github.com/mgarbin/zabx_telegram_bot/config"
	"github.com/mgarbin/zabx_telegram_bot/internal/bot"
	"github.com/mgarbin/zabx_telegram_bot/internal/handler"
	"github.com/mgarbin/zabx_telegram_bot/internal/store"
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
	alertHandler := handler.New(tgBot, msgStore)

	mux := http.NewServeMux()
	mux.Handle("/zabbix/alert", alertHandler)

	log.Printf("zabx_telegram_bot listening on %s", cfg.ServerAddr)
	if err := http.ListenAndServe(cfg.ServerAddr, mux); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

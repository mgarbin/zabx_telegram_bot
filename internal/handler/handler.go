// Package handler implements the HTTP webhook handler that receives Zabbix
// alert notifications and forwards them to Telegram.
package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mgarbin/zabx_telegram_bot/internal/store"
)

// Sender is the interface the handler uses to interact with Telegram.
// Using an interface makes the handler easy to test without a real bot.
type Sender interface {
	SendMessage(text string) (int, error)
	EditMessage(messageID int, text string) error
}

// AlertStatus represents the status field sent by Zabbix.
type AlertStatus string

const (
	StatusProblem  AlertStatus = "PROBLEM"
	StatusResolved AlertStatus = "RESOLVED"
)

// ZabbixAlert is the JSON payload POSTed by Zabbix.
type ZabbixAlert struct {
	TriggerID   string      `json:"trigger_id"`
	TriggerName string      `json:"trigger_name"`
	Status      AlertStatus `json:"status"`
	Severity    string      `json:"severity"`
	Host        string      `json:"host"`
	EventID     string      `json:"event_id"`
	Message     string      `json:"message"`
	Secret      string      `json:"secret"`
}

// Handler processes incoming Zabbix alerts.
type Handler struct {
	bot    Sender
	store  *store.MessageStore
	secret string
}

// New creates a Handler wired to the given Telegram sender and message store.
// If secret is non-empty every incoming request must carry a matching "secret"
// field in its JSON body; otherwise the request is rejected with 401.
func New(bot Sender, s *store.MessageStore, secret string) *Handler {
	return &Handler{bot: bot, store: s, secret: secret}
}

// ServeHTTP handles POST /zabbix/alert requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var alert ZabbixAlert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if alert.EventID == "" {
		http.Error(w, "event_id is required", http.StatusBadRequest)
		return
	}

	if h.secret != "" && alert.Secret != h.secret {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	text := formatMessage(alert)

	switch alert.Status {
	case StatusProblem:
		msgID, err := h.bot.SendMessage(text)
		if err != nil {
			log.Printf("ERROR sending Telegram message for event %s: %v", alert.EventID, err)
			http.Error(w, "failed to send Telegram message", http.StatusInternalServerError)
			return
		}
		h.store.Set(alert.EventID, msgID)
		log.Printf("PROBLEM alert sent for event %s (message %d)", alert.EventID, msgID)

	case StatusResolved:
		if msgID, ok := h.store.Get(alert.EventID); ok {
			if err := h.bot.EditMessage(msgID, text); err != nil {
				log.Printf("ERROR editing Telegram message %d for event %s: %v", msgID, alert.EventID, err)
				http.Error(w, "failed to edit Telegram message", http.StatusInternalServerError)
				return
			}
			h.store.Delete(alert.EventID)
			log.Printf("RESOLVED alert updated for event %s (message %d)", alert.EventID, msgID)
		} else {
			// No tracked message found – send a new one so the resolution is not lost.
			msgID, err := h.bot.SendMessage(text)
			if err != nil {
				log.Printf("ERROR sending Telegram message for resolved event %s: %v", alert.EventID, err)
				http.Error(w, "failed to send Telegram message", http.StatusInternalServerError)
				return
			}
			log.Printf("RESOLVED alert sent (no prior message tracked) for event %s (message %d)", alert.EventID, msgID)
		}

	default:
		// Unknown status – send as a plain informational message.
		msgID, err := h.bot.SendMessage(text)
		if err != nil {
			log.Printf("ERROR sending Telegram message for event %s: %v", alert.EventID, err)
			http.Error(w, "failed to send Telegram message", http.StatusInternalServerError)
			return
		}
		log.Printf("INFO alert sent for event %s (message %d)", alert.EventID, msgID)
	}

	w.WriteHeader(http.StatusOK)
}

// formatMessage builds a human-readable HTML message from the alert payload.
func formatMessage(a ZabbixAlert) string {
	var sb strings.Builder

	statusEmoji := statusEmoji(a.Status)
	sb.WriteString(fmt.Sprintf("%s <b>%s</b>\n", statusEmoji, escapeHTML(string(a.Status))))
	if a.TriggerName != "" {
		sb.WriteString(fmt.Sprintf("🔔 <b>Trigger:</b> %s\n", escapeHTML(a.TriggerName)))
	}
	if a.Host != "" {
		sb.WriteString(fmt.Sprintf("🖥 <b>Host:</b> %s\n", escapeHTML(a.Host)))
	}
	if a.Severity != "" {
		sb.WriteString(fmt.Sprintf("⚠️ <b>Severity:</b> %s\n", escapeHTML(a.Severity)))
	}
	if a.Message != "" {
		sb.WriteString(fmt.Sprintf("📝 <b>Details:</b> %s\n", escapeHTML(a.Message)))
	}
	if a.EventID != "" {
		sb.WriteString(fmt.Sprintf("🆔 <b>Event ID:</b> %s\n", escapeHTML(a.EventID)))
	}
	sb.WriteString(fmt.Sprintf("🕐 <b>Time:</b> %s", time.Now().UTC().Format("2006-01-02 15:04:05 UTC")))

	return sb.String()
}

func statusEmoji(s AlertStatus) string {
	switch s {
	case StatusProblem:
		return "🔴"
	case StatusResolved:
		return "✅"
	default:
		return "ℹ️"
	}
}

// escapeHTML escapes the characters that have special meaning in Telegram's
// HTML parse mode: &, <, >.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

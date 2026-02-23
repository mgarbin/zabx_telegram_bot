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

	"github.com/mgarbin/zabbix-telegram-event-correlator/internal/store"
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

	timeFormat = "2006-01-02 15:04:05 MST"
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
	store  store.Store
	secret string
}

// New creates a Handler wired to the given Telegram sender and message store.
// If secret is non-empty every incoming request must carry a matching "secret"
// field in its JSON body; otherwise the request is rejected with 401.
func New(bot Sender, s store.Store, secret string) *Handler {
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

	switch alert.Status {
	case StatusProblem:
		now := time.Now()
		text := formatMessage(alert, now, "", "")
		msgID, err := h.bot.SendMessage(text)
		if err != nil {
			log.Printf("ERROR sending Telegram message for event %s: %v", alert.EventID, err)
			http.Error(w, "failed to send Telegram message", http.StatusInternalServerError)
			return
		}
		h.store.Set(alert.EventID, store.Entry{
			MessageID: msgID,
			StartTime: now.Format(timeFormat),
			Message:   alert.Message,
			Severity:  alert.Severity,
		})
		log.Printf("PROBLEM alert sent for event %s (message %d)", alert.EventID, msgID)

	case StatusResolved:
		if entry, ok := h.store.Get(alert.EventID); ok {
			if alert.Severity == "" && entry.Severity != "" {
				alert.Severity = entry.Severity
			}
			text := formatMessage(alert, time.Now(), entry.StartTime, entry.Message)
			if err := h.bot.EditMessage(entry.MessageID, text); err != nil {
				log.Printf("ERROR editing Telegram message %d for event %s: %v", entry.MessageID, alert.EventID, err)
				http.Error(w, "failed to edit Telegram message", http.StatusInternalServerError)
				return
			}
			h.store.Delete(alert.EventID)
			log.Printf("RESOLVED alert updated for event %s (message %d)", alert.EventID, entry.MessageID)
		} else {
			// No tracked message found – send a new one so the resolution is not lost.
			text := formatMessage(alert, time.Now(), "", "")
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
		text := formatMessage(alert, time.Now(), "", "")
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
// now is the current time used as Start Time (PROBLEM) or End Time (RESOLVED).
// startTime, if non-empty, is the Start Time preserved from the original PROBLEM event.
// origMessage, if non-empty, is the Details preserved from the original PROBLEM event.
func formatMessage(a ZabbixAlert, now time.Time, startTime, origMessage string) string {
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
		sb.WriteString(fmt.Sprintf("%s <b>Severity:</b> %s\n", severityEmoji(a.Severity), escapeHTML(a.Severity)))
	}
	// For RESOLVED, preserve the original Details from the PROBLEM event (if any).
	msg := a.Message
	if a.Status == StatusResolved && origMessage != "" {
		msg = origMessage
	}
	if msg != "" {
		sb.WriteString(fmt.Sprintf("📝 <b>Details:</b> %s\n", escapeHTML(msg)))
	}
	if a.EventID != "" {
		sb.WriteString(fmt.Sprintf("🆔 <b>Event ID:</b> %s\n", escapeHTML(a.EventID)))
	}
	if a.Status == StatusResolved {
		if startTime != "" {
			sb.WriteString(fmt.Sprintf("🕐 <b>Start Time:</b> %s\n", startTime))
		}
		sb.WriteString(fmt.Sprintf("🕑 <b>End Time:</b> %s", now.Format(timeFormat)))
	} else {
		sb.WriteString(fmt.Sprintf("🕐 <b>Start Time:</b> %s", now.Format(timeFormat)))
	}

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

func severityEmoji(sev string) string {
	switch strings.ToUpper(sev) {
	case "DISASTER":
		return "💀"
	case "HIGH":
		return "🔥"
	case "AVERAGE":
		return "⚡"
	case "WARNING":
		return "⚠️"
	case "INFORMATION":
		return "ℹ️"
	case "NOT_CLASSIFIED":
		return "❓"
	default:
		return "❔"
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

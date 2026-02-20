# zabx_telegram_bot

A Go service that receives [Zabbix](https://www.zabbix.com/) trigger alerts via
HTTP webhook and forwards them to a Telegram group chat using the
[Telegram Bot API](https://core.telegram.org/bots/api).

Key behaviour:

* When Zabbix fires a **PROBLEM** alert a new Telegram message is sent and the
  message ID is stored in memory, keyed by the Zabbix trigger ID.
* When Zabbix fires the matching **RESOLVED** alert the _same_ Telegram message
  is edited in-place (status changes from 🔴 PROBLEM → ✅ RESOLVED), so the
  chat history stays clean.
* If a RESOLVED arrives without a tracked PROBLEM message (e.g. after a restart)
  a new message is sent so the event is never silently dropped.

---

## Configuration

All configuration is provided through environment variables.

| Variable             | Required | Default | Description                                        |
|----------------------|----------|---------|----------------------------------------------------|
| `TELEGRAM_BOT_TOKEN` | ✅       |         | Bot token from [@BotFather](https://t.me/BotFather) |
| `TELEGRAM_CHAT_ID`   | ✅       |         | Numeric ID of the target group chat                |
| `SERVER_ADDR`        | ❌       | `:8080` | Address the HTTP server listens on                 |

> **Finding the chat ID** – Add the bot to the group, send a message, then call
> `https://api.telegram.org/bot<TOKEN>/getUpdates` to find the `chat.id` value.

---

## Running

```bash
export TELEGRAM_BOT_TOKEN="123456:ABC-DEF..."
export TELEGRAM_CHAT_ID="-100987654321"

go run .
```

The service starts an HTTP server on `:8080` (or the value of `SERVER_ADDR`).

---

## Zabbix webhook setup

1. In Zabbix go to **Administration → Media types → Create media type**.
2. Choose **Webhook** as the type.
3. Set the URL to `http://<your-server>:8080/zabbix/alert`.
4. Set the HTTP method to **POST**.
5. Set the Content-Type header to `application/json`.
6. Map the following parameters to the message body:

```json
{
  "trigger_id":   "{TRIGGER.ID}",
  "trigger_name": "{TRIGGER.NAME}",
  "status":       "{TRIGGER.STATUS}",
  "severity":     "{TRIGGER.SEVERITY}",
  "host":         "{HOST.NAME}",
  "event_id":     "{EVENT.ID}",
  "message":      "{ALERT.MESSAGE}"
}
```

### Accepted payload fields

| Field          | Type   | Required | Description                         |
|----------------|--------|----------|-------------------------------------|
| `trigger_id`   | string | ✅       | Unique ID of the Zabbix trigger      |
| `trigger_name` | string |          | Human-readable trigger name          |
| `status`       | string |          | `PROBLEM` or `RESOLVED`              |
| `severity`     | string |          | Trigger severity label               |
| `host`         | string |          | Affected host name                   |
| `event_id`     | string |          | Zabbix event ID                      |
| `message`      | string |          | Additional details / description     |

---

## Project layout

```
.
├── main.go                   # Entry point – wires config, bot, store and HTTP server
├── config/
│   └── config.go             # Load configuration from environment
├── internal/
│   ├── bot/
│   │   └── bot.go            # Telegram Bot API wrapper (send / edit messages)
│   ├── handler/
│   │   └── handler.go        # HTTP handler for POST /zabbix/alert
│   └── store/
│       └── store.go          # Thread-safe in-memory trigger→message-ID map
```

---

## Running tests

```bash
go test ./...
```
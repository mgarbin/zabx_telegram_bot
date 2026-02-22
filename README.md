# zabbix-telegram-event-correlator

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

Configuration can be provided through **environment variables**, an external
**YAML file**, or a combination of both. Environment variables always take
precedence over file values.

### Environment variables

| Variable             | Required | Default | Description                                        |
|----------------------|----------|---------|----------------------------------------------------|
| `TELEGRAM_BOT_TOKEN` | ✅       |         | Bot token from [@BotFather](https://t.me/BotFather) |
| `TELEGRAM_CHAT_ID`   | ✅       |         | Numeric ID of the target group chat                |
| `SERVER_ADDR`        | ❌       | `:8080` | Address the HTTP server listens on                 |
| `CONFIG_FILE`        | ❌       | `config.yaml` | Path to an optional YAML configuration file  |

> **Finding the chat ID** – Add the bot to the group, send a message, then call
> `https://api.telegram.org/bot<TOKEN>/getUpdates` to find the `chat.id` value.

### YAML configuration file

By default the bot looks for a `config.yaml` file in the current working
directory. If the file does not exist, environment variables must supply all
required values. Set `CONFIG_FILE` to use a different path (the bot returns an
error if that explicit path does not exist).

```yaml
telegram_bot_token: "123456:ABC-DEF..."
telegram_chat_id: "-100987654321"

# Optional – defaults to :8080
# server_addr: ":8080"

# Optional: shared secret that must be present in every incoming JSON body.
# When set, requests without a matching "secret" field are rejected with 401.
# server_secret: "change-me"
```

A ready-to-edit template is provided as `config.yaml.example`.

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
  "message":      "{ALERT.MESSAGE}",
  "secret":       "{API.SECRET}"
}
```
### Accepted payload fields

| Field          | Type   | Required | Description                                                                 |
|----------------|--------|----------|-----------------------------------------------------------------------------|
| `trigger_id`   | string |          | Unique ID of the Zabbix trigger                                             |
| `trigger_name` | string |          | Human-readable trigger name                                                 |
| `status`       | string |          | `PROBLEM` or `RESOLVED`                                                     |
| `severity`     | string |          | Trigger severity label                                                      |
| `host`         | string |          | Affected host name                                                          |
| `event_id`     | string |   ✅     | Zabbix event ID                                                             |
| `message`      | string |          | Additional details / description                                            |
| `secret`       | string |          | A secret key that allow to send payload to your http server in secure way   |

---

## Zabbix action setup

1. In zabbix go to **Configuration -> Action -> Trigger Action**.
2. Create a new trigger action.
3. Into **Operation** tab inside the **Operations** section add a new one.
4. **subject** of the new operation need this json :  
```json
{"status":"PROBLEM","severity":"{EVENT.SEVERITY}","host":"{HOST.NAME}","eventId":"{EVENT.ID}","eventName":"{EVENT.NAME}"}
```
4. Inside the **message** you can ad other zabbix values as you want
5. Into **Recovery operations** add a new one.
6. **subject** ot he new operation need this json :
```json
{"status":"RESOLVED","host":"{HOST.NAME}","eventId":"{EVENT.ID}","eventName":"{EVENT.NAME}"}
```

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
│       └── store.go          # Thread-safe in-memory event-ID → message-ID map
```

---

## Running tests

```bash
go test ./...
```

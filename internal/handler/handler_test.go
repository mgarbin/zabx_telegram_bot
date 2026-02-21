package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mgarbin/zabx_telegram_bot/internal/handler"
	"github.com/mgarbin/zabx_telegram_bot/internal/store"
)

// mockBot records which method was last called and with which arguments.
type mockBot struct {
	sentText    string
	sentMsgID   int
	editedMsgID int
	editedText  string
	sendErr     error
	editErr     error
}

func (m *mockBot) SendMessage(text string) (int, error) {
	m.sentText = text
	m.sentMsgID++
	return m.sentMsgID, m.sendErr
}

func (m *mockBot) EditMessage(messageID int, text string) error {
	m.editedMsgID = messageID
	m.editedText = text
	return m.editErr
}

func postAlert(t *testing.T, h http.Handler, alert handler.ZabbixAlert) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(alert)
	req := httptest.NewRequest(http.MethodPost, "/zabbix/alert", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestProblemSendsNewMessage(t *testing.T) {
	mb := &mockBot{}
	s := store.New()
	h := handler.New(mb, s, "")

	alert := handler.ZabbixAlert{
		TriggerID:   "100",
		TriggerName: "High CPU",
		Status:      handler.StatusProblem,
		Severity:    "High",
		Host:        "server1",
	}

	resp := postAlert(t, h, alert)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	// The message ID must have been stored.
	msgID, ok := s.Get("100")
	if !ok {
		t.Fatal("expected trigger ID to be stored after PROBLEM alert")
	}
	if msgID != 1 {
		t.Fatalf("expected stored message ID 1, got %d", msgID)
	}
}

func TestResolvedEditsExistingMessage(t *testing.T) {
	mb := &mockBot{}
	s := store.New()
	h := handler.New(mb, s, "")

	// First: a PROBLEM alert.
	postAlert(t, h, handler.ZabbixAlert{
		TriggerID:   "200",
		TriggerName: "Disk Full",
		Status:      handler.StatusProblem,
		Host:        "server2",
	})

	storedID, _ := s.Get("200")

	// Then: a RESOLVED alert for the same trigger.
	resp := postAlert(t, h, handler.ZabbixAlert{
		TriggerID:   "200",
		TriggerName: "Disk Full",
		Status:      handler.StatusResolved,
		Host:        "server2",
	})

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	// EditMessage must have been called with the stored ID.
	if mb.editedMsgID != storedID {
		t.Fatalf("expected EditMessage to be called with message ID %d, got %d", storedID, mb.editedMsgID)
	}

	// The entry must be removed from the store after resolution.
	if _, ok := s.Get("200"); ok {
		t.Fatal("expected trigger to be removed from store after RESOLVED")
	}
}

func TestResolvedWithNoTrackedMessageSendsNew(t *testing.T) {
	mb := &mockBot{}
	s := store.New()
	h := handler.New(mb, s, "")

	resp := postAlert(t, h, handler.ZabbixAlert{
		TriggerID:   "300",
		TriggerName: "Memory Low",
		Status:      handler.StatusResolved,
	})

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	// A new message should have been sent (SendMessage called, not Edit).
	if mb.sentText == "" {
		t.Fatal("expected SendMessage to be called for untracked RESOLVED alert")
	}
	if mb.editedMsgID != 0 {
		t.Fatal("expected EditMessage NOT to be called when no tracked message exists")
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h := handler.New(&mockBot{}, store.New(), "")
	req := httptest.NewRequest(http.MethodGet, "/zabbix/alert", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestInvalidJSON(t *testing.T) {
	h := handler.New(&mockBot{}, store.New(), "")
	req := httptest.NewRequest(http.MethodPost, "/zabbix/alert", bytes.NewBufferString("{bad json"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestMissingTriggerID(t *testing.T) {
	h := handler.New(&mockBot{}, store.New(), "")
	resp := postAlert(t, h, handler.ZabbixAlert{
		TriggerName: "Some trigger",
		Status:      handler.StatusProblem,
	})

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing trigger_id, got %d", resp.Code)
	}
}

func TestSecretValidRequest(t *testing.T) {
	mb := &mockBot{}
	h := handler.New(mb, store.New(), "mysecret")

	resp := postAlert(t, h, handler.ZabbixAlert{
		TriggerID:   "400",
		TriggerName: "High CPU",
		Status:      handler.StatusProblem,
		Secret:      "mysecret",
	})

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 with correct secret, got %d", resp.Code)
	}
}

func TestSecretWrongValue(t *testing.T) {
	h := handler.New(&mockBot{}, store.New(), "mysecret")

	resp := postAlert(t, h, handler.ZabbixAlert{
		TriggerID:   "401",
		TriggerName: "High CPU",
		Status:      handler.StatusProblem,
		Secret:      "wrongsecret",
	})

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong secret, got %d", resp.Code)
	}
}

func TestSecretMissing(t *testing.T) {
	h := handler.New(&mockBot{}, store.New(), "mysecret")

	resp := postAlert(t, h, handler.ZabbixAlert{
		TriggerID:   "402",
		TriggerName: "High CPU",
		Status:      handler.StatusProblem,
		// Secret omitted
	})

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when secret is missing, got %d", resp.Code)
	}
}

func TestNoSecretConfiguredAllowsAnyRequest(t *testing.T) {
	mb := &mockBot{}
	h := handler.New(mb, store.New(), "")

	resp := postAlert(t, h, handler.ZabbixAlert{
		TriggerID:   "403",
		TriggerName: "High CPU",
		Status:      handler.StatusProblem,
		// No secret in body – should still be allowed when none configured
	})

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 when no secret configured, got %d", resp.Code)
	}
}

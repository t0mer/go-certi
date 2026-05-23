package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/t0mer/go-certi/internal/ct"
	"github.com/t0mer/go-certi/internal/models"
	"github.com/t0mer/go-certi/internal/notify"
)

func TestDispatchGreenAPI(t *testing.T) {
	var received map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"idMessage":"test123"}`))
	}))
	defer srv.Close()

	cfgJSON, _ := json.Marshal(map[string]string{
		"instance_id":        "1234",
		"api_token_instance": "token",
		"chat_id":            "972501234567@c.us",
		"api_url":            srv.URL,
	})

	d := notify.New()
	d.Dispatch(context.Background(), models.NotificationChannel{
		ID:      "test",
		Name:    "test",
		Type:    "greenapi",
		Config:  string(cfgJSON),
		Enabled: true,
	}, "example.com", ct.Cert{
		IssuerCA:  "Let's Encrypt",
		SubjectCN: "example.com",
		SANs:      []string{"example.com"},
		NotBefore: time.Now().UTC().Format(time.RFC3339),
		NotAfter:  time.Now().Add(90 * 24 * time.Hour).UTC().Format(time.RFC3339),
		Serial:    "abc123",
	})

	if received["chatId"] != "972501234567@c.us" {
		t.Errorf("chatId = %v", received["chatId"])
	}
}

func TestDispatchWaWeb(t *testing.T) {
	var received map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfgJSON, _ := json.Marshal(map[string]string{
		"base_url": srv.URL,
		"phone":    "+972501234567",
	})

	d := notify.New()
	d.Dispatch(context.Background(), models.NotificationChannel{
		ID: "t", Name: "t", Type: "waweb", Config: string(cfgJSON), Enabled: true,
	}, "example.com", ct.Cert{
		IssuerCA:  "CA",
		SubjectCN: "example.com",
		SANs:      []string{"example.com"},
		NotBefore: time.Now().UTC().Format(time.RFC3339),
		NotAfter:  time.Now().Add(90 * 24 * time.Hour).UTC().Format(time.RFC3339),
		Serial:    "xyz",
	})

	if received["phone"] != "+972501234567" {
		t.Errorf("phone = %v", received["phone"])
	}
}

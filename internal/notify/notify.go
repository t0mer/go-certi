package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/t0mer/go-certi/internal/ct"
	"github.com/t0mer/go-certi/internal/models"
)

// EventType constants for notification events.
const (
	EventNewCert      = "new_cert"
	EventExpiringSoon = "expiring_soon"
	EventExpired      = "expired"
	EventRevoked      = "revoked"
	EventCAChanged    = "ca_changed"
)

// Dispatcher sends notifications for new certificates.
type Dispatcher struct{}

// New creates a Dispatcher.
func New() *Dispatcher { return &Dispatcher{} }

// formatMessage returns a human-readable notification string.
func formatMessage(fqdn string, cert ct.Cert) string {
	sanCount := len(cert.SANs)
	sanStr := ""
	if sanCount > 1 {
		sanStr = fmt.Sprintf(" (+%d SANs)", sanCount-1)
	}
	nb := cert.NotBefore
	na := cert.NotAfter
	// Try to parse and reformat as date only for readability
	if t, err := time.Parse(time.RFC3339, cert.NotBefore); err == nil {
		nb = t.Format("2006-01-02")
	}
	if t, err := time.Parse(time.RFC3339, cert.NotAfter); err == nil {
		na = t.Format("2006-01-02")
	}
	return fmt.Sprintf(
		"[go-certi] New certificate for %s\nIssuer: %s\nCN: %s%s\nValid: %s → %s\nSerial: %s",
		fqdn, cert.IssuerCA, cert.SubjectCN, sanStr, nb, na, cert.Serial,
	)
}

// Dispatch sends a notification via channel. Best-effort: errors are logged, never propagated.
func (d *Dispatcher) Dispatch(ctx context.Context, channel models.NotificationChannel, fqdn string, cert ct.Cert) {
	if !channel.Enabled {
		return
	}
	msg := formatMessage(fqdn, cert)
	var err error
	switch channel.Type {
	case "shoutrrr":
		err = dispatchShoutrrr(ctx, channel.Config, msg)
	case "greenapi":
		err = dispatchGreenAPI(ctx, channel.Config, msg)
	case "waweb":
		err = dispatchWaWeb(ctx, channel.Config, msg)
	default:
		slog.Warn("unknown channel type", "type", channel.Type, "channel", channel.Name)
		return
	}
	if err != nil {
		slog.Error("notification failed", "channel", channel.Name, "type", channel.Type, "err", err)
	}
}

// DispatchAll sends a notification to all channels. Best-effort.
func (d *Dispatcher) DispatchAll(ctx context.Context, channels []models.NotificationChannel, fqdn string, cert ct.Cert) {
	for _, ch := range channels {
		d.Dispatch(ctx, ch, fqdn, cert)
	}
}

// TestChannel sends a test message to a single channel.
func (d *Dispatcher) TestChannel(ctx context.Context, channel models.NotificationChannel) error {
	testCert := ct.Cert{
		IssuerCA:  "Test CA",
		SubjectCN: "test.example.com",
		SANs:      []string{"test.example.com"},
		NotBefore: time.Now().UTC().Format(time.RFC3339),
		NotAfter:  time.Now().Add(90 * 24 * time.Hour).UTC().Format(time.RFC3339),
		Serial:    "00:11:22:33",
	}
	msg := "[go-certi] Test notification\n" + formatMessage("test.example.com", testCert)
	switch channel.Type {
	case "shoutrrr":
		return dispatchShoutrrr(ctx, channel.Config, msg)
	case "greenapi":
		return dispatchGreenAPI(ctx, channel.Config, msg)
	case "waweb":
		return dispatchWaWeb(ctx, channel.Config, msg)
	}
	return fmt.Errorf("unknown channel type: %s", channel.Type)
}

// formatEventMessage formats a notification message for event-based alerts.
func formatEventMessage(event, fqdn string, cert models.Certificate, oldCA string) string {
	expiry := cert.NotAfter
	if t, err := time.Parse(time.RFC3339, cert.NotAfter); err == nil {
		expiry = t.Format("2006-01-02")
	}
	serialShort := cert.Serial
	if len(serialShort) > 16 {
		serialShort = serialShort[:16]
	}
	switch event {
	case EventExpiringSoon:
		return fmt.Sprintf("[go-certi] Certificate expiring soon for %s\nCN: %s\nIssuer: %s\nExpires: %s\nSerial: %s",
			fqdn, cert.SubjectCn, cert.IssuerCa, expiry, serialShort)
	case EventExpired:
		return fmt.Sprintf("[go-certi] Certificate EXPIRED for %s\nCN: %s\nIssuer: %s\nExpired: %s\nSerial: %s",
			fqdn, cert.SubjectCn, cert.IssuerCa, expiry, serialShort)
	case EventRevoked:
		return fmt.Sprintf("[go-certi] Certificate REVOKED for %s\nCN: %s\nIssuer: %s\nSerial: %s",
			fqdn, cert.SubjectCn, cert.IssuerCa, serialShort)
	case EventCAChanged:
		return fmt.Sprintf("[go-certi] Certificate Authority changed for %s\nWas: %s\nNow: %s\nCN: %s\nSerial: %s",
			fqdn, oldCA, cert.IssuerCa, cert.SubjectCn, serialShort)
	default:
		return fmt.Sprintf("[go-certi] Certificate event (%s) for %s\nCN: %s", event, fqdn, cert.SubjectCn)
	}
}

// DispatchEvent sends an event-based notification for a DB certificate record.
func (d *Dispatcher) DispatchEvent(ctx context.Context, channel models.NotificationChannel, event, fqdn string, cert models.Certificate, oldCA string) {
	if !channel.Enabled {
		return
	}
	msg := formatEventMessage(event, fqdn, cert, oldCA)
	var err error
	switch channel.Type {
	case "shoutrrr":
		err = dispatchShoutrrr(ctx, channel.Config, msg)
	case "greenapi":
		err = dispatchGreenAPI(ctx, channel.Config, msg)
	case "waweb":
		err = dispatchWaWeb(ctx, channel.Config, msg)
	default:
		slog.Warn("unknown channel type", "type", channel.Type)
		return
	}
	if err != nil {
		slog.Error("event notification failed", "channel", channel.Name, "event", event, "err", err)
	}
}

// DispatchEventAll sends an event notification to all channels. Best-effort.
func (d *Dispatcher) DispatchEventAll(ctx context.Context, channels []models.NotificationChannel, event, fqdn string, cert models.Certificate, oldCA string) {
	for _, ch := range channels {
		d.DispatchEvent(ctx, ch, event, fqdn, cert, oldCA)
	}
}

// parseConfig decodes a JSON config string into map[string]string.
func parseConfig(configJSON string) (map[string]string, error) {
	var m map[string]string
	err := json.Unmarshal([]byte(configJSON), &m)
	if err != nil {
		// Try map[string]any and convert
		var m2 map[string]any
		if err2 := json.Unmarshal([]byte(configJSON), &m2); err2 != nil {
			return nil, err
		}
		m = make(map[string]string, len(m2))
		for k, v := range m2 {
			m[k] = fmt.Sprintf("%v", v)
		}
	}
	return m, nil
}

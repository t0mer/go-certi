package scanner

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/t0mer/go-certi/internal/ct"
	"github.com/t0mer/go-certi/internal/models"
	"github.com/t0mer/go-certi/internal/notify"
)

// FetchFunc is injectable for testing.
type FetchFunc func(ctx context.Context, fqdn string, includeSubdomains bool) ([]ct.Cert, error)

// Scanner orchestrates CT scanning, dedup, and notification.
type Scanner struct {
	q    *models.Queries
	fn   FetchFunc
	disp *notify.Dispatcher
}

// New creates a Scanner using the real CT client.
func New(q *models.Queries, client *ct.Client, disp *notify.Dispatcher) *Scanner {
	return &Scanner{q: q, fn: client.FetchCerts, disp: disp}
}

// NewWithFetchFunc creates a Scanner with a custom fetch function (for testing).
func NewWithFetchFunc(q *models.Queries, fn FetchFunc, disp *notify.Dispatcher) *Scanner {
	return &Scanner{q: q, fn: fn, disp: disp}
}

// ScanFQDN fetches new certificates, stores new ones, and dispatches notifications.
func (s *Scanner) ScanFQDN(ctx context.Context, fqdn models.Fqdn) error {
	slog.Info("scanning", "fqdn", fqdn.Fqdn)

	certs, err := s.fn(ctx, fqdn.Fqdn, fqdn.IncludeSubdomains)
	if err != nil {
		return err
	}

	channels, err := s.q.GetFQDNChannels(ctx, fqdn.ID)
	if err != nil {
		return err
	}

	events := parseEvents(fqdn.NotificationEvents)
	wantNewCert := events[notify.EventNewCert]

	var newCerts []models.Certificate
	for _, cert := range certs {
		sans, _ := json.Marshal(cert.SANs)
		row, err := s.q.InsertCertificate(ctx, models.InsertCertificateParams{
			ID:           uuid.NewString(),
			FqdnID:       fqdn.ID,
			Serial:       cert.Serial,
			IssuerCa:     cert.IssuerCA,
			IssuerName:   cert.IssuerName,
			SubjectCn:    cert.SubjectCN,
			Sans:         string(sans),
			NotBefore:    cert.NotBefore,
			NotAfter:     cert.NotAfter,
			DiscoveredAt: time.Now().UTC().Format(time.RFC3339),
			Source:       cert.Source,
			Revoked:      cert.Revoked,
		})
		if err != nil {
			// INSERT OR IGNORE returns no rows when the serial already exists;
			// sql.ErrNoRows means duplicate — skip silently.
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			slog.Error("insert cert", "serial", cert.Serial, "err", err)
			continue
		}
		if row.ID == "" {
			// Defensive: treat empty ID as already-existing row.
			continue
		}
		newCerts = append(newCerts, row)

		if fqdn.NotificationsEnabled && wantNewCert {
			s.disp.DispatchAll(ctx, channels, fqdn.Fqdn, cert)
		}
	}

	slog.Info("scan complete", "fqdn", fqdn.Fqdn, "new", len(newCerts), "total", len(certs))

	if fqdn.NotificationsEnabled {
		s.checkEvents(ctx, fqdn, channels, events, newCerts)
	}
	return nil
}

// checkEvents fires expiry/revoked/CA-changed notifications with 24h dedup.
func (s *Scanner) checkEvents(ctx context.Context, fqdn models.Fqdn, channels []models.NotificationChannel, events map[string]bool, newCerts []models.Certificate) {
	now := time.Now().UTC().Format(time.RFC3339)
	threshold := int(fqdn.ExpiryThresholdDays)
	if threshold <= 0 {
		threshold = 10
	}
	cutoff := time.Now().UTC().Add(time.Duration(threshold) * 24 * time.Hour).Format(time.RFC3339)

	if events[notify.EventExpiringSoon] {
		certs, _ := s.q.GetCertsExpiringBefore(ctx, models.GetCertsExpiringBeforeParams{
			FqdnID:     fqdn.ID,
			NotAfter:   now,
			NotAfter_2: cutoff,
		})
		for _, cert := range certs {
			if !s.alreadySent(ctx, fqdn.ID, cert.ID, notify.EventExpiringSoon) {
				s.disp.DispatchEventAll(ctx, channels, notify.EventExpiringSoon, fqdn.Fqdn, cert, "")
				s.logNotification(ctx, fqdn.ID, cert.ID, notify.EventExpiringSoon)
			}
		}
	}

	if events[notify.EventExpired] {
		certs, _ := s.q.GetExpiredCerts(ctx, models.GetExpiredCertsParams{FqdnID: fqdn.ID, NotAfter: now})
		for _, cert := range certs {
			if !s.alreadySent(ctx, fqdn.ID, cert.ID, notify.EventExpired) {
				s.disp.DispatchEventAll(ctx, channels, notify.EventExpired, fqdn.Fqdn, cert, "")
				s.logNotification(ctx, fqdn.ID, cert.ID, notify.EventExpired)
			}
		}
	}

	if events[notify.EventRevoked] {
		certs, _ := s.q.GetRevokedCerts(ctx, fqdn.ID)
		for _, cert := range certs {
			if !s.alreadySent(ctx, fqdn.ID, cert.ID, notify.EventRevoked) {
				s.disp.DispatchEventAll(ctx, channels, notify.EventRevoked, fqdn.Fqdn, cert, "")
				s.logNotification(ctx, fqdn.ID, cert.ID, notify.EventRevoked)
			}
		}
	}

	if events[notify.EventCAChanged] {
		for _, newCert := range newCerts {
			prev, err := s.q.GetPreviousCertForFQDN(ctx, models.GetPreviousCertForFQDNParams{
				FqdnID: fqdn.ID,
				Serial: newCert.Serial,
			})
			if err != nil {
				continue
			}
			if prev.IssuerCa != "" && newCert.IssuerCa != "" && prev.IssuerCa != newCert.IssuerCa {
				if !s.alreadySent(ctx, fqdn.ID, newCert.ID, notify.EventCAChanged) {
					s.disp.DispatchEventAll(ctx, channels, notify.EventCAChanged, fqdn.Fqdn, newCert, prev.IssuerCa)
					s.logNotification(ctx, fqdn.ID, newCert.ID, notify.EventCAChanged)
				}
			}
		}
	}
}

func (s *Scanner) alreadySent(ctx context.Context, fqdnID, certID, event string) bool {
	sent, err := s.q.WasNotificationSent(ctx, models.WasNotificationSentParams{
		FqdnID: fqdnID,
		CertID: certID,
		Event:  event,
	})
	if err != nil {
		return false
	}
	return sent
}

func (s *Scanner) logNotification(ctx context.Context, fqdnID, certID, event string) {
	if err := s.q.LogNotification(ctx, models.LogNotificationParams{ //nolint:errcheck
		ID:     uuid.NewString(),
		FqdnID: fqdnID,
		CertID: certID,
		Event:  event,
	}); err != nil {
		slog.Warn("log notification failed", "err", err)
	}
}

// parseEvents parses a JSON array of event strings into a set map.
func parseEvents(raw string) map[string]bool {
	var list []string
	json.Unmarshal([]byte(raw), &list) //nolint:errcheck
	m := make(map[string]bool, len(list))
	for _, e := range list {
		m[e] = true
	}
	return m
}

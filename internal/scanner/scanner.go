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

	newCount := 0
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
		newCount++
		if fqdn.NotificationsEnabled {
			s.disp.DispatchAll(ctx, channels, fqdn.Fqdn, cert)
		}
	}

	slog.Info("scan complete", "fqdn", fqdn.Fqdn, "new", newCount, "total", len(certs))
	return nil
}

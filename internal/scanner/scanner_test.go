package scanner_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/t0mer/go-certi/internal/ct"
	"github.com/t0mer/go-certi/internal/db"
	"github.com/t0mer/go-certi/internal/models"
	"github.com/t0mer/go-certi/internal/notify"
	"github.com/t0mer/go-certi/internal/scanner"
)

func TestScanFQDN_NewCertsStored(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	q := models.New(conn)
	ctx := context.Background()

	fqdn, err := q.CreateFQDN(ctx, models.CreateFQDNParams{
		ID: uuid.NewString(), Fqdn: "scan-test.com",
		Enabled: true, NotificationsEnabled: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	disp := notify.New()
	scn := scanner.NewWithFetchFunc(q, func(_ context.Context, _ string, _ bool) ([]ct.Cert, error) {
		return []ct.Cert{
			{
				Serial:    "aabbcc",
				IssuerCA:  "Let's Encrypt",
				SubjectCN: "scan-test.com",
				SANs:      []string{"scan-test.com"},
				NotBefore: time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339),
				NotAfter:  time.Now().Add(90 * 24 * time.Hour).UTC().Format(time.RFC3339),
				Source:    "sslmate",
			},
		}, nil
	}, disp)

	if err := scn.ScanFQDN(ctx, fqdn); err != nil {
		t.Fatalf("ScanFQDN: %v", err)
	}

	count, err := q.CountCertificatesByFQDN(ctx, fqdn.ID)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected 1 cert stored, got %d", count)
	}

	// Scan again — should not duplicate
	scn.ScanFQDN(ctx, fqdn)
	count, _ = q.CountCertificatesByFQDN(ctx, fqdn.ID)
	if count != 1 {
		t.Errorf("expected still 1 cert after duplicate scan, got %d", count)
	}
}

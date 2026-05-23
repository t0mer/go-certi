package models_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/t0mer/go-certi/internal/db"
	"github.com/t0mer/go-certi/internal/models"
)

func openDB(t *testing.T) *models.Queries {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return models.New(conn)
}

func TestScheduleCRUD(t *testing.T) {
	q := openDB(t)
	ctx := context.Background()

	sched, err := q.CreateSchedule(ctx, models.CreateScheduleParams{
		ID:        uuid.NewString(),
		Name:      "Test Schedule",
		CronExpr:  "@every 1h",
		IsDefault: false,
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("CreateSchedule: %v", err)
	}
	if sched.Name != "Test Schedule" {
		t.Errorf("Name = %q, want Test Schedule", sched.Name)
	}

	got, err := q.GetSchedule(ctx, sched.ID)
	if err != nil {
		t.Fatalf("GetSchedule: %v", err)
	}
	if got.ID != sched.ID {
		t.Errorf("ID mismatch: got %q, want %q", got.ID, sched.ID)
	}

	// Migration seeds a default schedule; after creating one more we expect >= 2
	list, err := q.ListSchedules(ctx)
	if err != nil {
		t.Fatalf("ListSchedules: %v", err)
	}
	if len(list) < 2 {
		t.Errorf("expected >=2 schedules, got %d", len(list))
	}

	updated, err := q.UpdateSchedule(ctx, models.UpdateScheduleParams{
		Name:      "Updated",
		CronExpr:  "@every 4h",
		IsDefault: false,
		Enabled:   true,
		ID:        sched.ID,
	})
	if err != nil {
		t.Fatalf("UpdateSchedule: %v", err)
	}
	if updated.Name != "Updated" {
		t.Errorf("Updated name = %q, want Updated", updated.Name)
	}

	if err := q.DeleteSchedule(ctx, sched.ID); err != nil {
		t.Fatalf("DeleteSchedule: %v", err)
	}

	// GetDefaultSchedule should return the seeded row
	def, err := q.GetDefaultSchedule(ctx)
	if err != nil {
		t.Fatalf("GetDefaultSchedule: %v", err)
	}
	if !def.IsDefault {
		t.Error("GetDefaultSchedule returned a row with is_default=false")
	}

	// UnsetDefaultSchedules
	if err := q.UnsetDefaultSchedules(ctx); err != nil {
		t.Fatalf("UnsetDefaultSchedules: %v", err)
	}
}

func TestFQDNCRUD(t *testing.T) {
	q := openDB(t)
	ctx := context.Background()

	fqdn, err := q.CreateFQDN(ctx, models.CreateFQDNParams{
		ID:                   uuid.NewString(),
		Fqdn:                 "example.com",
		IncludeSubdomains:    false,
		Enabled:              true,
		NotificationsEnabled: true,
		ScheduleID:           nil,
	})
	if err != nil {
		t.Fatalf("CreateFQDN: %v", err)
	}
	if fqdn.Fqdn != "example.com" {
		t.Errorf("Fqdn = %q, want example.com", fqdn.Fqdn)
	}

	got, err := q.GetFQDN(ctx, fqdn.ID)
	if err != nil {
		t.Fatalf("GetFQDN: %v", err)
	}
	if got.ID != fqdn.ID {
		t.Errorf("ID mismatch: got %q, want %q", got.ID, fqdn.ID)
	}

	byName, err := q.GetFQDNByName(ctx, "example.com")
	if err != nil {
		t.Fatalf("GetFQDNByName: %v", err)
	}
	if byName.ID != fqdn.ID {
		t.Errorf("GetFQDNByName ID mismatch")
	}

	list, err := q.ListFQDNs(ctx)
	if err != nil {
		t.Fatalf("ListFQDNs: %v", err)
	}
	if len(list) < 1 {
		t.Error("expected at least 1 FQDN")
	}

	enabled, err := q.ListEnabledFQDNs(ctx)
	if err != nil {
		t.Fatalf("ListEnabledFQDNs: %v", err)
	}
	if len(enabled) < 1 {
		t.Error("expected at least 1 enabled FQDN")
	}

	updated, err := q.UpdateFQDN(ctx, models.UpdateFQDNParams{
		Fqdn:                 "example.com",
		IncludeSubdomains:    true,
		Enabled:              true,
		NotificationsEnabled: false,
		ScheduleID:           nil,
		ID:                   fqdn.ID,
	})
	if err != nil {
		t.Fatalf("UpdateFQDN: %v", err)
	}
	if !updated.IncludeSubdomains {
		t.Error("IncludeSubdomains should be true after update")
	}

	if err := q.DeleteFQDN(ctx, fqdn.ID); err != nil {
		t.Fatalf("DeleteFQDN: %v", err)
	}
}

func TestChannelCRUD(t *testing.T) {
	q := openDB(t)
	ctx := context.Background()

	ch, err := q.CreateChannel(ctx, models.CreateChannelParams{
		ID:      uuid.NewString(),
		Name:    "Test Telegram",
		Type:    "shoutrrr",
		Config:  `{"url":"telegram://token@telegram?chats=123"}`,
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("CreateChannel: %v", err)
	}
	if ch.Name != "Test Telegram" {
		t.Errorf("Name = %q, want Test Telegram", ch.Name)
	}

	got, err := q.GetChannel(ctx, ch.ID)
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	if got.ID != ch.ID {
		t.Errorf("ID mismatch")
	}

	list, err := q.ListChannels(ctx)
	if err != nil {
		t.Fatalf("ListChannels: %v", err)
	}
	if len(list) < 1 {
		t.Error("expected at least 1 channel")
	}

	updated, err := q.UpdateChannel(ctx, models.UpdateChannelParams{
		Name:    "Updated Telegram",
		Type:    "shoutrrr",
		Config:  `{"url":"telegram://token@telegram?chats=456"}`,
		Enabled: true,
		ID:      ch.ID,
	})
	if err != nil {
		t.Fatalf("UpdateChannel: %v", err)
	}
	if updated.Name != "Updated Telegram" {
		t.Errorf("Updated name = %q, want Updated Telegram", updated.Name)
	}

	// Create an FQDN, link the channel, then verify GetFQDNChannels / GetFQDNChannelIDs
	fqdn, err := q.CreateFQDN(ctx, models.CreateFQDNParams{
		ID: uuid.NewString(), Fqdn: "channel-test.com",
		Enabled: true, NotificationsEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateFQDN for channel test: %v", err)
	}

	if err := q.AddFQDNChannel(ctx, models.AddFQDNChannelParams{
		FqdnID: fqdn.ID, ChannelID: ch.ID,
	}); err != nil {
		t.Fatalf("AddFQDNChannel: %v", err)
	}

	chans, err := q.GetFQDNChannels(ctx, fqdn.ID)
	if err != nil {
		t.Fatalf("GetFQDNChannels: %v", err)
	}
	if len(chans) != 1 || chans[0].ID != ch.ID {
		t.Errorf("GetFQDNChannels: expected 1 channel with ID %q, got %v", ch.ID, chans)
	}

	ids, err := q.GetFQDNChannelIDs(ctx, fqdn.ID)
	if err != nil {
		t.Fatalf("GetFQDNChannelIDs: %v", err)
	}
	if len(ids) != 1 || ids[0] != ch.ID {
		t.Errorf("GetFQDNChannelIDs: expected [%q], got %v", ch.ID, ids)
	}

	if err := q.DeleteFQDNChannels(ctx, fqdn.ID); err != nil {
		t.Fatalf("DeleteFQDNChannels: %v", err)
	}

	if err := q.DeleteChannel(ctx, ch.ID); err != nil {
		t.Fatalf("DeleteChannel: %v", err)
	}
}

func TestCertificateInsert(t *testing.T) {
	q := openDB(t)
	ctx := context.Background()

	fqdn, err := q.CreateFQDN(ctx, models.CreateFQDNParams{
		ID: uuid.NewString(), Fqdn: "cert-test.com",
		Enabled: true, NotificationsEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateFQDN: %v", err)
	}

	cert, err := q.InsertCertificate(ctx, models.InsertCertificateParams{
		ID:           uuid.NewString(),
		FqdnID:       fqdn.ID,
		Serial:       "01:02:03",
		IssuerCa:     "Let's Encrypt",
		SubjectCn:    "cert-test.com",
		Sans:         `["cert-test.com"]`,
		NotBefore:    "2026-01-01T00:00:00Z",
		NotAfter:     "2026-04-01T00:00:00Z",
		DiscoveredAt: "2026-01-01T12:00:00Z",
		Source:       "sslmate",
	})
	if err != nil {
		t.Fatalf("InsertCertificate: %v", err)
	}
	if cert.Serial != "01:02:03" {
		t.Errorf("Serial = %q, want 01:02:03", cert.Serial)
	}

	// GetCertificate
	got, err := q.GetCertificate(ctx, cert.ID)
	if err != nil {
		t.Fatalf("GetCertificate: %v", err)
	}
	if got.FqdnID != fqdn.ID {
		t.Errorf("FqdnID mismatch")
	}

	// GetCertificateBySerial
	bySerial, err := q.GetCertificateBySerial(ctx, models.GetCertificateBySerialParams{
		FqdnID: fqdn.ID, Serial: "01:02:03",
	})
	if err != nil {
		t.Fatalf("GetCertificateBySerial: %v", err)
	}
	if bySerial.ID != cert.ID {
		t.Errorf("GetCertificateBySerial ID mismatch")
	}

	// CountCertificates
	count, err := q.CountCertificates(ctx)
	if err != nil {
		t.Fatalf("CountCertificates: %v", err)
	}
	if count < 1 {
		t.Errorf("CountCertificates = %d, want >= 1", count)
	}

	// CountCertificatesByFQDN
	fqdnCount, err := q.CountCertificatesByFQDN(ctx, fqdn.ID)
	if err != nil {
		t.Fatalf("CountCertificatesByFQDN: %v", err)
	}
	if fqdnCount != 1 {
		t.Errorf("CountCertificatesByFQDN = %d, want 1", fqdnCount)
	}

	// ListCertificates
	certs, err := q.ListCertificates(ctx, models.ListCertificatesParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListCertificates: %v", err)
	}
	if len(certs) < 1 {
		t.Error("expected at least 1 certificate")
	}

	// ListCertificatesByFQDN
	fqdnCerts, err := q.ListCertificatesByFQDN(ctx, models.ListCertificatesByFQDNParams{
		FqdnID: fqdn.ID, Limit: 10, Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListCertificatesByFQDN: %v", err)
	}
	if len(fqdnCerts) != 1 {
		t.Errorf("ListCertificatesByFQDN: got %d, want 1", len(fqdnCerts))
	}

	// ListDistinctCAs
	cas, err := q.ListDistinctCAs(ctx)
	if err != nil {
		t.Fatalf("ListDistinctCAs: %v", err)
	}
	if len(cas) < 1 {
		t.Error("expected at least 1 distinct CA")
	}
}

func TestSettingsGetUpdate(t *testing.T) {
	q := openDB(t)
	ctx := context.Background()

	s, err := q.GetSettings(ctx)
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if s.Theme != "system" {
		t.Errorf("default theme = %q, want system", s.Theme)
	}

	updated, err := q.UpdateSettings(ctx, models.UpdateSettingsParams{
		AuthEnabled:               false,
		Username:                  nil,
		PasswordHash:              nil,
		ApiTokenProtectionEnabled: false,
		ApiToken:                  nil,
		Theme:                     "dark",
		SslmateApiKey:             "test-key",
		DefaultScheduleID:         nil,
	})
	if err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if updated.Theme != "dark" {
		t.Errorf("updated theme = %q, want dark", updated.Theme)
	}
	if updated.SslmateApiKey != "test-key" {
		t.Errorf("updated sslmate_api_key = %q, want test-key", updated.SslmateApiKey)
	}
}

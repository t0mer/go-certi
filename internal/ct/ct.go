package ct

import (
	"context"
	"fmt"
	"log/slog"
)

// Cert holds data extracted from a CT log entry.
type Cert struct {
	Serial    string
	IssuerCA  string
	SubjectCN string
	SANs      []string
	NotBefore string // RFC3339
	NotAfter  string // RFC3339
	Source    string // "sslmate" or "crt.sh"
}

// Client fetches CT log entries for an FQDN.
type Client struct {
	sslmateKey  string
	sslmateBase string
	crtshBase   string
	httpClient  interface{} // not used directly; clients use their own
}

// New creates a CT Client with production base URLs.
func New(sslmateAPIKey string) *Client {
	return &Client{
		sslmateKey:  sslmateAPIKey,
		sslmateBase: "https://api.certspotter.com/v1/issuances",
		crtshBase:   "https://crt.sh",
	}
}

// NewWithBaseURL creates a Client with configurable base URLs (for testing).
func NewWithBaseURL(sslmateKey, sslmateBaseURL, crtshBaseURL string) *Client {
	c := New(sslmateKey)
	if sslmateBaseURL != "" {
		c.sslmateBase = sslmateBaseURL
	}
	if crtshBaseURL != "" {
		c.crtshBase = crtshBaseURL
	}
	return c
}

// FetchCerts retrieves certificates for fqdn from sslmate (primary) with crt.sh fallback.
// sslmate is always tried first — it supports anonymous requests with reduced rate limits.
// The API key, if set, unlocks higher rate limits.
func (c *Client) FetchCerts(ctx context.Context, fqdn string, includeSubdomains bool) ([]Cert, error) {
	certs, err := fetchSslmate(ctx, c.sslmateKey, c.sslmateBase, fqdn, includeSubdomains)
	if err == nil {
		return certs, nil
	}
	slog.Warn("sslmate fetch failed, falling back to crt.sh", "fqdn", fqdn, "err", err)
	certs, err = fetchCrtsh(ctx, c.crtshBase, fqdn, includeSubdomains)
	if err != nil {
		return nil, fmt.Errorf("ct fetch: %w", err)
	}
	return certs, nil
}

package ct

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type sslmateEntry struct {
	TBSSha256 string   `json:"tbs_sha256"`
	DNSNames  []string `json:"dns_names"`
	Issuer    struct {
		Name string `json:"name"`
	} `json:"issuer"`
	NotBefore string `json:"not_before"`
	NotAfter  string `json:"not_after"`
}

func fetchSslmate(ctx context.Context, apiKey, baseURL, fqdn string, includeSubdomains bool) ([]Cert, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("domain", fqdn)
	q.Set("include_subdomains", fmt.Sprintf("%t", includeSubdomains))
	q.Set("expand", "dns_names")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("sslmate rate limit exceeded (429)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sslmate status %d", resp.StatusCode)
	}

	var entries []sslmateEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}

	certs := make([]Cert, 0, len(entries))
	for _, e := range entries {
		nb := normalizeTime(e.NotBefore)
		na := normalizeTime(e.NotAfter)
		cn := ""
		if len(e.DNSNames) > 0 {
			cn = e.DNSNames[0]
		}
		certs = append(certs, Cert{
			Serial:    e.TBSSha256,
			IssuerCA:  e.Issuer.Name,
			SubjectCN: cn,
			SANs:      e.DNSNames,
			NotBefore: nb,
			NotAfter:  na,
			Source:    "sslmate",
		})
	}
	return certs, nil
}

// normalizeTime parses common time formats and returns RFC3339.
func normalizeTime(s string) string {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC().Format(time.RFC3339)
		}
	}
	return s
}

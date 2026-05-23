package ct

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type crtshEntry struct {
	SerialNumber string `json:"serial_number"`
	IssuerName   string `json:"issuer_name"`
	CommonName   string `json:"common_name"`
	NameValue    string `json:"name_value"`
	NotBefore    string `json:"not_before"`
	NotAfter     string `json:"not_after"`
}

func fetchCrtsh(ctx context.Context, baseURL, fqdn string, includeSubdomains bool) ([]Cert, error) {
	q := fqdn
	if includeSubdomains {
		q = "%" + fqdn
	}
	u := fmt.Sprintf("%s/?q=%s&output=json", baseURL, url.QueryEscape(q))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crt.sh status %d", resp.StatusCode)
	}
	var entries []crtshEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	certs := make([]Cert, 0, len(entries))
	for _, e := range entries {
		if seen[e.SerialNumber] {
			continue
		}
		seen[e.SerialNumber] = true
		nb := normalizeTime(e.NotBefore)
		na := normalizeTime(e.NotAfter)
		sans := parseSANs(e.NameValue)
		certs = append(certs, Cert{
			Serial:    e.SerialNumber,
			IssuerCA:  e.IssuerName,
			SubjectCN: e.CommonName,
			SANs:      sans,
			NotBefore: nb,
			NotAfter:  na,
			Source:    "crt.sh",
		})
	}
	return certs, nil
}

func parseSANs(nameValue string) []string {
	parts := strings.Split(nameValue, "\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

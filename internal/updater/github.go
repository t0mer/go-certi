package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// Release describes a GitHub release relevant to the updater.
type Release struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	Body        string  `json:"body"`
	HTMLURL     string  `json:"html_url"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`
}

// Asset is a single downloadable file attached to a release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// LatestRelease fetches https://api.github.com/repos/{owner}/{repo}/releases/latest.
func LatestRelease(ctx context.Context, owner, repo string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned %d", resp.StatusCode)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// FindAssetForPlatform returns the asset matching the current OS+arch,
// or nil if no asset matches. Asset names follow the release-script
// pattern: go-certi_{version}_{os}_{display_arch}[.exe].
func (r *Release) FindAssetForPlatform() *Asset {
	osName, arch := runtime.GOOS, runtime.GOARCH
	wantExe := osName == "windows"
	for i := range r.Assets {
		n := strings.ToLower(r.Assets[i].Name)
		if !strings.Contains(n, osName) {
			continue
		}
		if !strings.Contains(n, arch) {
			continue
		}
		if wantExe != strings.HasSuffix(n, ".exe") {
			continue
		}
		return &r.Assets[i]
	}
	return nil
}

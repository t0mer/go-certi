package updater

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/t0mer/go-certi/internal/version"
)

const (
	githubOwner = "t0mer"
	githubRepo  = "go-certi"
)

// Status describes the result of a GitHub release check from the API's
// perspective. The frontend manages user preferences (skipped version,
// remind-later time, periodic-check toggle) in localStorage — this
// service is stateless on the server.
type Status struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	UpdateAvailable bool   `json:"update_available"`
	ReleaseURL      string `json:"release_url"`
	ReleaseNotes    string `json:"release_notes"`
	ReleaseName     string `json:"release_name"`
	PublishedAt     string `json:"published_at"`
}

// Service is a stateless wrapper around the GitHub-release fetch.
type Service struct{}

// New creates a Service. No dependencies — kept as a struct so the API
// handler signature can hold a pointer and check for nil.
func New() *Service {
	return &Service{}
}

// Check fetches the latest GitHub release and returns the comparison
// against the running binary version.
func (s *Service) Check(ctx context.Context) (*Status, error) {
	rel, err := LatestRelease(ctx, githubOwner, githubRepo)
	if err != nil {
		return nil, fmt.Errorf("fetch release: %w", err)
	}
	return &Status{
		CurrentVersion:  version.Version,
		LatestVersion:   rel.TagName,
		UpdateAvailable: Compare(version.Version, rel.TagName) < 0,
		ReleaseURL:      rel.HTMLURL,
		ReleaseNotes:    rel.Body,
		ReleaseName:     rel.Name,
		PublishedAt:     rel.PublishedAt,
	}, nil
}

// Apply downloads the release asset matching the current platform and
// replaces the running binary. Caller should call Restart() afterward.
func (s *Service) Apply(ctx context.Context) error {
	rel, err := LatestRelease(ctx, githubOwner, githubRepo)
	if err != nil {
		return err
	}
	asset := rel.FindAssetForPlatform()
	if asset == nil {
		return fmt.Errorf("no release asset available for this platform")
	}
	slog.Info("downloading update", "asset", asset.Name, "size", asset.Size)
	path, err := Download(ctx, asset)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	slog.Info("applying update", "path", path)
	if err := Apply(path); err != nil {
		return fmt.Errorf("apply: %w", err)
	}
	slog.Info("update applied; restart required")
	return nil
}

package notify

import (
	"context"
	"fmt"

	"github.com/containrrr/shoutrrr"
)

func dispatchShoutrrr(_ context.Context, configJSON, msg string) error {
	cfg, err := parseConfig(configJSON)
	if err != nil {
		return fmt.Errorf("shoutrrr: parse config: %w", err)
	}
	u, ok := cfg["url"]
	if !ok || u == "" {
		return fmt.Errorf("shoutrrr: missing 'url' in config")
	}
	return shoutrrr.Send(u, msg)
}

//go:build windows

package updater

import (
	"fmt"
	"os"
)

// On Windows the running .exe can't be overwritten. We move it aside
// to <binary>.old and rename the downloaded file into place. The .old
// file can be deleted on next startup.
func applyPlatform(current, downloaded string) error {
	old := current + ".old"
	_ = os.Remove(old)
	if err := os.Rename(current, old); err != nil {
		return fmt.Errorf("move old binary aside: %w", err)
	}
	if err := os.Rename(downloaded, current); err != nil {
		_ = os.Rename(old, current) // best-effort restore
		return fmt.Errorf("install new binary: %w", err)
	}
	return nil
}

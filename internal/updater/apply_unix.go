//go:build !windows

package updater

import "os"

// On Unix the running binary's inode stays open in memory; os.Rename
// atomically replaces the file. Future executions (incl. our re-exec)
// see the new binary.
func applyPlatform(current, downloaded string) error {
	return os.Rename(downloaded, current)
}

//go:build !windows

package updater

import (
	"os"
	"syscall"
)

// Restart re-execs the current binary in place. PID is preserved, which
// keeps systemd happy and works fine in interactive shells. Returns only
// on error; on success the process is replaced.
func Restart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return syscall.Exec(exe, os.Args, os.Environ())
}

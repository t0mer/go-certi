//go:build windows

package updater

import (
	"os"
	"os/exec"
)

// Restart spawns a fresh process with the same args and exits the current
// one. Works for both interactive runs and Windows services (the SCM sees
// the process exit and is configured to restart it).
func Restart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	os.Exit(0)
	return nil
}

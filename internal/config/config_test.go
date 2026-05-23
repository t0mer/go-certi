package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/t0mer/go-certi/internal/config"
)

func TestDefaultConfDir_XDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-test")
	got := config.DefaultConfDir()
	want := "/tmp/xdg-test/go-certi"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDefaultConfDir_FallsBackToHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()
	got := config.DefaultConfDir()
	want := filepath.Join(home, ".config", "go-certi")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestEnvOverride_ReplacesValue(t *testing.T) {
	t.Setenv("TEST_PORT", "9090")
	val := "8111"
	config.OverrideFromEnv("TEST_PORT", &val)
	if val != "9090" {
		t.Fatalf("expected env to override value, got %q", val)
	}
}

func TestEnvOverride_KeepsOriginalWhenUnset(t *testing.T) {
	t.Setenv("TEST_PORT_UNSET", "")
	val := "8111"
	config.OverrideFromEnv("TEST_PORT_UNSET", &val)
	if val != "8111" {
		t.Fatalf("expected value unchanged, got %q", val)
	}
}

func TestLoadOrCreate_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	confDir := filepath.Join(dir, "go-certi")
	cfg, err := config.LoadOrCreate(confDir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8111 {
		t.Fatalf("expected default port 8111, got %d", cfg.Port)
	}
	if _, err := os.Stat(confDir); os.IsNotExist(err) {
		t.Fatal("expected confDir to be created")
	}
}

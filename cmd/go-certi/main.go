package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/pflag"

	"github.com/t0mer/go-certi/internal/api"
	"github.com/t0mer/go-certi/internal/auth"
	"github.com/t0mer/go-certi/internal/config"
	"github.com/t0mer/go-certi/internal/db"
	"github.com/t0mer/go-certi/internal/models"
	"github.com/t0mer/go-certi/internal/version"
	webui "github.com/t0mer/go-certi/web"
)

func main() {
	// --- Flag definitions ---
	var (
		port        = pflag.IntP("port", "p", 8111, "Server port\n  env: GO_CERTI_PORT")
		confDir     = pflag.String("conf", config.DefaultConfDir(), "Config + DB directory\n  env: GO_CERTI_CONF")
		sslmateKey  = pflag.String("sslmate-api-key", "", "sslmate Cert Spotter API key\n  env: GO_CERTI_SSLMATE_API_KEY")
		resetPwd    = pflag.Bool("reset-password", false, "Generate new password, print plaintext, exit\n  env: GO_CERTI_RESET_PASSWORD")
		resetToken  = pflag.Bool("reset-api-token", false, "Generate new API token, print it, exit\n  env: GO_CERTI_RESET_API_TOKEN")
		showVersion = pflag.Bool("version", false, "Print version and exit")
	)

	pflag.CommandLine.SortFlags = false
	pflag.Parse()

	// --- Env overrides (env wins over CLI flag) ---
	// Documented precedence: env > flag. This is intentional and container-friendly.
	// See CLAUDE.md §8 for the rationale.
	overrides := []struct {
		env   string
		desc  string
		apply func(string)
	}{
		{"GO_CERTI_PORT", "port", func(v string) {
			if n, err := strconv.Atoi(v); err == nil {
				*port = n
			}
		}},
		{"GO_CERTI_CONF", "conf", func(v string) { *confDir = v }},
		{"GO_CERTI_SSLMATE_API_KEY", "sslmate-api-key", func(v string) { *sslmateKey = v }},
		{"GO_CERTI_RESET_PASSWORD", "reset-password", func(v string) {
			if v == "true" || v == "1" {
				*resetPwd = true
			}
		}},
		{"GO_CERTI_RESET_API_TOKEN", "reset-api-token", func(v string) {
			if v == "true" || v == "1" {
				*resetToken = true
			}
		}},
	}
	for _, o := range overrides {
		if v, ok := os.LookupEnv(o.env); ok && v != "" {
			slog.Debug("config source", "key", o.desc, "source", "env")
			o.apply(v)
		} else {
			slog.Debug("config source", "key", o.desc, "source", "flag")
		}
	}

	// --- Version flag ---
	if *showVersion {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	// --- Placeholder actions (implemented in auth plan) ---
	if *resetPwd || *resetToken {
		slog.Error("--reset-password and --reset-api-token not yet implemented")
		os.Exit(1)
	}

	// --- Load or create config ---
	cfg, err := config.LoadOrCreate(*confDir)
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	cfg.Port = *port
	_ = sslmateKey // stored to DB settings in a later plan

	// --- Open database ---
	dbConn, err := db.Open(filepath.Join(*confDir, "go-certi.db"))
	if err != nil {
		slog.Error("database", "err", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	q := models.New(dbConn)
	authSvc := auth.New("go-certi-jwt-secret") // TODO: load from settings/env

	// --- Start HTTP server ---
	srv := api.New(dbConn, q, authSvc, nil, nil, webui.FS())
	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("go-certi starting", "version", version.Version, "addr", addr, "conf", *confDir)

	if err := http.ListenAndServe(addr, srv); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}

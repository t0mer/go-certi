// @title           go-certi API
// @version         1.0
// @description     SSL Certificate Transparency monitor — REST API
// @host            localhost:8111
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     JWT token or opaque API token: "Bearer <token>"

package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/pflag"

	"github.com/t0mer/go-certi/internal/api"
	"github.com/t0mer/go-certi/internal/auth"
	"github.com/t0mer/go-certi/internal/config"
	"github.com/t0mer/go-certi/internal/ct"
	"github.com/t0mer/go-certi/internal/db"
	"github.com/t0mer/go-certi/internal/models"
	"github.com/t0mer/go-certi/internal/notify"
	"github.com/t0mer/go-certi/internal/scanner"
	"github.com/t0mer/go-certi/internal/scheduler"
	"github.com/t0mer/go-certi/internal/updater"
	"github.com/t0mer/go-certi/internal/version"
	webui "github.com/t0mer/go-certi/web"
)

func main() {
	// --- Flag definitions ---
	var (
		port          = pflag.IntP("port", "p", 8111, "Server port\n  env: GO_CERTI_PORT")
		confDir       = pflag.String("conf", config.DefaultConfDir(), "Config + DB directory\n  env: GO_CERTI_CONF")
		sslmateKey    = pflag.String("sslmate-api-key", "", "sslmate Cert Spotter API key\n  env: GO_CERTI_SSLMATE_API_KEY")
		resetPwd      = pflag.Bool("reset-password", false, "Generate new password, print plaintext, exit\n  env: GO_CERTI_RESET_PASSWORD")
		resetToken    = pflag.Bool("reset-api-token", false, "Generate new API token, print it, exit\n  env: GO_CERTI_RESET_API_TOKEN")
		serviceAction = pflag.String("service", "", "Manage as a system service: install | uninstall | start | stop | restart | status\n  env: GO_CERTI_SERVICE")
		showVersion   = pflag.Bool("version", false, "Print version and exit")
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
		{"GO_CERTI_SERVICE", "service", func(v string) { *serviceAction = v }},
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

	// --- Service management actions (install/uninstall/etc.) ---
	// Handled before loading config/DB so install works on a fresh host
	// without side-effects (no config dir is created on the installer's box;
	// the service will create it when it first runs).
	if *serviceAction != "" {
		if err := runServiceAction(*serviceAction, *confDir, *port); err != nil {
			slog.Error("service", "action", *serviceAction, "err", err)
			os.Exit(1)
		}
		return
	}

	// --- Load or create config ---
	cfg, err := config.LoadOrCreate(*confDir)
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	cfg.Port = *port

	// --- Open database ---
	dbConn, err := db.Open(filepath.Join(*confDir, "go-certi.db"))
	if err != nil {
		slog.Error("database", "err", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	q := models.New(dbConn)
	authSvc := auth.New("go-certi-jwt-secret") // TODO: load from settings/env

	// --- Reset password ---

	if *resetPwd {
		newPwd := generateRandomString(16)
		hash, err := authSvc.HashPassword(newPwd)
		if err != nil {
			slog.Error("hash password", "err", err)
			os.Exit(1)
		}
		existing, _ := q.GetSettings(context.Background())
		q.UpdateSettings(context.Background(), models.UpdateSettingsParams{ //nolint:errcheck
			AuthEnabled:               true,
			Username:                  existing.Username,
			PasswordHash:              &hash,
			ApiTokenProtectionEnabled: existing.ApiTokenProtectionEnabled,
			ApiToken:                  existing.ApiToken,
			Theme:                     existing.Theme,
			SslmateApiKey:             existing.SslmateApiKey,
			DefaultScheduleID:         existing.DefaultScheduleID,
		})
		fmt.Printf("New password: %s\n", newPwd)
		slog.Warn("password reset via CLI")
		os.Exit(0)
	}

	// --- Reset API token ---
	if *resetToken {
		tok, err := authSvc.GenerateAPIToken()
		if err != nil {
			slog.Error("generate token", "err", err)
			os.Exit(1)
		}
		existing, _ := q.GetSettings(context.Background())
		q.UpdateSettings(context.Background(), models.UpdateSettingsParams{ //nolint:errcheck
			AuthEnabled:               existing.AuthEnabled,
			Username:                  existing.Username,
			PasswordHash:              existing.PasswordHash,
			ApiTokenProtectionEnabled: true,
			ApiToken:                  &tok,
			Theme:                     existing.Theme,
			SslmateApiKey:             existing.SslmateApiKey,
			DefaultScheduleID:         existing.DefaultScheduleID,
		})
		fmt.Printf("New API token: %s\n", tok)
		slog.Warn("API token reset via CLI")
		os.Exit(0)
	}

	// --- Wire CT, notify, scanner, scheduler ---
	ctClient := ct.New(*sslmateKey)
	notifier := notify.New()
	scn := scanner.New(q, ctClient, notifier)
	sched := scheduler.New(q, scn)

	startCtx, startCancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := sched.Start(startCtx); err != nil {
		slog.Warn("scheduler start warning", "err", err)
	}
	startCancel()
	defer sched.Stop()

	// --- Start HTTP server (wrapped in service framework) ---
	upd := updater.New()
	srv := api.New(dbConn, q, authSvc, scn, notifier, upd, webui.FS())
	addr := fmt.Sprintf(":%d", cfg.Port)
	httpSrv := &http.Server{Addr: addr, Handler: srv}

	slog.Info("go-certi starting", "version", version.Version, "addr", addr, "conf", *confDir)

	// runUnderService blocks until the service is stopped — either by Ctrl+C
	// in interactive mode or by the OS service manager. It works identically
	// under systemd, Windows SCM, launchd, and plain terminal use.
	if err := runUnderService(httpSrv, *confDir, *port, func() { sched.Stop() }); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}

func generateRandomString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

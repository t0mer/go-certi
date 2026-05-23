package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/kardianos/service"
)

const (
	serviceName    = "go-certi"
	serviceDisplay = "go-certi"
	serviceDesc    = "SSL Certificate Transparency monitor"
)

// program adapts the HTTP server to the kardianos/service interface so the
// binary can run either interactively or under systemd / Windows SCM.
type program struct {
	httpSrv *http.Server
	onStop  func() // optional cleanup callback invoked before HTTP shutdown
}

// Start is called by the service manager (or by service.Run() in interactive
// mode). It must not block — start the real work in a goroutine.
func (p *program) Start(s service.Service) error {
	go func() {
		if err := p.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server", "err", err)
			// Signal the service framework to stop so it doesn't hang.
			_ = s.Stop()
		}
	}()
	return nil
}

// Stop is called when the OS or interactive signal asks the service to halt.
// It runs the user-supplied cleanup and gracefully drains the HTTP server.
func (p *program) Stop(_ service.Service) error {
	slog.Info("go-certi stopping")
	if p.onStop != nil {
		p.onStop()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.httpSrv.Shutdown(ctx)
}

// buildServiceConfig assembles the service config from the current binary
// location and the runtime --conf / --port values.
func buildServiceConfig(confDir string, port int) (*service.Config, error) {
	binary, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("locate binary: %w", err)
	}
	absConf, err := filepath.Abs(confDir)
	if err != nil {
		absConf = confDir
	}
	return &service.Config{
		Name:        serviceName,
		DisplayName: serviceDisplay,
		Description: serviceDesc,
		Executable:  binary,
		Arguments: []string{
			"--conf", absConf,
			"--port", strconv.Itoa(port),
		},
	}, nil
}

// runServiceAction handles `--service install|uninstall|start|stop|restart|status`.
// It does not load config or open the DB — those are not needed for service
// management, and skipping them avoids creating directories on hosts where
// the service hasn't run yet.
func runServiceAction(action, confDir string, port int) error {
	cfg, err := buildServiceConfig(confDir, port)
	if err != nil {
		return err
	}
	// A no-op program is sufficient for Install/Uninstall/Start/Stop — the
	// real program is only constructed when the service actually runs.
	noop := &program{httpSrv: &http.Server{}}
	svc, err := service.New(noop, cfg)
	if err != nil {
		return fmt.Errorf("service.New: %w", err)
	}

	switch action {
	case "install":
		if err := svc.Install(); err != nil {
			return fmt.Errorf("install: %w", err)
		}
		fmt.Printf("Service %q installed.\n", serviceName)
		fmt.Printf("  Binary: %s\n  Conf:   %s\n  Port:   %d\n", cfg.Executable, cfg.Arguments[1], port)
		if err := svc.Start(); err != nil {
			fmt.Printf("\nService installed but failed to start: %v\n", err)
			printManualStartHelp()
		} else {
			fmt.Println("\nService started.")
		}
		printPersistenceHelp()
		return nil

	case "uninstall":
		_ = svc.Stop() // best-effort
		if err := svc.Uninstall(); err != nil {
			return fmt.Errorf("uninstall: %w", err)
		}
		fmt.Printf("Service %q uninstalled.\n", serviceName)
		return nil

	case "start":
		if err := svc.Start(); err != nil {
			return fmt.Errorf("start: %w", err)
		}
		fmt.Println("Service started.")
		return nil

	case "stop":
		if err := svc.Stop(); err != nil {
			return fmt.Errorf("stop: %w", err)
		}
		fmt.Println("Service stopped.")
		return nil

	case "restart":
		if err := svc.Restart(); err != nil {
			return fmt.Errorf("restart: %w", err)
		}
		fmt.Println("Service restarted.")
		return nil

	case "status":
		st, err := svc.Status()
		if err != nil {
			return fmt.Errorf("status: %w", err)
		}
		fmt.Printf("Service %q status: %s\n", serviceName, statusString(st))
		return nil

	default:
		return fmt.Errorf("invalid --service value %q (use: install, uninstall, start, stop, restart, status)", action)
	}
}

// runUnderService wraps the HTTP server in the service framework so the same
// binary works correctly in interactive mode and under Windows SCM / systemd.
func runUnderService(httpSrv *http.Server, confDir string, port int, onStop func()) error {
	cfg, err := buildServiceConfig(confDir, port)
	if err != nil {
		return err
	}
	prg := &program{httpSrv: httpSrv, onStop: onStop}
	svc, err := service.New(prg, cfg)
	if err != nil {
		return fmt.Errorf("service.New: %w", err)
	}
	// In interactive mode this calls prg.Start() and blocks until Ctrl+C
	// (then prg.Stop()). Under SCM/systemd it handles the platform protocol.
	return svc.Run()
}

func statusString(s service.Status) string {
	switch s {
	case service.StatusRunning:
		return "running"
	case service.StatusStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

func printManualStartHelp() {
	fmt.Println("\nStart it manually:")
	switch service.ChosenSystem().String() {
	case "unix-systemv":
		fmt.Println("  service go-certi start")
	case "linux-systemd":
		fmt.Println("  sudo systemctl start go-certi")
		fmt.Println("  sudo systemctl enable go-certi   # auto-start on boot")
	case "windows-service":
		fmt.Println("  sc start go-certi")
	case "darwin-launchd":
		fmt.Println("  sudo launchctl load /Library/LaunchDaemons/go-certi.plist")
	default:
		fmt.Println("  (use your platform's service manager)")
	}
}

func printPersistenceHelp() {
	if service.ChosenSystem().String() == "linux-systemd" {
		fmt.Println("\nTo enable auto-start on boot:")
		fmt.Println("  sudo systemctl enable go-certi")
	}
}

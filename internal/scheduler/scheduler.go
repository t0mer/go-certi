package scheduler

import (
	"context"
	"log/slog"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/t0mer/go-certi/internal/models"
)

// Scannable can scan a single FQDN.
type Scannable interface {
	ScanFQDN(ctx context.Context, fqdn models.Fqdn) error
}

// Scheduler wraps robfig/cron and registers one job per enabled FQDN.
type Scheduler struct {
	cron    *cron.Cron
	q       *models.Queries
	scanner Scannable
	jobs    map[string]cron.EntryID
	mu      sync.Mutex
}

// New creates a Scheduler (not started yet).
func New(q *models.Queries, s Scannable) *Scheduler {
	return &Scheduler{
		cron:    cron.New(cron.WithSeconds()),
		q:       q,
		scanner: s,
		jobs:    make(map[string]cron.EntryID),
	}
}

// Start loads all enabled FQDNs and begins the cron loop.
func (s *Scheduler) Start(ctx context.Context) error {
	fqdns, err := s.q.ListEnabledFQDNs(ctx)
	if err != nil {
		return err
	}
	for _, f := range fqdns {
		if err := s.register(ctx, f); err != nil {
			slog.Error("scheduler: failed to register FQDN", "fqdn", f.Fqdn, "err", err)
		}
	}
	s.cron.Start()
	slog.Info("scheduler started", "jobs", len(s.jobs))
	return nil
}

// Stop gracefully shuts down the scheduler.
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// RegisterFQDN registers or re-registers a FQDN job (call after create/update).
func (s *Scheduler) RegisterFQDN(ctx context.Context, f models.Fqdn) error {
	s.mu.Lock()
	if id, ok := s.jobs[f.ID]; ok {
		s.cron.Remove(id)
		delete(s.jobs, f.ID)
	}
	s.mu.Unlock()
	if !f.Enabled {
		return nil
	}
	return s.register(ctx, f)
}

// DeregisterFQDN removes the scheduled job for a FQDN.
func (s *Scheduler) DeregisterFQDN(fqdnID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.jobs[fqdnID]; ok {
		s.cron.Remove(id)
		delete(s.jobs, fqdnID)
	}
}

// IsRunning reports whether the scheduler cron is active.
func (s *Scheduler) IsRunning() bool {
	return len(s.cron.Entries()) >= 0 // always true once started; used by readyz
}

func (s *Scheduler) register(ctx context.Context, f models.Fqdn) error {
	cronExpr := "@every 2h"
	if f.ScheduleID != nil {
		if sched, err := s.q.GetSchedule(ctx, *f.ScheduleID); err == nil {
			cronExpr = sched.CronExpr
		}
	} else {
		if defaultSched, err := s.q.GetDefaultSchedule(ctx); err == nil {
			cronExpr = defaultSched.CronExpr
		}
	}

	fqdnCopy := f
	id, err := s.cron.AddFunc(cronExpr, func() {
		if err := s.scanner.ScanFQDN(context.Background(), fqdnCopy); err != nil {
			slog.Error("scheduled scan failed", "fqdn", fqdnCopy.Fqdn, "err", err)
		}
	})
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.jobs[f.ID] = id
	s.mu.Unlock()
	slog.Info("registered fqdn in scheduler", "fqdn", f.Fqdn, "cron", cronExpr)
	return nil
}

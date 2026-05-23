package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/t0mer/go-certi/internal/auth"
	"github.com/t0mer/go-certi/internal/models"
	"github.com/t0mer/go-certi/internal/updater"
)

// ScannerInterface allows the handler to trigger scans.
type ScannerInterface interface {
	ScanFQDN(ctx context.Context, fqdn models.Fqdn) error
}

// NotifyInterface allows the handler to send test notifications.
type NotifyInterface interface {
	TestChannel(ctx context.Context, channel models.NotificationChannel) error
}

// Handler holds shared dependencies for all route handlers.
type Handler struct {
	q       *models.Queries
	db      *sql.DB
	authSvc *auth.Service
	scanner ScannerInterface
	notify  NotifyInterface
	updater *updater.Service
}

func newHandler(db *sql.DB, q *models.Queries, authSvc *auth.Service, scn ScannerInterface, notif NotifyInterface, upd *updater.Service) *Handler {
	return &Handler{q: q, db: db, authSvc: authSvc, scanner: scn, notify: notif, updater: upd}
}

func respondError(c *gin.Context, status int, msg string) {
	c.JSON(status, ErrorResponse{Error: msg})
}

func marshalConfig(cfg map[string]any) (string, error) {
	b, err := json.Marshal(cfg)
	return string(b), err
}

func unmarshalSANs(s string) []string {
	var v []string
	json.Unmarshal([]byte(s), &v) //nolint:errcheck
	if v == nil {
		return []string{}
	}
	return v
}

func isTrue(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

func notFoundOr500(c *gin.Context, err error) {
	if err == sql.ErrNoRows {
		respondError(c, http.StatusNotFound, "not found")
	} else {
		respondError(c, http.StatusInternalServerError, err.Error())
	}
}

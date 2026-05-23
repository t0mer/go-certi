package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/t0mer/go-certi/internal/updater"
)

// GetUpdateStatus returns the current version compared to the latest GitHub release.
// @Summary  Get update status
// @Tags     updates
// @Produce  json
// @Success  200  {object}  updater.Status
// @Router   /updates/status [get]
// @Security BearerAuth
func (h *Handler) GetUpdateStatus(c *gin.Context) {
	if h.updater == nil {
		respondError(c, http.StatusServiceUnavailable, "updater unavailable")
		return
	}
	st, err := h.updater.Check(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusBadGateway, "update check failed: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, st)
}

// ApplyUpdate downloads the latest release for this platform and restarts.
// @Summary  Download and apply the latest release, then restart
// @Tags     updates
// @Produce  json
// @Success  202
// @Router   /updates/apply [post]
// @Security BearerAuth
func (h *Handler) ApplyUpdate(c *gin.Context) {
	if h.updater == nil {
		respondError(c, http.StatusServiceUnavailable, "updater unavailable")
		return
	}
	if err := h.updater.Apply(c.Request.Context()); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	// Respond before restarting so the HTTP response can flush.
	c.JSON(http.StatusAccepted, gin.H{"status": "update applied, restarting"})

	go func() {
		time.Sleep(500 * time.Millisecond)
		_ = updater.Restart()
	}()
}

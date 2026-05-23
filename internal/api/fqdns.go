package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/t0mer/go-certi/internal/models"
)

func (h *Handler) buildFQDNResponse(c *gin.Context, f models.Fqdn) (FQDNResponse, error) {
	channelIDs, err := h.q.GetFQDNChannelIDs(c.Request.Context(), f.ID)
	if err != nil {
		return FQDNResponse{}, err
	}
	ids := make([]string, len(channelIDs))
	for i, id := range channelIDs {
		ids[i] = id
	}
	return FQDNResponse{
		ID:                   f.ID,
		FQDN:                 f.Fqdn,
		IncludeSubdomains:    f.IncludeSubdomains,
		Enabled:              f.Enabled,
		NotificationsEnabled: f.NotificationsEnabled,
		ScheduleID:           f.ScheduleID,
		ChannelIDs:           ids,
		CreatedAt:            f.CreatedAt,
		UpdatedAt:            f.UpdatedAt,
	}, nil
}

// ListFQDNs godoc
// @Summary List all FQDNs
// @Tags fqdns
// @Produce json
// @Success 200 {array} FQDNResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/fqdns [get]
func (h *Handler) ListFQDNs(c *gin.Context) {
	fqdns, err := h.q.ListFQDNs(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]FQDNResponse, 0, len(fqdns))
	for _, f := range fqdns {
		r, err := h.buildFQDNResponse(c, f)
		if err != nil {
			respondError(c, http.StatusInternalServerError, err.Error())
			return
		}
		resp = append(resp, r)
	}
	c.JSON(http.StatusOK, resp)
}

// GetFQDN godoc
// @Summary Get a single FQDN by ID
// @Tags fqdns
// @Produce json
// @Param id path string true "FQDN ID"
// @Success 200 {object} FQDNResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/fqdns/{id} [get]
func (h *Handler) GetFQDN(c *gin.Context) {
	f, err := h.q.GetFQDN(c.Request.Context(), c.Param("id"))
	if err != nil {
		notFoundOr500(c, err)
		return
	}
	resp, err := h.buildFQDNResponse(c, f)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateFQDN godoc
// @Summary Create a new FQDN to monitor
// @Tags fqdns
// @Accept json
// @Produce json
// @Param body body CreateFQDNRequest true "FQDN to create"
// @Success 201 {object} FQDNResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/fqdns [post]
func (h *Handler) CreateFQDN(c *gin.Context) {
	var req CreateFQDNRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	f, err := h.q.CreateFQDN(c.Request.Context(), models.CreateFQDNParams{
		ID:                   uuid.NewString(),
		Fqdn:                 req.FQDN,
		IncludeSubdomains:    req.IncludeSubdomains,
		Enabled:              isTrue(req.Enabled, true),
		NotificationsEnabled: isTrue(req.NotificationsEnabled, true),
		ScheduleID:           req.ScheduleID,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.setFQDNChannels(c, f.ID, req.ChannelIDs); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp, _ := h.buildFQDNResponse(c, f)
	c.JSON(http.StatusCreated, resp)
}

// UpdateFQDN godoc
// @Summary Update an existing FQDN
// @Tags fqdns
// @Accept json
// @Produce json
// @Param id path string true "FQDN ID"
// @Param body body UpdateFQDNRequest true "Updated FQDN data"
// @Success 200 {object} FQDNResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/fqdns/{id} [put]
func (h *Handler) UpdateFQDN(c *gin.Context) {
	var req UpdateFQDNRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	f, err := h.q.UpdateFQDN(c.Request.Context(), models.UpdateFQDNParams{
		Fqdn:                 req.FQDN,
		IncludeSubdomains:    req.IncludeSubdomains,
		Enabled:              req.Enabled,
		NotificationsEnabled: req.NotificationsEnabled,
		ScheduleID:           req.ScheduleID,
		ID:                   c.Param("id"),
	})
	if err != nil {
		notFoundOr500(c, err)
		return
	}
	if err := h.setFQDNChannels(c, f.ID, req.ChannelIDs); err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp, _ := h.buildFQDNResponse(c, f)
	c.JSON(http.StatusOK, resp)
}

// DeleteFQDN godoc
// @Summary Delete an FQDN
// @Tags fqdns
// @Param id path string true "FQDN ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/fqdns/{id} [delete]
func (h *Handler) DeleteFQDN(c *gin.Context) {
	if err := h.q.DeleteFQDN(c.Request.Context(), c.Param("id")); err != nil {
		notFoundOr500(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// TriggerScan godoc
// @Summary Trigger an immediate certificate scan for an FQDN
// @Tags fqdns
// @Param id path string true "FQDN ID"
// @Success 202 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/fqdns/{id}/scan [post]
func (h *Handler) TriggerScan(c *gin.Context) {
	f, err := h.q.GetFQDN(c.Request.Context(), c.Param("id"))
	if err != nil {
		notFoundOr500(c, err)
		return
	}
	if h.scanner != nil {
		go func() {
			if err := h.scanner.ScanFQDN(context.Background(), f); err != nil {
				slog.Error("trigger scan failed", "fqdn", f.Fqdn, "err", err)
			}
		}()
	}
	c.JSON(http.StatusAccepted, gin.H{"status": "scan started"})
}

// setFQDNChannels replaces all channel associations for an FQDN.
func (h *Handler) setFQDNChannels(c *gin.Context, fqdnID string, channelIDs []string) error {
	ctx := c.Request.Context()
	if err := h.q.DeleteFQDNChannels(ctx, fqdnID); err != nil {
		return err
	}
	for _, cid := range channelIDs {
		if err := h.q.AddFQDNChannel(ctx, models.AddFQDNChannelParams{
			FqdnID: fqdnID, ChannelID: cid,
		}); err != nil {
			return err
		}
	}
	return nil
}

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/t0mer/go-certi/internal/models"
)

// ListChannels godoc
// @Summary List all notification channels
// @Tags channels
// @Produce json
// @Success 200 {array} models.NotificationChannel
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/channels [get]
func (h *Handler) ListChannels(c *gin.Context) {
	chs, err := h.q.ListChannels(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if chs == nil {
		chs = []models.NotificationChannel{}
	}
	c.JSON(http.StatusOK, chs)
}

// GetChannel godoc
// @Summary Get a single notification channel by ID
// @Tags channels
// @Produce json
// @Param id path string true "Channel ID"
// @Success 200 {object} models.NotificationChannel
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/channels/{id} [get]
func (h *Handler) GetChannel(c *gin.Context) {
	ch, err := h.q.GetChannel(c.Request.Context(), c.Param("id"))
	if err != nil {
		notFoundOr500(c, err)
		return
	}
	c.JSON(http.StatusOK, ch)
}

// CreateChannel godoc
// @Summary Create a new notification channel
// @Tags channels
// @Accept json
// @Produce json
// @Param body body CreateChannelRequest true "Channel to create"
// @Success 201 {object} models.NotificationChannel
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/channels [post]
func (h *Handler) CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	cfgStr, err := marshalConfig(req.Config)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid config JSON")
		return
	}
	ch, err := h.q.CreateChannel(c.Request.Context(), models.CreateChannelParams{
		ID:      uuid.NewString(),
		Name:    req.Name,
		Type:    req.Type,
		Config:  cfgStr,
		Enabled: isTrue(req.Enabled, true),
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, ch)
}

// UpdateChannel godoc
// @Summary Update an existing notification channel
// @Tags channels
// @Accept json
// @Produce json
// @Param id path string true "Channel ID"
// @Param body body UpdateChannelRequest true "Updated channel data"
// @Success 200 {object} models.NotificationChannel
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/channels/{id} [put]
func (h *Handler) UpdateChannel(c *gin.Context) {
	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	cfgStr, err := marshalConfig(req.Config)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid config JSON")
		return
	}
	ch, err := h.q.UpdateChannel(c.Request.Context(), models.UpdateChannelParams{
		Name:    req.Name,
		Type:    req.Type,
		Config:  cfgStr,
		Enabled: req.Enabled,
		ID:      c.Param("id"),
	})
	if err != nil {
		notFoundOr500(c, err)
		return
	}
	c.JSON(http.StatusOK, ch)
}

// DeleteChannel godoc
// @Summary Delete a notification channel
// @Tags channels
// @Param id path string true "Channel ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/channels/{id} [delete]
func (h *Handler) DeleteChannel(c *gin.Context) {
	if err := h.q.DeleteChannel(c.Request.Context(), c.Param("id")); err != nil {
		notFoundOr500(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// TestChannel godoc
// @Summary Send a test notification via the specified channel
// @Tags channels
// @Param id path string true "Channel ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Failure 502 {object} ErrorResponse
// @Router /api/v1/channels/{id}/test [post]
func (h *Handler) TestChannel(c *gin.Context) {
	ch, err := h.q.GetChannel(c.Request.Context(), c.Param("id"))
	if err != nil {
		notFoundOr500(c, err)
		return
	}
	if h.notify != nil {
		if err := h.notify.TestChannel(c.Request.Context(), ch); err != nil {
			respondError(c, http.StatusBadGateway, "notification failed: "+err.Error())
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "test notification sent"})
}

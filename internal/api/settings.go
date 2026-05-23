package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/t0mer/go-certi/internal/models"
)

// GetSettings godoc
// @Summary Get application settings (secrets are omitted)
// @Tags settings
// @Produce json
// @Success 200 {object} SettingsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/settings [get]
func (h *Handler) GetSettings(c *gin.Context) {
	s, err := h.q.GetSettings(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, settingsToResponse(s))
}

// UpdateSettings godoc
// @Summary Update application settings
// @Tags settings
// @Accept json
// @Produce json
// @Param body body UpdateSettingsRequest true "Settings to update"
// @Success 200 {object} SettingsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/settings [put]
func (h *Handler) UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	existing, err := h.q.GetSettings(ctx)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	params := models.UpdateSettingsParams{
		AuthEnabled:               req.AuthEnabled,
		Username:                  req.Username,
		PasswordHash:              existing.PasswordHash,
		ApiTokenProtectionEnabled: req.APITokenProtectionEnabled,
		ApiToken:                  existing.ApiToken,
		Theme:                     req.Theme,
		SslmateApiKey:             req.SslmateAPIKey,
		DefaultScheduleID:         req.DefaultScheduleID,
	}
	if req.Theme == "" {
		params.Theme = existing.Theme
	}
	if req.Password != nil && *req.Password != "" {
		hash, err := h.authSvc.HashPassword(*req.Password)
		if err != nil {
			respondError(c, http.StatusInternalServerError, "failed to hash password")
			return
		}
		params.PasswordHash = &hash
	}
	updated, err := h.q.UpdateSettings(ctx, params)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, settingsToResponse(updated))
}

// RotateAPIToken godoc
// @Summary Generate and store a new API token (only time it is shown in plaintext)
// @Tags settings
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/settings/api-token/rotate [post]
func (h *Handler) RotateAPIToken(c *gin.Context) {
	tok, err := h.authSvc.GenerateAPIToken()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to generate token")
		return
	}
	ctx := c.Request.Context()
	existing, err := h.q.GetSettings(ctx)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	_, err = h.q.UpdateSettings(ctx, models.UpdateSettingsParams{ //nolint:errcheck
		AuthEnabled:               existing.AuthEnabled,
		Username:                  existing.Username,
		PasswordHash:              existing.PasswordHash,
		ApiTokenProtectionEnabled: existing.ApiTokenProtectionEnabled,
		ApiToken:                  &tok,
		Theme:                     existing.Theme,
		SslmateApiKey:             existing.SslmateApiKey,
		DefaultScheduleID:         existing.DefaultScheduleID,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tok})
}

func settingsToResponse(s models.Setting) SettingsResponse {
	return SettingsResponse{
		AuthEnabled:               s.AuthEnabled,
		Username:                  s.Username,
		APITokenProtectionEnabled: s.ApiTokenProtectionEnabled,
		Theme:                     s.Theme,
		SslmateAPIKey:             s.SslmateApiKey,
		DefaultScheduleID:         s.DefaultScheduleID,
	}
}

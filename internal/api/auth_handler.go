package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Login godoc
// @Summary Authenticate with username and password, receive a JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	s, err := h.q.GetSettings(ctx)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if !s.AuthEnabled {
		respondError(c, http.StatusBadRequest, "auth not enabled")
		return
	}
	if s.Username == nil || *s.Username != req.Username {
		respondError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if s.PasswordHash == nil || !h.authSvc.CheckPassword(req.Password, *s.PasswordHash) {
		respondError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, err := h.authSvc.IssueJWT(req.Username)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to issue token")
		return
	}
	c.SetCookie(jwtCookieName, token, int(24*time.Hour/time.Second), "/", "", false, true)
	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

// Logout godoc
// @Summary Clear the auth cookie
// @Tags auth
// @Success 204
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie(jwtCookieName, "", -1, "/", "", false, true)
	c.Status(http.StatusNoContent)
}

// Me godoc
// @Summary Return the currently authenticated username
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	tok, err := c.Cookie(jwtCookieName)
	if err != nil {
		hdr := c.GetHeader("Authorization")
		tok = strings.TrimPrefix(hdr, "Bearer ")
	}
	subject, err := h.authSvc.VerifyJWT(tok)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	c.JSON(http.StatusOK, gin.H{"username": subject})
}

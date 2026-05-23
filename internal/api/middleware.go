package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/t0mer/go-certi/internal/models"
)

const jwtCookieName = "go_certi_token"

// AuthRequired returns a Gin middleware that enforces auth when enabled in settings.
func (h *Handler) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		settings, err := h.q.GetSettings(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: "settings unavailable"})
			return
		}

		if settings.ApiTokenProtectionEnabled {
			if !h.checkBearer(c, settings) && !h.checkJWT(c) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
				return
			}
		} else if settings.AuthEnabled {
			if !h.checkJWT(c) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
				return
			}
		}
		c.Next()
	}
}

func (h *Handler) checkBearer(c *gin.Context, s models.Setting) bool {
	hdr := c.GetHeader("Authorization")
	if !strings.HasPrefix(hdr, "Bearer ") {
		return false
	}
	token := strings.TrimPrefix(hdr, "Bearer ")
	if s.ApiToken != nil && token == *s.ApiToken {
		return true
	}
	return h.verifyJWTStr(token)
}

func (h *Handler) checkJWT(c *gin.Context) bool {
	tok, err := c.Cookie(jwtCookieName)
	if err != nil {
		hdr := c.GetHeader("Authorization")
		tok = strings.TrimPrefix(hdr, "Bearer ")
	}
	return h.verifyJWTStr(tok)
}

func (h *Handler) verifyJWTStr(tok string) bool {
	if tok == "" {
		return false
	}
	_, err := h.authSvc.VerifyJWT(tok)
	return err == nil
}

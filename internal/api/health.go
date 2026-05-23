package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// healthz is always 200 — no auth, no DB check. Used by Docker healthchecks.
func (s *Server) healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// readyz checks the DB is reachable and returns 200/503.
func (s *Server) readyz(c *gin.Context) {
	if err := s.db.PingContext(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db unreachable", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

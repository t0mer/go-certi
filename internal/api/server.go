package api

import (
	"database/sql"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Server holds the Gin engine and shared dependencies.
type Server struct {
	engine *gin.Engine
	db     *sql.DB
}

// New constructs the server and registers all routes.
// webFS is an fs.FS rooted at the compiled frontend directory (web/dist).
// Pass nil to skip static file serving (useful in tests).
func New(db *sql.DB, webFS fs.FS) *Server {
	gin.SetMode(gin.ReleaseMode)
	s := &Server{engine: gin.New(), db: db}
	s.engine.Use(gin.Recovery())

	// Health — always unauthenticated
	s.engine.GET("/healthz", s.healthz)
	s.engine.GET("/readyz", s.readyz)

	// Frontend — serve embedded web/dist when provided
	if webFS != nil {
		s.engine.NoRoute(gin.WrapH(http.FileServer(http.FS(webFS))))
	}

	return s
}

// ServeHTTP implements http.Handler so the server can be used with httptest.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.ServeHTTP(w, r)
}

// Engine exposes the underlying Gin engine for route registration in main.
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

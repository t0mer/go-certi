package api

import (
	"database/sql"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/t0mer/go-certi/internal/auth"
	"github.com/t0mer/go-certi/internal/models"

	_ "github.com/t0mer/go-certi/docs"
)

// Server holds the Gin engine and shared dependencies.
type Server struct {
	engine *gin.Engine
	db     *sql.DB
}

// New constructs the server with all routes registered.
// Pass nil for scn/notif during bootstrap (stubs); they'll be wired in CT plan.
func New(db *sql.DB, q *models.Queries, authSvc *auth.Service, scn ScannerInterface, notif NotifyInterface, webFS fs.FS) *Server {
	gin.SetMode(gin.ReleaseMode)
	s := &Server{engine: gin.New(), db: db}
	s.engine.Use(gin.Recovery())

	h := newHandler(db, q, authSvc, scn, notif)

	// Health — always unauthenticated
	s.engine.GET("/healthz", s.healthz)
	s.engine.GET("/readyz", s.readyz)

	// Swagger UI
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1
	api := s.engine.Group("/api/v1")

	// Auth endpoints (no auth middleware)
	authG := api.Group("/auth")
	authG.POST("/login", h.Login)
	authG.POST("/logout", h.Logout)
	authG.GET("/me", h.AuthRequired(), h.Me)

	// Protected endpoints
	protected := api.Group("")
	protected.Use(h.AuthRequired())

	// FQDNs
	protected.GET("/fqdns", h.ListFQDNs)
	protected.POST("/fqdns", h.CreateFQDN)
	protected.GET("/fqdns/:id", h.GetFQDN)
	protected.PUT("/fqdns/:id", h.UpdateFQDN)
	protected.DELETE("/fqdns/:id", h.DeleteFQDN)
	protected.POST("/fqdns/:id/scan", h.TriggerScan)

	// Certificates — register /cas BEFORE /:id to avoid routing conflict
	protected.GET("/certificates/cas", h.ListCAs)
	protected.GET("/certificates", h.ListCertificates)
	protected.GET("/certificates/:id", h.GetCertificate)

	// Channels
	protected.GET("/channels", h.ListChannels)
	protected.POST("/channels", h.CreateChannel)
	protected.GET("/channels/:id", h.GetChannel)
	protected.PUT("/channels/:id", h.UpdateChannel)
	protected.DELETE("/channels/:id", h.DeleteChannel)
	protected.POST("/channels/:id/test", h.TestChannel)

	// Schedules
	protected.GET("/schedules", h.ListSchedules)
	protected.POST("/schedules", h.CreateSchedule)
	protected.GET("/schedules/:id", h.GetSchedule)
	protected.PUT("/schedules/:id", h.UpdateSchedule)
	protected.DELETE("/schedules/:id", h.DeleteSchedule)

	// Settings
	protected.GET("/settings", h.GetSettings)
	protected.PUT("/settings", h.UpdateSettings)
	protected.POST("/settings/api-token/rotate", h.RotateAPIToken)

	// Frontend — serve embedded web/dist
	if webFS != nil {
		s.engine.NoRoute(gin.WrapH(http.FileServer(http.FS(webFS))))
	}

	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.ServeHTTP(w, r)
}

// Engine exposes the underlying Gin engine.
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

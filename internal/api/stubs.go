package api

import "github.com/gin-gonic/gin"

func (h *Handler) ListFQDNs(c *gin.Context)        {}
func (h *Handler) GetFQDN(c *gin.Context)           {}
func (h *Handler) CreateFQDN(c *gin.Context)        {}
func (h *Handler) UpdateFQDN(c *gin.Context)        {}
func (h *Handler) DeleteFQDN(c *gin.Context)        {}
func (h *Handler) TriggerScan(c *gin.Context)       {}
func (h *Handler) ListCertificates(c *gin.Context)  {}
func (h *Handler) GetCertificate(c *gin.Context)    {}
func (h *Handler) ListCAs(c *gin.Context)           {}
func (h *Handler) ListChannels(c *gin.Context)      {}
func (h *Handler) GetChannel(c *gin.Context)        {}
func (h *Handler) CreateChannel(c *gin.Context)     {}
func (h *Handler) UpdateChannel(c *gin.Context)     {}
func (h *Handler) DeleteChannel(c *gin.Context)     {}
func (h *Handler) TestChannel(c *gin.Context)       {}
func (h *Handler) ListSchedules(c *gin.Context)     {}
func (h *Handler) GetSchedule(c *gin.Context)       {}
func (h *Handler) CreateSchedule(c *gin.Context)    {}
func (h *Handler) UpdateSchedule(c *gin.Context)    {}
func (h *Handler) DeleteSchedule(c *gin.Context)    {}
func (h *Handler) GetSettings(c *gin.Context)       {}
func (h *Handler) UpdateSettings(c *gin.Context)    {}
func (h *Handler) RotateAPIToken(c *gin.Context)    {}
func (h *Handler) Login(c *gin.Context)             {}
func (h *Handler) Logout(c *gin.Context)            {}
func (h *Handler) Me(c *gin.Context)                {}

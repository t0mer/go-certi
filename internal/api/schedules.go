package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/t0mer/go-certi/internal/models"
)

// ListSchedules godoc
// @Summary List all schedules
// @Tags schedules
// @Produce json
// @Success 200 {array} models.Schedule
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/schedules [get]
func (h *Handler) ListSchedules(c *gin.Context) {
	scheds, err := h.q.ListSchedules(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if scheds == nil {
		scheds = []models.Schedule{}
	}
	c.JSON(http.StatusOK, scheds)
}

// GetSchedule godoc
// @Summary Get a single schedule by ID
// @Tags schedules
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 200 {object} models.Schedule
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/schedules/{id} [get]
func (h *Handler) GetSchedule(c *gin.Context) {
	s, err := h.q.GetSchedule(c.Request.Context(), c.Param("id"))
	if err != nil {
		notFoundOr500(c, err)
		return
	}
	c.JSON(http.StatusOK, s)
}

// CreateSchedule godoc
// @Summary Create a new schedule
// @Tags schedules
// @Accept json
// @Produce json
// @Param body body CreateScheduleRequest true "Schedule to create"
// @Success 201 {object} models.Schedule
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/schedules [post]
func (h *Handler) CreateSchedule(c *gin.Context) {
	var req CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	if req.IsDefault {
		h.q.UnsetDefaultSchedules(ctx) //nolint:errcheck
	}
	s, err := h.q.CreateSchedule(ctx, models.CreateScheduleParams{
		ID:        uuid.NewString(),
		Name:      req.Name,
		CronExpr:  req.CronExpr,
		IsDefault: req.IsDefault,
		Enabled:   isTrue(req.Enabled, true),
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, s)
}

// UpdateSchedule godoc
// @Summary Update an existing schedule
// @Tags schedules
// @Accept json
// @Produce json
// @Param id path string true "Schedule ID"
// @Param body body UpdateScheduleRequest true "Updated schedule data"
// @Success 200 {object} models.Schedule
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/schedules/{id} [put]
func (h *Handler) UpdateSchedule(c *gin.Context) {
	var req UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	ctx := c.Request.Context()
	if req.IsDefault {
		h.q.UnsetDefaultSchedules(ctx) //nolint:errcheck
	}
	s, err := h.q.UpdateSchedule(ctx, models.UpdateScheduleParams{
		Name:      req.Name,
		CronExpr:  req.CronExpr,
		IsDefault: req.IsDefault,
		Enabled:   req.Enabled,
		ID:        c.Param("id"),
	})
	if err != nil {
		notFoundOr500(c, err)
		return
	}
	c.JSON(http.StatusOK, s)
}

// DeleteSchedule godoc
// @Summary Delete a schedule
// @Tags schedules
// @Param id path string true "Schedule ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/schedules/{id} [delete]
func (h *Handler) DeleteSchedule(c *gin.Context) {
	if err := h.q.DeleteSchedule(c.Request.Context(), c.Param("id")); err != nil {
		notFoundOr500(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

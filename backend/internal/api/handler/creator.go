package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service"
)

// Creator 达人分析 HTTP 处理器
type Creator struct {
	svc *service.CreatorService
}

func NewCreator(svc *service.CreatorService) *Creator {
	return &Creator{svc: svc}
}

// Kpi GET /api/creator/kpi
func (h *Creator) Kpi(c *gin.Context) {
	data, err := h.svc.Kpi(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// Trend GET /api/creator/trend
func (h *Creator) Trend(c *gin.Context) {
	data, err := h.svc.Trend(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// Platforms GET /api/creator/platforms
func (h *Creator) Platforms(c *gin.Context) {
	data, err := h.svc.Platforms(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// Tracks GET /api/creator/tracks
func (h *Creator) Tracks(c *gin.Context) {
	data, err := h.svc.Tracks(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// Audience GET /api/creator/audience
func (h *Creator) Audience(c *gin.Context) {
	data, err := h.svc.Audience(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// List GET /api/creator/list
func (h *Creator) List(c *gin.Context) {
	data, err := h.svc.List(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

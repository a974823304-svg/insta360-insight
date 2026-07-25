package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service"
)

// Insight 数据洞察 HTTP 处理器
type Insight struct {
	svc *service.InsightService
	ai  *service.AIService
}

func NewInsight(svc *service.InsightService, ai *service.AIService) *Insight {
	return &Insight{svc: svc, ai: ai}
}

// bindFilter 解析 querystring 中的筛选条件。
func bindFilter(c *gin.Context) model.Filter {
	return model.Filter{
		DateRange: c.QueryArray("date_range"),
		Regions:   c.QueryArray("regions"),
		Tracks:    c.QueryArray("tracks"),
		Platforms: c.QueryArray("platforms"),
		AgeBands:  c.QueryArray("age_bands"),
	}
}

func fail(c *gin.Context, err error) {
	c.JSON(http.StatusBadGateway, model.Fail(500, err.Error()))
}

// Kpi GET /api/kpi
func (h *Insight) Kpi(c *gin.Context) {
	data, err := h.svc.Kpi(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// ViewsTrend GET /api/views-trend
func (h *Insight) ViewsTrend(c *gin.Context) {
	data, err := h.svc.ViewsTrend(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// PlatformShare GET /api/platform-distribution
func (h *Insight) PlatformShare(c *gin.Context) {
	data, err := h.svc.PlatformShare(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// TrackPerformance GET /api/track-performance
func (h *Insight) TrackPerformance(c *gin.Context) {
	data, err := h.svc.TrackPerformance(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// Radar GET /api/explosive-radar
func (h *Insight) Radar(c *gin.Context) {
	data, err := h.svc.Radar(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// AudienceAge GET /api/audience-age
func (h *Insight) AudienceAge(c *gin.Context) {
	data, err := h.svc.AudienceAge(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// TopCreators GET /api/top-creators
func (h *Insight) TopCreators(c *gin.Context) {
	data, err := h.svc.TopCreators(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// Insights GET /api/insights —— AI 关键洞察
func (h *Insight) Insights(c *gin.Context) {
	data, err := h.ai.Generate(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

// FilterOptions GET /api/filters/options
func (h *Insight) FilterOptions(c *gin.Context) {
	data, err := h.svc.Options(c.Request.Context(), bindFilter(c))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, model.OK(data))
}

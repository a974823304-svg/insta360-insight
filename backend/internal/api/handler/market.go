package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service"
)

type Market struct{ svc *service.MarketService }

func NewMarket(svc *service.MarketService) *Market { return &Market{svc: svc} }

func (h *Market) Kpi(c *gin.Context) {
	d, e := h.svc.Kpi(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Market) Trend(c *gin.Context) {
	d, e := h.svc.Trend(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Market) Competitors(c *gin.Context) {
	d, e := h.svc.Competitors(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Market) Regions(c *gin.Context) {
	d, e := h.svc.Regions(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Market) Prices(c *gin.Context) {
	d, e := h.svc.Prices(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Market) List(c *gin.Context) {
	d, e := h.svc.List(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}

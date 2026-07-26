package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service"
)

type Brand struct{ svc *service.BrandService }

func NewBrand(svc *service.BrandService) *Brand { return &Brand{svc: svc} }

func (h *Brand) Kpi(c *gin.Context) {
	d, e := h.svc.Kpi(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Brand) Trend(c *gin.Context) {
	d, e := h.svc.Trend(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Brand) Platforms(c *gin.Context) {
	d, e := h.svc.Platforms(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Brand) Sentiment(c *gin.Context) {
	d, e := h.svc.Sentiment(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Brand) Keywords(c *gin.Context) {
	d, e := h.svc.Keywords(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Brand) List(c *gin.Context) {
	d, e := h.svc.List(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}

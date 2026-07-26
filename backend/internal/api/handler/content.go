package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service"
)

type Content struct{ svc *service.ContentService }

func NewContent(svc *service.ContentService) *Content { return &Content{svc: svc} }

func (h *Content) Kpi(c *gin.Context) {
	d, e := h.svc.Kpi(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Content) Trend(c *gin.Context) {
	d, e := h.svc.Trend(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Content) Forms(c *gin.Context) {
	d, e := h.svc.Forms(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Content) Topics(c *gin.Context) {
	d, e := h.svc.Topics(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Content) Durations(c *gin.Context) {
	d, e := h.svc.Durations(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}
func (h *Content) List(c *gin.Context) {
	d, e := h.svc.List(c.Request.Context(), bindFilter(c))
	if e != nil {
		fail(c, e)
		return
	}
	c.JSON(http.StatusOK, model.OK(d))
}

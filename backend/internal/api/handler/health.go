// Package handler 存放 HTTP 处理器(handler),负责参数解析 + 调用 service + 包装响应。
// 业务逻辑全部下沉到 service 层,handler 自身保持"瘦"。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
)

// Health 简单健康检查
type Health struct{}

func NewHealth() *Health { return &Health{} }

// Handle GET /api/health
func (h *Health) Handle(c *gin.Context) {
	c.JSON(http.StatusOK, model.OK(gin.H{
		"status":  "ok",
		"service": "insta360-insight-backend",
	}))
}

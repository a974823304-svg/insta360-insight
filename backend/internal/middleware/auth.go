package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service"
)

// TokenValidator 是 JWTAuth 依赖的最小接口,便于测试与解耦。
type TokenValidator interface {
	ParseToken(token string) (*service.Claims, error)
}

// JWTAuth 解析 Bearer token 并校验;disable 时注入 devUser(本地调试,须为真实账号身份)。
func JWTAuth(v TokenValidator, disable bool, devUser service.Claims) gin.HandlerFunc {
	return func(c *gin.Context) {
		if disable {
			c.Set("claims", devUser)
			c.Next()
			return
		}
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.Fail(401, "未提供有效令牌"))
			return
		}
		claims, err := v.ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.Fail(401, "令牌无效或已过期"))
			return
		}
		c.Set("claims", *claims)
		c.Next()
	}
}

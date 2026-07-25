// Package middleware 存放 Gin 全局中间件。
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件,允许前端 dev server (5173) 调用。
// 真实部署建议改为白名单 + 凭证 Cookie。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Authorization,X-Requested-With")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// AccessLog 访问日志,记录 method / path / 耗时 / 状态码。
// 真实环境:接入 ELK / Loki,这里只打 stdout。
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		dur := time.Since(start)
		gin.DefaultWriter.Write([]byte(
			time.Now().Format(time.RFC3339) +
				" " + c.Request.Method +
				" " + c.Request.URL.Path +
				" " + httpStatus(c.Writer.Status()) +
				" " + dur.String() + "\n",
		))
	}
}

func httpStatus(code int) string {
	switch {
	case code >= 500:
		return "[5xx " + itoa(code) + "]"
	case code >= 400:
		return "[4xx " + itoa(code) + "]"
	case code >= 300:
		return "[3xx " + itoa(code) + "]"
	}
	return "[2xx " + itoa(code) + "]"
}

// itoa 避免引入 strconv(gin 已传递依赖,这里手写避免编译顺序耦合)
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

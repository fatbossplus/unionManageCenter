package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 结构化访问日志，记录：时间 / 请求ID / 方法 / 路径 / 状态码 / 耗时 / 客户端IP
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		status := c.Writer.Status()
		// 根据状态码着色（仅在开发模式有效）
		color := colorByStatus(status)
		reset := "\033[0m"

		fmt.Printf("[GW] %s | %s%3d%s | %12v | %-15s | %-7s %s | requestId=%s\n",
			start.Format("2006-01-02 15:04:05"),
			color, status, reset,
			latency,
			c.ClientIP(),
			c.Request.Method,
			c.Request.URL.Path,
			GetRequestID(c),
		)
	}
}

func colorByStatus(code int) string {
	switch {
	case code >= 500:
		return "\033[31m" // 红
	case code >= 400:
		return "\033[33m" // 黄
	case code >= 300:
		return "\033[36m" // 青
	default:
		return "\033[32m" // 绿
	}
}

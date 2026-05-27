package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout 请求超时控制：超过 d 时间未完成则返回 503
func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()

		// 把带超时的 ctx 注入 request
		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{}, 1)
		go func() {
			c.Next()
			done <- struct{}{}
		}()

		select {
		case <-done:
			// 正常完成
		case <-ctx.Done():
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "请求处理超时，请稍后重试",
			})
		}
	}
}

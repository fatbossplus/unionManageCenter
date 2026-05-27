package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery 最外层：捕获 panic，返回统一 JSON，不暴露堆栈给客户端
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 打印完整堆栈到服务端日志
				fmt.Printf("[PANIC] requestId=%s %s %s\n%s\n",
					GetRequestID(c), c.Request.Method, c.Request.URL.Path,
					debug.Stack(),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "服务器内部错误，请稍后重试",
				})
			}
		}()
		c.Next()
	}
}

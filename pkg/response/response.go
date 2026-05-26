package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// OK 返回成功响应
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Result{Code: 0, Message: "success", Data: data})
}

// OKMsg 返回带自定义消息的成功响应
func OKMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Result{Code: 0, Message: msg})
}

// Fail 返回失败响应
func Fail(c *gin.Context, code int, msg string) {
	c.AbortWithStatusJSON(http.StatusOK, Result{Code: code, Message: msg})
}

// Unauthorized 返回 401 未授权响应
func Unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, Result{Code: 401, Message: "未授权，请先登录"})
}

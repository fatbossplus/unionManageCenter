package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const headerRequestID = "X-Request-ID"
const ctxKeyRequestID = "requestID"

// RequestID 为每个请求注入唯一 ID，便于链路追踪
// 优先使用客户端传入的 X-Request-ID，否则自动生成
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(headerRequestID)
		if rid == "" {
			rid = newRequestID()
		}
		c.Set(ctxKeyRequestID, rid)
		c.Header(headerRequestID, rid) // 透传给响应头
		c.Next()
	}
}

// GetRequestID 从 context 取出请求 ID
func GetRequestID(c *gin.Context) string {
	v, _ := c.Get(ctxKeyRequestID)
	s, _ := v.(string)
	return s
}

func newRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

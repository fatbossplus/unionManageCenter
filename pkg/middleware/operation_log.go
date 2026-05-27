package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OperationLogEntry 与 operation_logs 表对应
type OperationLogEntry struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement"`
	UserID     uint64    `gorm:"column:user_id"`
	Username   string    `gorm:"column:username;size:64"`
	Action     string    `gorm:"column:action;size:64"`
	Resource   string    `gorm:"column:resource;size:64"`
	ResourceID uint64    `gorm:"column:resource_id"`
	Detail     string    `gorm:"column:detail;type:text"`
	IP         string    `gorm:"column:ip;size:64"`
	UserAgent  string    `gorm:"column:user_agent;size:256"`
	Status     int8      `gorm:"column:status;default:1"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (OperationLogEntry) TableName() string { return "operation_logs" }

// responseWriter 包装 gin.ResponseWriter 捕获响应体
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// OperationLog 最内层（紧贴业务 handler）：
// 自动将 POST / PUT / DELETE 请求记录到 operation_logs 表
func OperationLog(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		// 只记录写操作
		if method != "POST" && method != "PUT" && method != "DELETE" {
			c.Next()
			return
		}

		// 读取请求体（读完后要回写，否则 handler 拿不到）
		var reqBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			reqBody = string(bodyBytes)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 包装 ResponseWriter 以捕获响应
		rw := &responseWriter{c.Writer, &bytes.Buffer{}}
		c.Writer = rw

		c.Next()

		// Handler 执行完毕，异步写日志（不阻塞响应）
		userID := GetUserID(c)
		username := GetUsername(c)
		status := int8(1)
		if c.Writer.Status() >= 400 {
			status = 0
		}

		entry := OperationLogEntry{
			UserID:    userID,
			Username:  username,
			Action:    methodToAction(method),
			Resource:  pathToResource(c.FullPath()),
			IP:        c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			Status:    status,
			Detail:    truncate(reqBody, 2000),
		}

		go func() {
			db.Create(&entry)
		}()
	}
}

func methodToAction(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return strings.ToLower(method)
	}
}

func pathToResource(path string) string {
	// /api/v1/users/:id  →  users
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return path
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

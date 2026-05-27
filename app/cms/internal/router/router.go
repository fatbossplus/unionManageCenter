// Package router CMS 服务路由（洋葱模型，与 gateway 保持一致）
package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"unionManageCenter/cms/internal/handler"
	"unionManageCenter/pkg/middleware"
)

func New() *gin.Engine {
	r := gin.New()

	// 洋葱中间件（顺序与 gateway 保持一致）
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.RateLimit(300, time.Minute))
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Cors())

	v1 := r.Group("/api/v1/cms")
	v1.Use(middleware.JWTAuth())

	// ── 平台账号管理 ──────────────────────────────────────────
	accH := handler.NewAccountHandler()
	acc := v1.Group("/accounts")
	{
		acc.GET("", accH.List)
		acc.POST("", accH.Create)
		acc.PUT("/:id", accH.Update)
		acc.DELETE("/:id", accH.Delete)
		acc.GET("/:id/audit-logs", accH.AuditLogs)
	}

	// ── 驱动配置管理 ──────────────────────────────────────────
	drvH := handler.NewDriverHandler()
	drv := v1.Group("/drivers")
	{
		drv.GET("", drvH.ListAll)
		drv.PUT("", drvH.Update)
		drv.GET("/meta", drvH.DriverMeta)
	}

	// ── 采集任务管理 ──────────────────────────────────────────
	taskH := handler.NewTaskHandler()
	task := v1.Group("/tasks")
	{
		task.GET("", taskH.List)
		task.POST("", taskH.Create)
		task.PUT("/:id", taskH.Update)
		task.DELETE("/:id", taskH.Delete)
		task.POST("/:id/run", taskH.Run)
	}

	// ── 内容管理 ──────────────────────────────────────────────
	contentH := handler.NewContentHandler()
	content := v1.Group("/contents")
	{
		content.GET("", contentH.List)
		content.GET("/:id", contentH.Get)
		content.POST("/:id/process", contentH.Process)
		content.POST("/:id/skip", contentH.Skip)
		content.GET("/stats", contentH.Stats)
	}

	// ── 发布任务管理 ──────────────────────────────────────────
	pubH := handler.NewPublishHandler()
	pub := v1.Group("/publishes")
	{
		pub.GET("", pubH.List)
		pub.GET("/:id", pubH.Get)
		pub.PUT("/:id", pubH.Update)
		pub.POST("/:id/approve", pubH.Approve)
		pub.POST("/:id/reject", pubH.Reject)
		pub.POST("/:id/publish", pubH.Publish)
		pub.DELETE("/:id", pubH.Delete)
		pub.GET("/stats", pubH.Stats)
	}

	return r
}

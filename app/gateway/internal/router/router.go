package router

// 洋葱模型中间件调用顺序（请求由外向内，响应由内向外）：
//
//  ┌──────── Recovery（panic 捕获）─────────────────────────────┐
//  │ ┌────── RequestID（注入唯一追踪ID）──────────────────────┐  │
//  │ │ ┌──── Logger（访问日志 + 耗时）─────────────────────┐  │  │
//  │ │ │ ┌── RateLimit（IP 限流 600次/分钟）─────────────┐  │  │  │
//  │ │ │ │ ┌─ Timeout（30s 超时保护）──────────────────┐  │  │  │  │
//  │ │ │ │ │ ┌ CORS（跨域）──────────────────────────┐  │  │  │  │  │
//  │ │ │ │ │ │ ┌ JWTAuth（身份验证）───────────────┐  │  │  │  │  │  │
//  │ │ │ │ │ │ │ ┌ OperationLog（写操作入库）────┐  │  │  │  │  │  │  │
//  │ │ │ │ │ │ │ │        [ Handler ]           │  │  │  │  │  │  │  │
//  │ │ │ │ │ │ │ └──────────────────────────────┘  │  │  │  │  │  │  │
//  │ │ │ │ │ │ └──────────────────────────────────┘  │  │  │  │  │  │
//  │ │ │ │ │ └────────────────────────────────────────┘  │  │  │  │  │
//  │ │ │ │ └──────────────────────────────────────────────┘  │  │  │  │
//  │ │ │ └────────────────────────────────────────────────────┘  │  │  │
//  │ │ └──────────────────────────────────────────────────────────┘  │  │
//  │ └────────────────────────────────────────────────────────────────┘  │
//  └──────────────────────────────────────────────────────────────────────┘

import (
	"time"

	"github.com/gin-gonic/gin"
	"unionManageCenter/gateway/internal/handler"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/middleware"
)

func New() *gin.Engine {
	// 关闭 gin 默认 logger，由我们的 Logger 中间件统一输出
	r := gin.New()

	// ── 第1层（最外）：panic 捕获 ──────────────────────────────
	r.Use(middleware.Recovery())

	// ── 第2层：请求唯一 ID ────────────────────────────────────
	r.Use(middleware.RequestID())

	// ── 第3层：访问日志 ───────────────────────────────────────
	r.Use(middleware.Logger())

	// ── 第4层：IP 限流（600次/分钟）─────────────────────────
	r.Use(middleware.RateLimit(600, time.Minute))

	// ── 第5层：请求超时（30s）────────────────────────────────
	r.Use(middleware.Timeout(30 * time.Second))

	// ── 第6层：CORS 跨域 ──────────────────────────────────────
	r.Use(middleware.Cors())

	v1 := r.Group("/api/v1")

	// 公开路由（不需要 JWT）
	authH := handler.NewAuthHandler()
	v1.POST("/auth/login", authH.Login)

	// ── 第7层：JWT 身份验证 ───────────────────────────────────
	secured := v1.Group("", middleware.JWTAuth())

	// ── 第8层（最内）：写操作自动入库 operation_logs ──────────
	db := database.Get()
	secured.Use(middleware.OperationLog(db))

	// 权限校验快捷函数
	perm := func(code string) gin.HandlerFunc {
		return middleware.RequirePermission(db, code)
	}

	{
		// 当前管理员权限码列表（前端初始化用）
		secured.GET("/auth/permissions", middleware.GetPermissions(db))
		secured.GET("/auth/menus", authH.Menus)
		secured.POST("/auth/logout", authH.Logout)

		// 管理员账号管理（仅 superadmin）
		adm := handler.NewAdminHandler()
		secured.GET("/admins", perm("admin:list"), adm.List)
		secured.GET("/admins/me", adm.Me)
		secured.GET("/admins/:id", perm("admin:list"), adm.Get)
		secured.POST("/admins", perm("admin:create"), adm.Create)
		secured.PUT("/admins/:id", perm("admin:update"), adm.Update)
		secured.DELETE("/admins/:id", perm("admin:delete"), adm.Delete)
		secured.POST("/admins/:id/reset-password", perm("admin:update"), adm.ResetPassword)
		secured.GET("/users/me", adm.Me)

		// 平台用户管理
		u := handler.NewUserHandler()
		secured.GET("/users", u.List)
		secured.GET("/users/:id", u.Get)
		secured.POST("/users", perm("user:create"), u.Create)
		secured.PUT("/users/:id", perm("user:update"), u.Update)
		secured.DELETE("/users/:id", perm("user:delete"), u.Delete)
		secured.POST("/users/batch-enable", perm("user:update"), u.BatchEnable)
		secured.POST("/users/batch-disable", perm("user:update"), u.BatchDisable)

		// 角色 & 权限（仅 superadmin 可改）
		ro := handler.NewRoleHandler()
		secured.GET("/roles", ro.List)
		secured.GET("/roles/:id", ro.Get)
		secured.POST("/roles", perm("permission"), ro.Create)
		secured.PUT("/roles/:id", perm("permission"), ro.Update)
		secured.DELETE("/roles/:id", perm("permission"), ro.Delete)
		secured.PUT("/roles/:id/permissions", perm("permission"), ro.AssignPermissions)
		// 前端用 POST，同时兼容
		secured.POST("/roles/:id/permissions", perm("permission"), ro.AssignPermissions)

		pm := handler.NewPermissionHandler()
		secured.GET("/permissions", pm.Tree)
		secured.POST("/permissions", perm("permission"), pm.Create)
		secured.PUT("/permissions/:id", perm("permission"), pm.Update)
		secured.DELETE("/permissions/:id", perm("permission"), pm.Delete)

		// 联盟
		og := handler.NewOrgHandler()
		secured.GET("/orgs", og.List)
		secured.GET("/orgs/:id", og.Get)
		secured.POST("/orgs", perm("org:create"), og.Create)
		// 写操作：全局权限 OR 本联盟权限班子成员
		orgPerm := func(code string) gin.HandlerFunc {
			return middleware.RequireOrgPermission(db, code, "id")
		}
		secured.PUT("/orgs/:id", orgPerm("org:update"), og.Update)
		secured.DELETE("/orgs/:id", orgPerm("org:delete"), og.Delete)
		secured.GET("/orgs/:id/members", og.Members)
		secured.POST("/orgs/:id/members", orgPerm("org:update"), og.AddMember)
		secured.DELETE("/orgs/:id/members/:uid", orgPerm("org:update"), og.RemoveMember)
		// 权限班子管理（仅全局 org:update 或 superadmin）
		secured.GET("/orgs/:id/team", og.ListOrgAdmins)
		secured.POST("/orgs/:id/team", perm("org:update"), og.AddOrgAdmin)
		secured.PUT("/orgs/:id/team/:admin_id", perm("org:update"), og.UpdateOrgAdmin)
		secured.DELETE("/orgs/:id/team/:admin_id", perm("org:update"), og.RemoveOrgAdmin)

		// 订单
		od := handler.NewOrderHandler()
		secured.GET("/orders", od.List)
		secured.GET("/orders/:id", od.Get)
		secured.POST("/orders/:id/refund", perm("order:refund"), od.Refund)

		// 财务
		fi := handler.NewFinanceHandler()
		secured.GET("/finance", fi.List)
		secured.GET("/finance/:id", fi.Get)
		secured.POST("/finance/:id/settle", perm("finance:settle"), fi.Settle)
		secured.GET("/finance/accounts", fi.ListAccounts)
		secured.POST("/finance/accounts", perm("finance:settle"), fi.CreateAccount)

		// 消息
		msg := handler.NewMessageHandler()
		secured.GET("/messages", msg.List)
		secured.POST("/messages/read", msg.MarkRead)
		secured.POST("/messages/read-all", msg.MarkAllRead)

		// 仪表盘 & 报表
		dash := handler.NewDashboardHandler()
		secured.GET("/dashboard/stats", dash.Stats)
		secured.GET("/dashboard/trend", dash.Trend)
		secured.GET("/dashboard/org-types", dash.OrgTypes)
		secured.GET("/dashboard/org-rank", dash.OrgRank)
		secured.GET("/dashboard/events", dash.Events)

		rpt := handler.NewReportHandler()
		secured.GET("/reports/summary", rpt.Summary)
		secured.GET("/reports/daily", rpt.Daily)
		secured.GET("/reports/roles", rpt.RoleStats)
	}

	return r
}

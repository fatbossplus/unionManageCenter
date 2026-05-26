package router

import (
	"github.com/gin-gonic/gin"
	"unionManageCenter/gateway/internal/handler"
	"unionManageCenter/pkg/middleware"
)

func New() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Cors())

	v1 := r.Group("/api/v1")

	auth := handler.NewAuthHandler()
	v1.POST("/auth/login", auth.Login)

	secured := v1.Group("", middleware.JWTAuth())
	{
		u := handler.NewUserHandler()
		secured.GET("/users", u.List)
		secured.GET("/users/:id", u.Get)
		secured.POST("/users", u.Create)
		secured.PUT("/users/:id", u.Update)
		secured.DELETE("/users/:id", u.Delete)
		secured.POST("/users/batch-enable", u.BatchEnable)
		secured.POST("/users/batch-disable", u.BatchDisable)
		secured.GET("/users/me", u.Me)

		ro := handler.NewRoleHandler()
		secured.GET("/roles", ro.List)
		secured.POST("/roles", ro.Create)
		secured.PUT("/roles/:id", ro.Update)
		secured.DELETE("/roles/:id", ro.Delete)
		secured.PUT("/roles/:id/permissions", ro.AssignPermissions)

		pm := handler.NewPermissionHandler()
		secured.GET("/permissions", pm.Tree)
		secured.POST("/permissions", pm.Create)
		secured.PUT("/permissions/:id", pm.Update)
		secured.DELETE("/permissions/:id", pm.Delete)

		og := handler.NewOrgHandler()
		secured.GET("/orgs", og.List)
		secured.GET("/orgs/:id", og.Get)
		secured.POST("/orgs", og.Create)
		secured.PUT("/orgs/:id", og.Update)
		secured.DELETE("/orgs/:id", og.Delete)
		secured.GET("/orgs/:id/members", og.Members)
		secured.POST("/orgs/:id/members", og.AddMember)
		secured.DELETE("/orgs/:id/members/:uid", og.RemoveMember)

		od := handler.NewOrderHandler()
		secured.GET("/orders", od.List)
		secured.GET("/orders/:id", od.Get)
		secured.POST("/orders/:id/refund", od.Refund)

		fi := handler.NewFinanceHandler()
		secured.GET("/finance", fi.List)
		secured.GET("/finance/:id", fi.Get)
		secured.POST("/finance/:id/settle", fi.Settle)
		secured.GET("/finance/accounts", fi.ListAccounts)
		secured.POST("/finance/accounts", fi.CreateAccount)

		msg := handler.NewMessageHandler()
		secured.GET("/messages", msg.List)
		secured.POST("/messages/read", msg.MarkRead)
		secured.POST("/messages/read-all", msg.MarkAllRead)

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

		secured.GET("/auth/menus", auth.Menus)
		secured.POST("/auth/logout", auth.Logout)
	}

	return r
}

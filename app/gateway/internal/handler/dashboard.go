package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"unionManageCenter/gateway/internal/model"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/response"
)

type DashboardHandler struct{ db *gorm.DB }

func NewDashboardHandler() *DashboardHandler { return &DashboardHandler{db: database.Get()} }

func (h *DashboardHandler) Stats(c *gin.Context) {
	var totalUsers, activeOrgs, pendingOrders int64
	h.db.Model(&model.User{}).Where("deleted_at IS NULL").Count(&totalUsers)
	h.db.Model(&model.Org{}).Where("deleted_at IS NULL AND status = 1").Count(&activeOrgs)
	h.db.Model(&model.Order{}).Where("deleted_at IS NULL AND status = 1").Count(&pendingOrders)

	var monthlyRevenue struct{ Total float64 }
	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	h.db.Model(&model.Order{}).
		Where("deleted_at IS NULL AND status = 2 AND paid_at >= ?", firstDay).
		Select("COALESCE(SUM(amount),0) as total").Scan(&monthlyRevenue)

	var todayRevenue struct{ Total float64 }
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	h.db.Model(&model.Order{}).
		Where("deleted_at IS NULL AND status = 2 AND paid_at >= ?", today).
		Select("COALESCE(SUM(amount),0) as total").Scan(&todayRevenue)

	var todayNew int64
	h.db.Model(&model.User{}).Where("deleted_at IS NULL AND created_at >= ?", today).Count(&todayNew)

	// 在线用户：15 分钟内有登录行为的用户
	var onlineUsers int64
	h.db.Model(&model.User{}).
		Where("deleted_at IS NULL AND last_login_at >= ?", now.Add(-15*time.Minute)).
		Count(&onlineUsers)

	response.OK(c, gin.H{
		"total_users":     totalUsers,
		"active_orgs":     activeOrgs,
		"monthly_revenue": monthlyRevenue.Total,
		"pending_orders":  pendingOrders,
		"online_users":    onlineUsers,
		"today_revenue":   todayRevenue.Total,
		"today_new_users": todayNew,
	})
}

func (h *DashboardHandler) Trend(c *gin.Context) {
	period := c.DefaultQuery("period", "month")
	var days int
	switch period {
	case "quarter":
		days = 90
	case "year":
		days = 365
	default:
		days = 30
	}

	type TrendRow struct {
		Date    string  `json:"date"`
		Users   int64   `json:"users"`
		Revenue float64 `json:"revenue"`
	}

	var result []TrendRow
	h.db.Raw(`
		SELECT 
			DATE_FORMAT(d.date, '%m-%d') as date,
			COALESCE((SELECT COUNT(*) FROM users WHERE DATE(created_at) = d.date AND deleted_at IS NULL), 0) as users,
			COALESCE((SELECT SUM(amount) FROM orders WHERE DATE(paid_at) = d.date AND status = 2 AND deleted_at IS NULL), 0) as revenue
		FROM (
			SELECT DATE(DATE_SUB(NOW(), INTERVAL n DAY)) as date
			FROM (
				SELECT a.N + b.N * 10 + c.N * 100 as n
				FROM (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) a,
				     (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) b,
				     (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3) c
			) nums
			WHERE n < ?
			ORDER BY n DESC
		) d
	`, days).Scan(&result)

	response.OK(c, result)
}

func (h *DashboardHandler) OrgTypes(c *gin.Context) {
	type TypeRow struct {
		Name  string `json:"name"`
		Value int64  `json:"value"`
		Color string `json:"color"`
	}
	colorMap := map[string]string{
		"ec": "#1e40af", "service": "#3b82f6",
		"content": "#60a5fa", "other": "#93c5fd",
	}
	labelMap := map[string]string{
		"ec": "电商联盟", "service": "服务联盟",
		"content": "内容联盟", "other": "其他",
	}
	var rows []struct {
		Type  string
		Count int64
	}
	h.db.Model(&model.Org{}).Where("deleted_at IS NULL").
		Select("type, COUNT(*) as count").Group("type").Scan(&rows)

	result := make([]TypeRow, 0, len(rows))
	for _, r := range rows {
		label := labelMap[r.Type]
		if label == "" {
			label = r.Type
		}
		color := colorMap[r.Type]
		if color == "" {
			color = "#9ca3af"
		}
		result = append(result, TypeRow{Name: label, Value: r.Count, Color: color})
	}
	response.OK(c, result)
}

func (h *DashboardHandler) OrgRank(c *gin.Context) {
	type RankRow struct {
		Name    string  `json:"name"`
		Revenue float64 `json:"revenue"`
	}
	var result []RankRow
	h.db.Table("orgs o").
		Joins("LEFT JOIN orders od ON od.org_id = o.id AND od.status = 2 AND od.deleted_at IS NULL").
		Where("o.deleted_at IS NULL").
		Select("o.name, COALESCE(SUM(od.amount),0) as revenue").
		Group("o.id, o.name").
		Order("revenue DESC").
		Limit(5).Scan(&result)
	response.OK(c, result)
}

func (h *DashboardHandler) Events(c *gin.Context) {
	var msgs []model.Message
	h.db.Where("user_id = 0").Order("created_at DESC").Limit(8).Find(&msgs)

	type EventItem struct {
		ID    uint64 `json:"id"`
		Color string `json:"color"`
		Text  string `json:"text"`
		Ts    int64  `json:"ts"`
	}
	colorMap := map[string]string{
		"system": "#3b82f6", "order": "#f59e0b",
		"finance": "#10b981", "security": "#ef4444",
	}
	result := make([]EventItem, 0, len(msgs))
	for _, m := range msgs {
		color := colorMap[m.Type]
		if color == "" {
			color = "#6b7280"
		}
		result = append(result, EventItem{
			ID: m.ID, Color: color,
			Text: m.Content, Ts: m.CreatedAt.UnixMilli(),
		})
	}
	response.OK(c, result)
}

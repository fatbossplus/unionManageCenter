package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"unionManageCenter/gateway/internal/model"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/response"
)

type ReportHandler struct{ db *gorm.DB }

func NewReportHandler() *ReportHandler { return &ReportHandler{db: database.Get()} }

// Summary GET /reports/summary — 本月汇总 KPI
func (h *ReportHandler) Summary(c *gin.Context) {
	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var monthNewUsers int64
	h.db.Model(&model.User{}).
		Where("deleted_at IS NULL AND created_at >= ?", firstDay).
		Count(&monthNewUsers)

	var monthRevenue struct{ Total float64 }
	h.db.Model(&model.Order{}).
		Where("deleted_at IS NULL AND status = 2 AND paid_at >= ?", firstDay).
		Select("COALESCE(SUM(amount),0) as total").Scan(&monthRevenue)

	var activeOrgs int64
	h.db.Model(&model.Org{}).Where("deleted_at IS NULL AND status = 1").Count(&activeOrgs)

	var monthOrders int64
	h.db.Model(&model.Order{}).
		Where("deleted_at IS NULL AND created_at >= ?", firstDay).
		Count(&monthOrders)

	// 对比上月
	prevFirst := firstDay.AddDate(0, -1, 0)
	prevLast := firstDay.Add(-time.Second)

	var prevNewUsers int64
	h.db.Model(&model.User{}).
		Where("deleted_at IS NULL AND created_at BETWEEN ? AND ?", prevFirst, prevLast).
		Count(&prevNewUsers)

	var prevRevenue struct{ Total float64 }
	h.db.Model(&model.Order{}).
		Where("deleted_at IS NULL AND status = 2 AND paid_at BETWEEN ? AND ?", prevFirst, prevLast).
		Select("COALESCE(SUM(amount),0) as total").Scan(&prevRevenue)

	var prevOrders int64
	h.db.Model(&model.Order{}).
		Where("deleted_at IS NULL AND created_at BETWEEN ? AND ?", prevFirst, prevLast).
		Count(&prevOrders)

	pctChange := func(cur, prev float64) string {
		if prev == 0 {
			if cur > 0 {
				return "+100%"
			}
			return "0%"
		}
		v := (cur - prev) / prev * 100
		if v >= 0 {
			return "+" + formatPct(v) + "%"
		}
		return formatPct(v) + "%"
	}

	response.OK(c, gin.H{
		"month_new_users": monthNewUsers,
		"month_revenue":   monthRevenue.Total,
		"active_orgs":     activeOrgs,
		"month_orders":    monthOrders,
		"trends": gin.H{
			"users":   pctChange(float64(monthNewUsers), float64(prevNewUsers)),
			"revenue": pctChange(monthRevenue.Total, prevRevenue.Total),
			"orders":  pctChange(float64(monthOrders), float64(prevOrders)),
		},
	})
}

// Daily GET /reports/daily?days=30 — 近 N 日明细
func (h *ReportHandler) Daily(c *gin.Context) {
	days := 30
	switch c.Query("days") {
	case "7":
		days = 7
	case "90":
		days = 90
	}

	type DailyRow struct {
		Date        string  `json:"date"`
		NewUsers    int64   `json:"new_users"`
		ActiveUsers int64   `json:"active_users"`
		Revenue     float64 `json:"revenue"`
		Orders      int64   `json:"orders"`
	}

	var rows []DailyRow
	h.db.Raw(`
		SELECT
			d.date,
			COALESCE((SELECT COUNT(*) FROM users WHERE DATE(created_at)=d.date AND deleted_at IS NULL),0)      AS new_users,
			COALESCE((SELECT COUNT(*) FROM users WHERE DATE(last_login_at)=d.date AND deleted_at IS NULL),0)   AS active_users,
			COALESCE((SELECT SUM(amount) FROM orders WHERE DATE(paid_at)=d.date AND status=2 AND deleted_at IS NULL),0) AS revenue,
			COALESCE((SELECT COUNT(*) FROM orders WHERE DATE(created_at)=d.date AND deleted_at IS NULL),0)     AS orders
		FROM (
			SELECT DATE(DATE_SUB(NOW(), INTERVAL n DAY)) AS date
			FROM (
				SELECT a.N + b.N*10 AS n
				FROM (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4
				      UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) a,
				     (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4
				      UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) b
			) nums WHERE n < ?
			ORDER BY n DESC
		) d
	`, days).Scan(&rows)

	// 补充环比列
	type OutputRow struct {
		DailyRow
		Trend string `json:"trend"`
	}
	out := make([]OutputRow, len(rows))
	for i, r := range rows {
		trend := "—"
		if i+1 < len(rows) {
			prev := rows[i+1].Revenue
			if prev > 0 {
				v := (r.Revenue - prev) / prev * 100
				if v >= 0 {
					trend = "↑ " + formatPct(v) + "%"
				} else {
					trend = "↓ " + formatPct(-v) + "%"
				}
			}
		}
		out[i] = OutputRow{DailyRow: r, Trend: trend}
	}
	response.OK(c, out)
}

// RoleStats GET /reports/roles — 角色分布统计（管理员维度）
func (h *ReportHandler) RoleStats(c *gin.Context) {
	type RoleRow struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Code        string `json:"code"`
		Description string `json:"description"`
		UserCount   int64  `json:"user_count"`
		PermCount   int64  `json:"perm_count"`
		Status      int8   `json:"status"`
	}
	var rows []RoleRow
	// user_count 改为统计 admins 表（原 user_roles 表已废弃）
	h.db.Table("roles r").
		Select("r.id, r.name, r.code, r.description, r.status, "+
			"(SELECT COUNT(*) FROM admins a WHERE a.role_id=r.id AND a.deleted_at IS NULL) AS user_count, "+
			"(SELECT COUNT(*) FROM role_permissions rp WHERE rp.role_id=r.id) AS perm_count").
		Where("r.deleted_at IS NULL").
		Order("r.sort ASC").
		Scan(&rows)

	var totalAdmins int64
	h.db.Table("admins").Where("deleted_at IS NULL").Count(&totalAdmins)

	var totalPerms int64
	h.db.Model(&model.Permission{}).Where("deleted_at IS NULL").Count(&totalPerms)

	response.OK(c, gin.H{
		"roles":       rows,
		"total_users": totalAdmins,
		"total_perms": totalPerms,
	})
}

func formatPct(v float64) string {
	if v == float64(int(v)) {
		return itoa(int(v))
	}
	return ftoa(v)
}

func itoa(n int) string {
	if n < 0 {
		return "-" + itoa(-n)
	}
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func ftoa(f float64) string {
	// 1位小数格式，负数整体取反处理
	if f < 0 {
		return "-" + ftoa(-f)
	}
	i := int(f * 10)
	frac := i % 10
	if frac < 0 {
		frac = -frac
	}
	return itoa(i/10) + "." + itoa(frac)
}

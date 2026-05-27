package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"unionManageCenter/pkg/response"
)

// rolePermCache 按 role_id 缓存权限 code 集合，避免每次请求查库
type rolePermCache struct {
	mu      sync.RWMutex
	data    map[uint64]map[string]struct{}
	loadAt  map[uint64]time.Time
	ttl     time.Duration
}

var permCache = &rolePermCache{
	data:   make(map[uint64]map[string]struct{}),
	loadAt: make(map[uint64]time.Time),
	ttl:    5 * time.Minute,
}

func (c *rolePermCache) load(db *gorm.DB, roleID uint64) map[string]struct{} {
	c.mu.RLock()
	if codes, ok := c.data[roleID]; ok {
		if time.Since(c.loadAt[roleID]) < c.ttl {
			c.mu.RUnlock()
			return codes
		}
	}
	c.mu.RUnlock()

	// 查库
	var codes []string
	db.Table("permissions p").
		Select("p.code").
		Joins("JOIN role_permissions rp ON rp.permission_id = p.id").
		Where("rp.role_id = ? AND p.status = 1", roleID).
		Pluck("p.code", &codes)

	set := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		set[code] = struct{}{}
	}

	c.mu.Lock()
	c.data[roleID] = set
	c.loadAt[roleID] = time.Now()
	c.mu.Unlock()
	return set
}

// InvalidateRole 角色权限变更时调用，清除对应缓存
func InvalidateRole(roleID uint64) {
	permCache.mu.Lock()
	delete(permCache.data, roleID)
	delete(permCache.loadAt, roleID)
	permCache.mu.Unlock()
}

// RequirePermission 检查当前管理员是否拥有指定权限码
// superadmin 跳过所有检查；其他角色查 role_permissions 表
func RequirePermission(db *gorm.DB, permCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleCode := GetRoleCode(c)
		if roleCode == "superadmin" {
			c.Next()
			return
		}

		// 从 admins 表拿 role_id
		adminID := GetUserID(c)
		var roleID uint64
		if err := db.Table("admins").Select("role_id").
			Where("id = ? AND deleted_at IS NULL", adminID).
			Scan(&roleID).Error; err != nil || roleID == 0 {
			response.Forbidden(c)
			c.Abort()
			return
		}

		codes := permCache.load(db, roleID)
		if _, ok := codes[permCode]; !ok {
			response.Forbidden(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetPermissions 返回当前管理员所有权限码（用于前端初始化）
func GetPermissions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleCode := GetRoleCode(c)

		// superadmin 返回所有权限码
		if roleCode == "superadmin" {
			var codes []string
			db.Table("permissions").Where("status = 1").Pluck("code", &codes)
			c.JSON(200, gin.H{"code": 0, "message": "success", "data": codes})
			return
		}

		adminID := GetUserID(c)
		var roleID uint64
		db.Table("admins").Select("role_id").
			Where("id = ? AND deleted_at IS NULL", adminID).Scan(&roleID)

		codes := permCache.load(db, roleID)
		list := make([]string, 0, len(codes))
		for code := range codes {
			list = append(list, code)
		}
		c.JSON(200, gin.H{"code": 0, "message": "success", "data": list})
	}
}

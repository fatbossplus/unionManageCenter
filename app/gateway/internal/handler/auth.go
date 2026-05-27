package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"unionManageCenter/gateway/internal/model"
	"unionManageCenter/pkg/auth"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/middleware"
	"unionManageCenter/pkg/response"
)

type AuthHandler struct{ db *gorm.DB }

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{db: database.Get()}
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 后台管理员登录，查 admins 表
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	var admin model.Admin
	if err := h.db.Where("username = ? AND deleted_at IS NULL", req.Username).First(&admin).Error; err != nil {
		response.Fail(c, 400, "账号或密码错误")
		return
	}
	if admin.Status == 0 {
		response.Fail(c, 403, "账号已禁用")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		response.Fail(c, 400, "账号或密码错误")
		return
	}

	// 获取角色编码
	var role model.Role
	roleCode := "operator"
	if h.db.First(&role, admin.RoleID).Error == nil {
		roleCode = role.Code
	}

	now := time.Now()
	h.db.Model(&admin).Updates(map[string]any{
		"last_login_at": now,
		"last_login_ip": c.ClientIP(),
	})

	email := ""
	if admin.Email != nil {
		email = *admin.Email
	}
	token, _ := auth.GenerateToken(admin.ID, admin.Username, roleCode)
	response.OK(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":        admin.ID,
			"username":  admin.Username,
			"email":     email,
			"real_name": admin.RealName,
			"avatar":    admin.Avatar,
			"role":      roleCode,
		},
	})
}

// Menus 根据管理员角色返回菜单树
func (h *AuthHandler) Menus(c *gin.Context) {
	adminID := middleware.GetUserID(c)
	roleCode := middleware.GetRoleCode(c)

	var perms []model.Permission
	if roleCode == "superadmin" {
		h.db.Where("type = 1 AND status = 1 AND deleted_at IS NULL").
			Order("sort ASC").Find(&perms)
	} else {
		var admin model.Admin
		if h.db.First(&admin, adminID).Error != nil {
			response.Fail(c, 401, "管理员不存在")
			return
		}
		h.db.Table("permissions p").
			Joins("JOIN role_permissions rp ON rp.permission_id = p.id").
			Where("rp.role_id = ? AND p.type = 1 AND p.status = 1 AND p.deleted_at IS NULL", admin.RoleID).
			Order("p.sort ASC").Find(&perms)
	}

	response.OK(c, buildTree(perms, 0))
}

func buildTree(perms []model.Permission, parentID uint64) []model.Permission {
	var result []model.Permission
	for _, p := range perms {
		if p.ParentID == parentID {
			p.Children = buildTree(perms, p.ID)
			result = append(result, p)
		}
	}
	return result
}

func (h *AuthHandler) Logout(c *gin.Context) {
	response.OKMsg(c, "已退出登录")
}

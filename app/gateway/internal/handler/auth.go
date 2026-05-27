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

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	var user model.User
	if err := h.db.Where("username = ? AND deleted_at IS NULL", req.Username).First(&user).Error; err != nil {
		response.Fail(c, 400, "账号或密码错误")
		return
	}
	if user.Status == 3 {
		response.Fail(c, 403, "账号已禁用")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.Fail(c, 400, "账号或密码错误")
		return
	}

	var userRoles []model.UserRole
	h.db.Where("user_id = ?", user.ID).Find(&userRoles)
	roleCode := "member"
	if len(userRoles) > 0 {
		var role model.Role
		if h.db.First(&role, userRoles[0].RoleID).Error == nil {
			roleCode = role.Code
		}
	}

	now := time.Now()
	h.db.Model(&user).Updates(map[string]any{
		"last_login_at": now,
		"last_login_ip": c.ClientIP(),
	})

	token, _ := auth.GenerateToken(user.ID, user.Username, roleCode)
	response.OK(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"email":     func() string { if user.Email != nil { return *user.Email }; return "" }(),
			"real_name": user.RealName,
			"avatar":    user.Avatar,
			"role":      roleCode,
		},
	})
}

func (h *AuthHandler) Menus(c *gin.Context) {
	userID := middleware.GetUserID(c)
	roleCode := middleware.GetRoleCode(c)

	var perms []model.Permission
	if roleCode == "superadmin" {
		h.db.Where("type = 1 AND status = 1 AND deleted_at IS NULL").
			Order("sort ASC").Find(&perms)
	} else {
		h.db.Table("permissions p").
			Joins("JOIN role_permissions rp ON rp.permission_id = p.id").
			Joins("JOIN user_roles ur ON ur.role_id = rp.role_id").
			Where("ur.user_id = ? AND p.type = 1 AND p.status = 1 AND p.deleted_at IS NULL", userID).
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

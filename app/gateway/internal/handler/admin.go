package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"unionManageCenter/gateway/internal/model"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/middleware"
	"unionManageCenter/pkg/response"
)

type AdminHandler struct{ db *gorm.DB }

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{db: database.Get()}
}

// List GET /admins
func (h *AdminHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 { page = 1 }
	if size < 1 || size > 100 { size = 20 }

	db := h.db.Model(&model.Admin{}).Where("deleted_at IS NULL")
	if kw := c.Query("keyword"); kw != "" {
		db = db.Where("username LIKE ? OR real_name LIKE ?", "%"+kw+"%", "%"+kw+"%")
	}
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}

	var total int64
	db.Count(&total)

	var admins []model.Admin
	db.Order("id ASC").Offset((page - 1) * size).Limit(size).Find(&admins)

	// 查角色信息
	var roles []model.Role
	h.db.Find(&roles)
	roleMap := map[uint64]model.Role{}
	for _, r := range roles { roleMap[r.ID] = r }

	type AdminVO struct {
		model.Admin
		RoleName string `json:"role_name"`
		RoleCode string `json:"role_code"`
	}
	list := make([]AdminVO, 0, len(admins))
	for _, a := range admins {
		vo := AdminVO{Admin: a}
		if r, ok := roleMap[a.RoleID]; ok {
			vo.RoleName = r.Name
			vo.RoleCode = r.Code
		}
		list = append(list, vo)
	}
	response.OK(c, gin.H{"list": list, "total": total, "page": page, "page_size": size})
}

// Get GET /admins/:id
func (h *AdminHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var admin model.Admin
	if h.db.Where("id = ? AND deleted_at IS NULL", id).First(&admin).Error != nil {
		response.Fail(c, 404, "管理员不存在")
		return
	}
	response.OK(c, admin)
}

// Me GET /admins/me
func (h *AdminHandler) Me(c *gin.Context) {
	adminID := middleware.GetUserID(c)
	var admin model.Admin
	if h.db.Where("id = ? AND deleted_at IS NULL", adminID).First(&admin).Error != nil {
		response.Fail(c, 404, "管理员不存在")
		return
	}
	var role model.Role
	roleCode := ""
	if h.db.First(&role, admin.RoleID).Error == nil {
		roleCode = role.Code
	}
	email := ""
	if admin.Email != nil { email = *admin.Email }
	response.OK(c, gin.H{
		"id":        admin.ID,
		"username":  admin.Username,
		"email":     email,
		"phone":     admin.Phone,
		"real_name": admin.RealName,
		"avatar":    admin.Avatar,
		"role_id":   admin.RoleID,
		"role_code": roleCode,
		"status":    admin.Status,
	})
}

type createAdminReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email"    binding:"omitempty,email"`
	Phone    string `json:"phone"`
	RealName string `json:"real_name"`
	RoleID   uint64 `json:"role_id"`
	Status   int8   `json:"status"`
}

// Create POST /admins
func (h *AdminHandler) Create(c *gin.Context) {
	var req createAdminReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	admin := model.Admin{
		Username: req.Username,
		Password: string(hash),
		Phone:    req.Phone,
		RealName: req.RealName,
		RoleID:   req.RoleID,
		Status:   req.Status,
	}
	if req.Email != "" { admin.Email = &req.Email }
	if admin.RoleID == 0 { admin.RoleID = 4 } // 默认运营人员
	if admin.Status == 0 { admin.Status = 1 }
	if err := h.db.Create(&admin).Error; err != nil {
		response.Fail(c, 500, "创建失败: "+err.Error())
		return
	}
	response.OK(c, admin)
}

// Update PUT /admins/:id
func (h *AdminHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req map[string]any
	c.ShouldBindJSON(&req)
	// 密码单独处理
	if pwd, ok := req["password"].(string); ok && pwd != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err == nil {
			h.db.Model(&model.Admin{}).Where("id = ?", id).Update("password", string(hash))
		}
	}
	delete(req, "password")
	if len(req) > 0 {
		h.db.Model(&model.Admin{}).Where("id = ?", id).Updates(req)
	}
	response.OKMsg(c, "更新成功")
}

// Delete DELETE /admins/:id
func (h *AdminHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	adminID := middleware.GetUserID(c)
	if id == adminID {
		response.Fail(c, 400, "不能删除自己")
		return
	}
	h.db.Model(&model.Admin{}).Where("id = ?", id).Update("deleted_at", "NOW()")
	response.OKMsg(c, "删除成功")
}

// ResetPassword POST /admins/:id/reset-password
func (h *AdminHandler) ResetPassword(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	h.db.Model(&model.Admin{}).Where("id = ?", id).Update("password", string(hash))
	response.OKMsg(c, "密码已重置")
}

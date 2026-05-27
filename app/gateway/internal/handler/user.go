package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"unionManageCenter/gateway/internal/model"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/response"
)

type UserHandler struct{ db *gorm.DB }

func NewUserHandler() *UserHandler { return &UserHandler{db: database.Get()} }

// List GET /users
func (h *UserHandler) List(c *gin.Context) {
	var req struct {
		Page       int    `form:"page"`
		PageSize   int    `form:"page_size"`
		Keyword    string `form:"keyword"`
		OrgName    string `form:"org_name"`
		Role       string `form:"role"`
		Status     string `form:"status"`
		CertStatus string `form:"cert_status"`
		Source     string `form:"source"`
		StartDate  string `form:"start_date"`
		EndDate    string `form:"end_date"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	offset := (req.Page - 1) * req.PageSize

	db := h.db.Model(&model.User{}).Where("deleted_at IS NULL")
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		db = db.Where("username LIKE ? OR email LIKE ? OR phone LIKE ?", kw, kw, kw)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.CertStatus != "" {
		db = db.Where("cert_status = ?", req.CertStatus)
	}
	if req.Source != "" {
		db = db.Where("source = ?", req.Source)
	}
	if req.StartDate != "" {
		db = db.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		db = db.Where("created_at <= ?", req.EndDate+" 23:59:59")
	}

	var total int64
	db.Count(&total)

	var users []model.User
	db.Preload("Roles").Offset(offset).Limit(req.PageSize).
		Order("created_at DESC").Find(&users)

	response.OK(c, model.PageResult{List: users, Total: total, Page: req.Page, Size: req.PageSize})
}

// Get GET /users/:id
func (h *UserHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var user model.User
	if err := h.db.Preload("Roles").Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		response.Fail(c, 404, "用户不存在")
		return
	}
	response.OK(c, user)
}

// Me GET /users/me
func (h *UserHandler) Me(c *gin.Context) {
	userID, _ := c.Get("userID")
	var user model.User
	h.db.Preload("Roles").Where("id = ? AND deleted_at IS NULL", userID).First(&user)
	response.OK(c, user)
}

type createUserReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email"    binding:"omitempty,email"`
	Phone    string `json:"phone"`
	RealName string `json:"real_name"`
	Status   int8   `json:"status"`
	RoleID   uint64 `json:"role_id"`
}

// Create POST /users
func (h *UserHandler) Create(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := model.User{
		Username: req.Username,
		Password: string(hash),
		Phone:    req.Phone,
		RealName: req.RealName,
		Status:   req.Status,
	}
	if req.Email != "" {
		user.Email = &req.Email
	}
	if user.Status == 0 {
		user.Status = 1
	}
	if err := h.db.Create(&user).Error; err != nil {
		response.Fail(c, 500, "创建失败: "+err.Error())
		return
	}
	response.OK(c, user)
}

// Update PUT /users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req map[string]any
	c.ShouldBindJSON(&req)
	// 密码单独处理：需要重新 hash
	if pwd, ok := req["password"].(string); ok && pwd != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err == nil {
			h.db.Model(&model.User{}).Where("id = ?", id).Update("password", string(hash))
		}
	}
	delete(req, "password")
	if len(req) > 0 {
		h.db.Model(&model.User{}).Where("id = ?", id).Updates(req)
	}
	response.OKMsg(c, "更新成功")
}

// Delete DELETE /users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	h.db.Model(&model.User{}).Where("id = ?", id).Update("deleted_at", "NOW()")
	response.OKMsg(c, "删除成功")
}

// BatchEnable POST /users/batch-enable
func (h *UserHandler) BatchEnable(c *gin.Context) {
	var req struct {
		IDs []uint64 `json:"ids"`
	}
	c.ShouldBindJSON(&req)
	h.db.Model(&model.User{}).Where("id IN ?", req.IDs).Update("status", 1)
	response.OKMsg(c, "批量启用成功")
}

// BatchDisable POST /users/batch-disable
func (h *UserHandler) BatchDisable(c *gin.Context) {
	var req struct {
		IDs []uint64 `json:"ids"`
	}
	c.ShouldBindJSON(&req)
	h.db.Model(&model.User{}).Where("id IN ?", req.IDs).Update("status", 3)
	response.OKMsg(c, "批量禁用成功")
}

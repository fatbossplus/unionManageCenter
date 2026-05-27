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

	db := h.db.Model(&model.User{}).Where("u.deleted_at IS NULL")
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		db = db.Where("(u.username LIKE ? OR u.email LIKE ? OR u.phone LIKE ?)", kw, kw, kw)
	}
	if req.Status != "" {
		db = db.Where("u.status = ?", req.Status)
	}
	if req.CertStatus != "" {
		db = db.Where("u.cert_status = ?", req.CertStatus)
	}
	if req.Source != "" {
		db = db.Where("u.source = ?", req.Source)
	}
	if req.StartDate != "" {
		db = db.Where("u.created_at >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		db = db.Where("u.created_at <= ?", req.EndDate+" 23:59:59")
	}

	// 带联盟名称筛选时 JOIN orgs
	if req.OrgName != "" {
		db = db.Joins("JOIN org_members om_f ON om_f.user_id = u.id").
			Joins("JOIN orgs o_f ON o_f.id = om_f.org_id AND o_f.deleted_at IS NULL").
			Where("o_f.name LIKE ?", "%"+req.OrgName+"%")
	}

	var total int64
	db.Table("users u").Count(&total)

	// 补充每个用户的主联盟信息（取 created_at 最早的一条 org_member）
	type UserRow struct {
		model.User
		OrgRole string `json:"org_role"`
		OrgName string `json:"org_name"`
	}
	var users []UserRow
	h.db.Table("users u").
		Select("u.*, COALESCE(om.role,'') AS org_role, COALESCE(o.name,'') AS org_name").
		Joins("LEFT JOIN org_members om ON om.user_id = u.id AND om.status=1").
		Joins("LEFT JOIN orgs o ON o.id = om.org_id AND o.deleted_at IS NULL").
		Where("u.deleted_at IS NULL").
		Group("u.id, om.id, o.id").
		Offset(offset).Limit(req.PageSize).
		Order("u.created_at DESC").Scan(&users)

	response.OK(c, model.PageResult{List: users, Total: total, Page: req.Page, Size: req.PageSize})
}

// Get GET /users/:id
func (h *UserHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	type UserRow struct {
		model.User
		OrgRole string `json:"org_role"`
		OrgName string `json:"org_name"`
	}
	var user UserRow
	err := h.db.Table("users u").
		Select("u.*, COALESCE(om.role,'') AS org_role, COALESCE(o.name,'') AS org_name").
		Joins("LEFT JOIN org_members om ON om.user_id = u.id AND om.status=1").
		Joins("LEFT JOIN orgs o ON o.id = om.org_id AND o.deleted_at IS NULL").
		Where("u.id = ? AND u.deleted_at IS NULL", id).
		First(&user).Error
	if err != nil {
		response.Fail(c, 404, "用户不存在")
		return
	}
	response.OK(c, user)
}

type createUserReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email"    binding:"omitempty,email"`
	Phone    string `json:"phone"`
	RealName string `json:"real_name"`
	Status   int8   `json:"status"`
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
		Source:   "admin",
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
	// 密码单独处理
	if pwd, ok := req["password"].(string); ok && pwd != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err == nil {
			h.db.Model(&model.User{}).Where("id = ?", id).Update("password", string(hash))
		}
	}
	delete(req, "password")
	delete(req, "id")
	if len(req) > 0 {
		h.db.Model(&model.User{}).Where("id = ?", id).Updates(req)
	}
	response.OKMsg(c, "更新成功")
}

// Delete DELETE /users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	h.db.Model(&model.User{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()"))
	response.OKMsg(c, "删除成功")
}

// BatchEnable POST /users/batch-enable
func (h *UserHandler) BatchEnable(c *gin.Context) {
	var req struct {
		IDs []uint64 `json:"ids"`
	}
	c.ShouldBindJSON(&req)
	if len(req.IDs) > 0 {
		h.db.Model(&model.User{}).Where("id IN ?", req.IDs).Update("status", 1)
	}
	response.OKMsg(c, "批量启用成功")
}

// BatchDisable POST /users/batch-disable
func (h *UserHandler) BatchDisable(c *gin.Context) {
	var req struct {
		IDs []uint64 `json:"ids"`
	}
	c.ShouldBindJSON(&req)
	if len(req.IDs) > 0 {
		h.db.Model(&model.User{}).Where("id IN ?", req.IDs).Update("status", 3)
	}
	response.OKMsg(c, "批量禁用成功")
}

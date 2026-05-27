package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"unionManageCenter/gateway/internal/model"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/middleware"
	"unionManageCenter/pkg/response"
)

type RoleHandler struct{ db *gorm.DB }

func NewRoleHandler() *RoleHandler { return &RoleHandler{db: database.Get()} }

func (h *RoleHandler) List(c *gin.Context) {
	var roles []model.Role
	h.db.Where("deleted_at IS NULL").Order("sort ASC").Find(&roles)
	response.OK(c, roles)
}

// Get GET /roles/:id — 返回角色详情（含已分配的权限列表）
func (h *RoleHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var role model.Role
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&role).Error; err != nil {
		response.Fail(c, 404, "角色不存在")
		return
	}

	// 查该角色已分配的权限 ID 列表
	var permIDs []uint64
	h.db.Table("role_permissions").Select("permission_id").
		Where("role_id = ?", id).Pluck("permission_id", &permIDs)

	response.OK(c, gin.H{
		"role":        role,
		"permissions": permIDs,
	})
}

func (h *RoleHandler) Create(c *gin.Context) {
	var role model.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if err := h.db.Create(&role).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req map[string]any
	c.ShouldBindJSON(&req)
	delete(req, "id")
	delete(req, "code")
	h.db.Model(&model.Role{}).Where("id = ?", id).Updates(req)
	response.OKMsg(c, "更新成功")
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	// 超级管理员角色不允许删除
	if id == 1 {
		response.Fail(c, 400, "超级管理员角色不可删除")
		return
	}
	h.db.Model(&model.Role{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()"))
	response.OKMsg(c, "删除成功")
}

// AssignPermissions PUT /roles/:id/permissions 或 POST /roles/:id/permissions
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		PermissionIDs []uint64 `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	h.db.Where("role_id = ?", id).Delete(&model.RolePermission{})
	for _, pid := range req.PermissionIDs {
		h.db.Exec("INSERT IGNORE INTO role_permissions(role_id,permission_id) VALUES(?,?)", id, pid)
	}
	// 清除该角色的权限缓存
	middleware.InvalidateRole(uint64(id))
	response.OKMsg(c, "权限分配成功")
}

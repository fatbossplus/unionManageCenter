package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"unionManageCenter/gateway/internal/model"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/response"
)

type RoleHandler struct{ db *gorm.DB }

func NewRoleHandler() *RoleHandler { return &RoleHandler{db: database.Get()} }

func (h *RoleHandler) List(c *gin.Context) {
	var roles []model.Role
	h.db.Where("deleted_at IS NULL").Order("sort ASC").Find(&roles)
	response.OK(c, roles)
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
	h.db.Model(&model.Role{}).Where("id = ?", id).Update("deleted_at", "NOW()")
	response.OKMsg(c, "删除成功")
}

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
	response.OKMsg(c, "权限分配成功")
}

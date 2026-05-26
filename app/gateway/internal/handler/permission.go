package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"unionManageCenter/gateway/internal/model"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/response"
)

type PermissionHandler struct{ db *gorm.DB }

func NewPermissionHandler() *PermissionHandler { return &PermissionHandler{db: database.Get()} }

func (h *PermissionHandler) Tree(c *gin.Context) {
	var perms []model.Permission
	h.db.Where("deleted_at IS NULL").Order("sort ASC").Find(&perms)
	response.OK(c, buildTree(perms, 0))
}

func (h *PermissionHandler) Create(c *gin.Context) {
	var perm model.Permission
	if err := c.ShouldBindJSON(&perm); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	h.db.Create(&perm)
	response.OK(c, perm)
}

func (h *PermissionHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req map[string]any
	c.ShouldBindJSON(&req)
	h.db.Model(&model.Permission{}).Where("id = ?", id).Updates(req)
	response.OKMsg(c, "更新成功")
}

func (h *PermissionHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	h.db.Model(&model.Permission{}).Where("id = ?", id).Update("deleted_at", "NOW()")
	response.OKMsg(c, "删除成功")
}

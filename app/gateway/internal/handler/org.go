package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"unionManageCenter/gateway/internal/model"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/response"
)

type OrgHandler struct{ db *gorm.DB }

func NewOrgHandler() *OrgHandler { return &OrgHandler{db: database.Get()} }

func (h *OrgHandler) List(c *gin.Context) {
	var req struct {
		model.PageReq
		Keyword   string `form:"keyword"`
		Type      string `form:"type"`
		Status    string `form:"status"`
		Region    string `form:"region"`
		StartDate string `form:"start_date"`
		EndDate   string `form:"end_date"`
	}
	c.ShouldBindQuery(&req)
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	db := h.db.Model(&model.Org{}).Where("deleted_at IS NULL")
	if req.Keyword != "" {
		db = db.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	if req.Type != "" {
		db = db.Where("type = ?", req.Type)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Region != "" {
		db = db.Where("region LIKE ?", "%"+req.Region+"%")
	}
	if req.StartDate != "" {
		db = db.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		db = db.Where("created_at <= ?", req.EndDate+" 23:59:59")
	}

	var total int64
	db.Count(&total)
	var orgs []model.Org
	db.Preload("Leader").Offset(req.Offset()).Limit(req.PageSize).Order("created_at DESC").Find(&orgs)
	response.OK(c, model.PageResult{List: orgs, Total: total, Page: req.Page, Size: req.PageSize})
}

func (h *OrgHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var org model.Org
	if err := h.db.Preload("Leader").Where("id = ? AND deleted_at IS NULL", id).First(&org).Error; err != nil {
		response.Fail(c, 404, "联盟不存在")
		return
	}
	response.OK(c, org)
}

func (h *OrgHandler) Create(c *gin.Context) {
	var org model.Org
	if err := c.ShouldBindJSON(&org); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	org.Status = 2
	if err := h.db.Create(&org).Error; err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, org)
}

func (h *OrgHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req map[string]any
	c.ShouldBindJSON(&req)
	delete(req, "id")
	h.db.Model(&model.Org{}).Where("id = ?", id).Updates(req)
	response.OKMsg(c, "更新成功")
}

func (h *OrgHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	h.db.Model(&model.Org{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()"))
	response.OKMsg(c, "删除成功")
}

func (h *OrgHandler) Members(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var members []model.OrgMember
	h.db.Preload("User").Where("org_id = ? AND status = 1", id).Find(&members)
	response.OK(c, members)
}

func (h *OrgHandler) AddMember(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		UserID uint64 `json:"user_id" binding:"required"`
		Role   string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if req.Role == "" {
		req.Role = "member"
	}
	member := model.OrgMember{OrgID: id, UserID: req.UserID, Role: req.Role, Status: 1}
	if err := h.db.Create(&member).Error; err != nil {
		response.Fail(c, 500, "添加失败: "+err.Error())
		return
	}
	h.db.Model(&model.Org{}).Where("id = ?", id).UpdateColumn("member_count", gorm.Expr("member_count + 1"))
	response.OKMsg(c, "添加成功")
}

func (h *OrgHandler) RemoveMember(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	uid, _ := strconv.ParseUint(c.Param("uid"), 10, 64)
	h.db.Where("org_id = ? AND user_id = ?", id, uid).Delete(&model.OrgMember{})
	h.db.Model(&model.Org{}).Where("id = ?", id).UpdateColumn("member_count", gorm.Expr("GREATEST(member_count - 1, 0)"))
	response.OKMsg(c, "已移除")
}

// ── 权限班子（OrgAdmin）─────────────────────────────────

// ListOrgAdmins GET /orgs/:id/team — 查询联盟权限班子
func (h *OrgHandler) ListOrgAdmins(c *gin.Context) {
	orgID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var list []model.OrgAdmin
	h.db.Preload("Admin").Preload("Role").
		Where("org_id = ? AND deleted_at IS NULL", orgID).
		Order("created_at ASC").Find(&list)
	response.OK(c, list)
}

// AddOrgAdmin POST /orgs/:id/team — 添加管理员到权限班子
func (h *OrgHandler) AddOrgAdmin(c *gin.Context) {
	orgID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		AdminID uint64 `json:"admin_id" binding:"required"`
		RoleID  uint   `json:"role_id"  binding:"required"`
		Remark  string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}

	// 检查联盟是否存在
	var count int64
	h.db.Model(&model.Org{}).Where("id = ? AND deleted_at IS NULL", orgID).Count(&count)
	if count == 0 {
		response.Fail(c, 404, "联盟不存在")
		return
	}

	oa := model.OrgAdmin{OrgID: orgID, AdminID: req.AdminID, RoleID: req.RoleID, Status: 1, Remark: req.Remark}
	if err := h.db.Create(&oa).Error; err != nil {
		response.Fail(c, 500, "添加失败: "+err.Error())
		return
	}
	h.db.Preload("Admin").Preload("Role").First(&oa, oa.ID)
	response.OK(c, oa)
}

// UpdateOrgAdmin PUT /orgs/:id/team/:admin_id — 调整角色或备注
func (h *OrgHandler) UpdateOrgAdmin(c *gin.Context) {
	orgID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	adminID, _ := strconv.ParseUint(c.Param("admin_id"), 10, 64)
	var req struct {
		RoleID uint   `json:"role_id"`
		Status int8   `json:"status"`
		Remark string `json:"remark"`
	}
	c.ShouldBindJSON(&req)

	updates := map[string]any{}
	if req.RoleID > 0 {
		updates["role_id"] = req.RoleID
	}
	if req.Status == 0 || req.Status == 1 {
		updates["status"] = req.Status
	}
	updates["remark"] = req.Remark

	h.db.Model(&model.OrgAdmin{}).
		Where("org_id = ? AND admin_id = ? AND deleted_at IS NULL", orgID, adminID).
		Updates(updates)
	response.OKMsg(c, "更新成功")
}

// RemoveOrgAdmin DELETE /orgs/:id/team/:admin_id — 从权限班子移除
func (h *OrgHandler) RemoveOrgAdmin(c *gin.Context) {
	orgID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	adminID, _ := strconv.ParseUint(c.Param("admin_id"), 10, 64)
	h.db.Where("org_id = ? AND admin_id = ? AND deleted_at IS NULL", orgID, adminID).
		Delete(&model.OrgAdmin{})
	response.OKMsg(c, "已移除")
}

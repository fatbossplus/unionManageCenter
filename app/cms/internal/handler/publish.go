package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"unionManageCenter/cms/internal/model"
	"unionManageCenter/cms/internal/service"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/response"
)

// PublishHandler 发布任务管理
type PublishHandler struct {
	db         *gorm.DB
	publishSvc *service.PublisherService
}

func NewPublishHandler() *PublishHandler {
	return &PublishHandler{
		db:         database.Get(),
		publishSvc: service.NewPublisherService(),
	}
}

// List 发布任务列表
func (h *PublishHandler) List(c *gin.Context) {
	var req struct {
		model.PageReq
		Platform string `form:"platform"`
		Status   string `form:"status"`
		OrgID    uint64 `form:"org_id"`
	}
	c.ShouldBindQuery(&req)
	req.Normalize()

	q := h.db.Model(&model.PublishTask{}).Where("deleted_at IS NULL")
	if req.Platform != "" {
		q = q.Where("target_platform = ?", req.Platform)
	}
	if req.Status != "" {
		q = q.Where("status = ?", req.Status)
	}
	if req.OrgID > 0 {
		q = q.Where("org_id = ?", req.OrgID)
	}

	var total int64
	q.Count(&total)
	var tasks []model.PublishTask
	q.Order("id DESC").Limit(req.PageSize).Offset(req.Offset()).Find(&tasks)

	response.OK(c, gin.H{"list": tasks, "total": total, "page": req.Page, "page_size": req.PageSize})
}

// Get 发布任务详情
func (h *PublishHandler) Get(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var task model.PublishTask
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&task).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "发布任务不存在")
		return
	}
	response.OK(c, task)
}

// Update 更新发布内容（仅 draft/failed 状态可改）
func (h *PublishHandler) Update(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var task model.PublishTask
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&task).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "发布任务不存在")
		return
	}
	if task.Status != model.PublishStatusDraft && task.Status != model.PublishStatusFailed {
		response.Fail(c, http.StatusBadRequest, "仅草稿或失败状态的任务可以修改")
		return
	}

	var req struct {
		FinalTitle  string              `json:"final_title"`
		FinalText   string              `json:"final_text"`
		FinalImages model.StringSlice   `json:"final_images"`
		FinalTags   model.StringSlice   `json:"final_tags"`
		AccountID   uint64              `json:"account_id"`
		ScheduledAt string              `json:"scheduled_at"` // "2006-01-02 15:04:05"
	}
	c.ShouldBindJSON(&req)

	updates := map[string]any{"status": model.PublishStatusDraft}
	if req.FinalTitle != "" { updates["final_title"] = req.FinalTitle }
	if req.FinalText != "" { updates["final_text"] = req.FinalText }
	if len(req.FinalImages) > 0 { updates["final_images"] = req.FinalImages }
	if len(req.FinalTags) > 0 { updates["final_tags"] = req.FinalTags }
	if req.AccountID > 0 { updates["account_id"] = req.AccountID }
	if req.ScheduledAt != "" {
		t, err := time.Parse("2006-01-02 15:04:05", req.ScheduledAt)
		if err == nil {
			updates["scheduled_at"] = t
			updates["status"] = model.PublishStatusScheduled
		}
	}

	h.db.Model(&task).Updates(updates)
	response.OK(c, task)
}

// Approve 审核通过（状态 draft -> approved）
func (h *PublishHandler) Approve(c *gin.Context) {
	id := parseUint(c.Param("id"))
	operatorID := getOperatorID(c)
	now := time.Now()
	h.db.Model(&model.PublishTask{}).Where("id = ?", id).Updates(map[string]any{
		"status":      model.PublishStatusApproved,
		"reviewed_by": operatorID,
		"reviewed_at": now,
	})
	response.OK(c, nil)
}

// Reject 审核拒绝（回到 draft）
func (h *PublishHandler) Reject(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)
	h.db.Model(&model.PublishTask{}).Where("id = ?", id).Updates(map[string]any{
		"status":         model.PublishStatusDraft,
		"failure_reason": req.Reason,
	})
	response.OK(c, nil)
}

// Publish 立即发布（需 approved 状态）
func (h *PublishHandler) Publish(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var task model.PublishTask
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&task).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "发布任务不存在")
		return
	}
	if task.Status != model.PublishStatusApproved && task.Status != model.PublishStatusDraft {
		response.Fail(c, http.StatusBadRequest, "仅审核通过或草稿状态可发布")
		return
	}
	if task.AccountID == 0 {
		response.Fail(c, http.StatusBadRequest, "请先关联发布账号")
		return
	}

	operatorID := getOperatorID(c)
	ip := c.ClientIP()

	go func() {
		h.publishSvc.Publish(&task, operatorID, ip)
	}()

	response.OK(c, gin.H{"message": "发布任务已提交，请稍后刷新状态"})
}

// Delete 删除草稿
func (h *PublishHandler) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	h.db.Model(&model.PublishTask{}).
		Where("id = ? AND status = 'draft'", id).
		Update("deleted_at", gorm.Expr("NOW()"))
	response.OK(c, nil)
}

// Stats 发布统计
func (h *PublishHandler) Stats(c *gin.Context) {
	type stat struct {
		Platform string `json:"platform"`
		Status   string `json:"status"`
		Count    int64  `json:"count"`
	}
	var stats []stat
	h.db.Model(&model.PublishTask{}).Where("deleted_at IS NULL").
		Select("target_platform as platform, status, COUNT(*) as count").
		Group("target_platform, status").
		Scan(&stats)
	response.OK(c, stats)
}

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

// TaskHandler 采集任务管理
type TaskHandler struct {
	db         *gorm.DB
	scraperSvc *service.ScraperService
}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		db:         database.Get(),
		scraperSvc: service.NewScraperService(),
	}
}

// List 采集任务列表
func (h *TaskHandler) List(c *gin.Context) {
	var req struct {
		model.PageReq
		Platform string `form:"platform"`
		Status   string `form:"status"`
		OrgID    uint64 `form:"org_id"`
	}
	c.ShouldBindQuery(&req)
	req.Normalize()

	q := h.db.Model(&model.ScrapeTask{}).Where("deleted_at IS NULL")
	if req.Platform != "" {
		q = q.Where("platform = ?", req.Platform)
	}
	if req.Status != "" {
		q = q.Where("status = ?", req.Status)
	}
	if req.OrgID > 0 {
		q = q.Where("org_id = ?", req.OrgID)
	}

	var total int64
	q.Count(&total)
	var tasks []model.ScrapeTask
	q.Order("id DESC").Limit(req.PageSize).Offset(req.Offset()).Find(&tasks)

	response.OK(c, gin.H{"list": tasks, "total": total, "page": req.Page, "page_size": req.PageSize})
}

// Create 创建采集任务
func (h *TaskHandler) Create(c *gin.Context) {
	var req struct {
		OrgID          uint64 `json:"org_id"`
		TaskName       string `json:"task_name"       binding:"required"`
		Platform       string `json:"platform"        binding:"required"`
		TaskType       string `json:"task_type"       binding:"required"`
		TargetParam    string `json:"target_param"    binding:"required"`
		TargetPlatform string `json:"target_platform" binding:"required"`
		AccountID      uint64 `json:"account_id"`
		CronExpr       string `json:"cron_expr"`
		FetchLimit     int    `json:"fetch_limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	if req.FetchLimit <= 0 || req.FetchLimit > 50 {
		req.FetchLimit = 5
	}

	task := &model.ScrapeTask{
		OrgID:          req.OrgID,
		TaskName:       req.TaskName,
		Platform:       req.Platform,
		TaskType:       req.TaskType,
		TargetParam:    req.TargetParam,
		TargetPlatform: req.TargetPlatform,
		AccountID:      req.AccountID,
		CronExpr:       req.CronExpr,
		FetchLimit:     req.FetchLimit,
		Status:         model.TaskStatusEnabled,
	}
	if err := h.db.Create(task).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, "创建任务失败: "+err.Error())
		return
	}
	response.OK(c, task)
}

// Update 更新采集任务
func (h *TaskHandler) Update(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var task model.ScrapeTask
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&task).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "任务不存在")
		return
	}

	var req struct {
		TaskName       string `json:"task_name"`
		TargetParam    string `json:"target_param"`
		TargetPlatform string `json:"target_platform"`
		AccountID      uint64 `json:"account_id"`
		CronExpr       string `json:"cron_expr"`
		FetchLimit     int    `json:"fetch_limit"`
		Status         *int8  `json:"status"`
	}
	c.ShouldBindJSON(&req)

	updates := map[string]any{}
	if req.TaskName != "" { updates["task_name"] = req.TaskName }
	if req.TargetParam != "" { updates["target_param"] = req.TargetParam }
	if req.TargetPlatform != "" { updates["target_platform"] = req.TargetPlatform }
	if req.AccountID > 0 { updates["account_id"] = req.AccountID }
	if req.CronExpr != "" { updates["cron_expr"] = req.CronExpr }
	if req.FetchLimit > 0 { updates["fetch_limit"] = req.FetchLimit }
	if req.Status != nil { updates["status"] = *req.Status }

	h.db.Model(&task).Updates(updates)
	response.OK(c, task)
}

// Delete 删除任务（软删除）
func (h *TaskHandler) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	h.db.Model(&model.ScrapeTask{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", gorm.Expr("NOW()"))
	response.OK(c, nil)
}

// Run 立即执行一次采集任务
func (h *TaskHandler) Run(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var task model.ScrapeTask
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&task).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "任务不存在")
		return
	}

	// 异步执行，立即返回
	go func() {
		count, err := h.scraperSvc.RunTask(&task)
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		h.db.Model(&task).Updates(map[string]any{
			"last_run_at": time.Now(),
			"last_error":  errMsg,
		})
		_ = count
	}()

	response.OK(c, gin.H{"message": "任务已开始执行，结果请查看内容列表"})
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"unionManageCenter/cms/internal/model"
	"unionManageCenter/cms/internal/service"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/response"
)

// ContentHandler 原始内容 + 流水线处理
type ContentHandler struct {
	db *gorm.DB
}

func NewContentHandler() *ContentHandler {
	return &ContentHandler{db: database.Get()}
}

// List 内容列表（带流水线状态）
func (h *ContentHandler) List(c *gin.Context) {
	var req struct {
		model.PageReq
		Platform   string `form:"platform"`
		ProcStatus string `form:"proc_status"`
		TaskID     uint64 `form:"task_id"`
		Keyword    string `form:"keyword"`
	}
	c.ShouldBindQuery(&req)
	req.Normalize()

	q := h.db.Model(&model.RawContent{})
	if req.Platform != "" {
		q = q.Where("platform = ?", req.Platform)
	}
	if req.ProcStatus != "" {
		q = q.Where("proc_status = ?", req.ProcStatus)
	}
	if req.TaskID > 0 {
		q = q.Where("task_id = ?", req.TaskID)
	}
	if req.Keyword != "" {
		q = q.Where("title LIKE ?", "%"+req.Keyword+"%")
	}

	var total int64
	q.Count(&total)
	var contents []model.RawContent
	q.Order("id DESC").Limit(req.PageSize).Offset(req.Offset()).Find(&contents)

	response.OK(c, gin.H{"list": contents, "total": total, "page": req.Page, "page_size": req.PageSize})
}

// Get 查看内容详情 + 所有处理记录
func (h *ContentHandler) Get(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var raw model.RawContent
	if err := h.db.First(&raw, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "内容不存在")
		return
	}

	var records []model.ProcessRecord
	h.db.Where("raw_id = ?", id).Order("round ASC").Find(&records)

	response.OK(c, gin.H{"content": raw, "process_records": records})
}

// Process 手动触发 AI 流水线处理
func (h *ContentHandler) Process(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var raw model.RawContent
	if err := h.db.First(&raw, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "内容不存在")
		return
	}

	var req struct {
		TargetPlatform string `json:"target_platform" binding:"required"`
		RewriterDriver string `json:"rewriter_driver"` // 可选，默认 ollama_free
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	if req.RewriterDriver == "" {
		req.RewriterDriver = "ollama_free"
	}

	// 异步处理
	go func() {
		pipeline := service.NewPipelineService(req.RewriterDriver, []string{"wordlist_free"})
		finalText, err := pipeline.ProcessRaw(&raw, req.TargetPlatform)
		if err == nil && finalText != "" {
			// 自动创建发布草稿
			h.db.Create(&model.PublishTask{
				OrgID:          raw.OrgID,
				RawID:          raw.ID,
				TargetPlatform: req.TargetPlatform,
				FinalTitle:     raw.Title,
				FinalText:      finalText,
				Status:         model.PublishStatusDraft,
			})
		}
	}()

	response.OK(c, gin.H{"message": "流水线处理已启动，完成后将自动生成发布草稿"})
}

// Skip 跳过该内容（标记为 skipped）
func (h *ContentHandler) Skip(c *gin.Context) {
	id := parseUint(c.Param("id"))
	h.db.Model(&model.RawContent{}).Where("id = ?", id).
		Update("proc_status", model.ProcStatusSkipped)
	response.OK(c, nil)
}

// Stats 流水线状态统计
func (h *ContentHandler) Stats(c *gin.Context) {
	type stat struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var stats []stat
	h.db.Model(&model.RawContent{}).
		Select("proc_status as status, COUNT(*) as count").
		Group("proc_status").
		Scan(&stats)
	response.OK(c, stats)
}

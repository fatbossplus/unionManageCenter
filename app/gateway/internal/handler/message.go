package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"unionManageCenter/gateway/internal/model"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/middleware"
	"unionManageCenter/pkg/response"
)

type MessageHandler struct{ db *gorm.DB }

func NewMessageHandler() *MessageHandler { return &MessageHandler{db: database.Get()} }

func (h *MessageHandler) List(c *gin.Context) {
	var req struct {
		model.PageReq
		Type   string `form:"type"`
		IsRead string `form:"is_read"`
	}
	c.ShouldBindQuery(&req)
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	userID := middleware.GetUserID(c)
	db := h.db.Model(&model.Message{}).Where("user_id = ? OR user_id = 0", userID)
	if req.Type != "" {
		db = db.Where("type = ?", req.Type)
	}
	if req.IsRead != "" {
		db = db.Where("is_read = ?", req.IsRead)
	}

	var total int64
	db.Count(&total)
	var msgs []model.Message
	db.Offset(req.Offset()).Limit(req.PageSize).Order("created_at DESC").Find(&msgs)
	response.OK(c, model.PageResult{List: msgs, Total: total, Page: req.Page, Size: req.PageSize})
}

func (h *MessageHandler) MarkRead(c *gin.Context) {
	var req struct {
		IDs []uint64 `json:"ids"`
	}
	c.ShouldBindJSON(&req)
	now := time.Now()
	h.db.Model(&model.Message{}).Where("id IN ?", req.IDs).Updates(map[string]any{
		"is_read": 1, "read_at": now,
	})
	response.OKMsg(c, "已标记已读")
}

func (h *MessageHandler) MarkAllRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	now := time.Now()
	h.db.Model(&model.Message{}).Where("(user_id = ? OR user_id = 0) AND is_read = 0", userID).Updates(map[string]any{
		"is_read": 1, "read_at": now,
	})
	response.OKMsg(c, "全部已读")
}

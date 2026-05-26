package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"unionManageCenter/gateway/internal/model"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/response"
)

type OrderHandler struct{ db *gorm.DB }

func NewOrderHandler() *OrderHandler { return &OrderHandler{db: database.Get()} }

func (h *OrderHandler) List(c *gin.Context) {
	var req struct {
		model.PageReq
		Keyword   string `form:"keyword"`
		Type      string `form:"type"`
		Status    string `form:"status"`
		PayMethod string `form:"pay_method"`
		MinAmount string `form:"min_amount"`
		MaxAmount string `form:"max_amount"`
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

	db := h.db.Model(&model.Order{}).Where("deleted_at IS NULL")
	if req.Keyword != "" {
		db = db.Where("order_no LIKE ?", "%"+req.Keyword+"%")
	}
	if req.Type != "" {
		db = db.Where("type = ?", req.Type)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.PayMethod != "" {
		db = db.Where("pay_method = ?", req.PayMethod)
	}
	if req.MinAmount != "" {
		db = db.Where("amount >= ?", req.MinAmount)
	}
	if req.MaxAmount != "" {
		db = db.Where("amount <= ?", req.MaxAmount)
	}
	if req.StartDate != "" {
		db = db.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		db = db.Where("created_at <= ?", req.EndDate+" 23:59:59")
	}

	var total int64
	db.Count(&total)
	var orders []model.Order
	db.Preload("User").Preload("Org").Offset(req.Offset()).Limit(req.PageSize).Order("created_at DESC").Find(&orders)
	response.OK(c, model.PageResult{List: orders, Total: total, Page: req.Page, Size: req.PageSize})
}

func (h *OrderHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var order model.Order
	if err := h.db.Preload("User").Preload("Org").Where("id = ? AND deleted_at IS NULL", id).First(&order).Error; err != nil {
		response.Fail(c, 404, "订单不存在")
		return
	}
	response.OK(c, order)
}

func (h *OrderHandler) Refund(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)
	now := time.Now()
	h.db.Model(&model.Order{}).Where("id = ? AND status = 2", id).Updates(map[string]any{
		"status": 3, "refunded_at": now, "refund_reason": req.Reason,
	})
	response.OKMsg(c, "退款成功")
}

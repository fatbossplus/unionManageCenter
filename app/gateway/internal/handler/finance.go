package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"unionManageCenter/gateway/internal/model"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/middleware"
	"unionManageCenter/pkg/response"
)

type FinanceHandler struct{ db *gorm.DB }

func NewFinanceHandler() *FinanceHandler { return &FinanceHandler{db: database.Get()} }

func (h *FinanceHandler) List(c *gin.Context) {
	var req struct {
		model.PageReq
		Status    string `form:"status"`
		Period    string `form:"period"`
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

	db := h.db.Model(&model.FinanceSettlement{}).Where("deleted_at IS NULL")
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Period != "" {
		db = db.Where("period = ?", req.Period)
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
	var list []model.FinanceSettlement
	db.Preload("Org").Offset(req.Offset()).Limit(req.PageSize).Order("created_at DESC").Find(&list)
	response.OK(c, model.PageResult{List: list, Total: total, Page: req.Page, Size: req.PageSize})
}

func (h *FinanceHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var item model.FinanceSettlement
	if err := h.db.Preload("Org").Where("id = ? AND deleted_at IS NULL", id).First(&item).Error; err != nil {
		response.Fail(c, 404, "记录不存在")
		return
	}
	response.OK(c, item)
}

func (h *FinanceHandler) Settle(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	operatorID := middleware.GetUserID(c)
	now := time.Now()
	h.db.Model(&model.FinanceSettlement{}).Where("id = ? AND status = 1", id).Updates(map[string]any{
		"status": 3, "settled_at": now, "operator_id": operatorID,
	})
	response.OKMsg(c, "结算成功")
}

func (h *FinanceHandler) ListAccounts(c *gin.Context) {
	orgID := c.Query("org_id")
	db := h.db.Model(&model.FinanceAccount{}).Where("deleted_at IS NULL")
	if orgID != "" {
		db = db.Where("org_id = ?", orgID)
	}
	var accounts []model.FinanceAccount
	db.Find(&accounts)
	response.OK(c, accounts)
}

func (h *FinanceHandler) CreateAccount(c *gin.Context) {
	var account model.FinanceAccount
	if err := c.ShouldBindJSON(&account); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	h.db.Create(&account)
	response.OK(c, account)
}

package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"unionManageCenter/cms/internal/model"
	"unionManageCenter/cms/internal/service"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/response"
)

// AccountHandler 平台账号管理
type AccountHandler struct {
	db      *gorm.DB
	credSvc *service.CredentialService
}

func NewAccountHandler() *AccountHandler {
	return &AccountHandler{
		db:      database.Get(),
		credSvc: service.NewCredentialService(),
	}
}

// List 账号列表
func (h *AccountHandler) List(c *gin.Context) {
	var req struct {
		model.PageReq
		Platform string `form:"platform"`
		Status   string `form:"status"`
		OrgID    uint64 `form:"org_id"`
	}
	c.ShouldBindQuery(&req)
	req.Normalize()

	q := h.db.Model(&model.PlatformAccount{}).Where("deleted_at IS NULL")
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

	var accounts []model.PlatformAccount
	q.Order("id DESC").Limit(req.PageSize).Offset(req.Offset()).Find(&accounts)

	response.OK(c, gin.H{"list": accounts, "total": total, "page": req.Page, "page_size": req.PageSize})
}

// Create 新增账号（凭证加密入库）
func (h *AccountHandler) Create(c *gin.Context) {
	var req struct {
		OrgID       uint64 `json:"org_id"`
		Platform    string `json:"platform"   binding:"required"`
		AccountName string `json:"account_name" binding:"required"`
		AccountUID  string `json:"account_uid"`
		CredJSON    string `json:"cred_json"  binding:"required"` // 明文凭证JSON
		ExpiresAt   string `json:"expires_at"`
		Remark      string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 验证凭证 JSON 格式
	if !json.Valid([]byte(req.CredJSON)) {
		response.Fail(c, http.StatusBadRequest, "cred_json 必须是合法的 JSON 格式")
		return
	}

	// 先创建记录拿到 ID（用于派生密钥）
	acc := &model.PlatformAccount{
		OrgID:       req.OrgID,
		Platform:    req.Platform,
		AccountName: req.AccountName,
		AccountUID:  req.AccountUID,
		Status:      model.AccountStatusNormal,
		Remark:      req.Remark,
		CredCipher:  "placeholder",
		CredIV:      "placeholder",
	}
	if req.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", req.ExpiresAt)
		if err == nil {
			acc.ExpiresAt = &t
		}
	}
	if err := h.db.Create(acc).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, "创建账号失败: "+err.Error())
		return
	}

	// 用 ID 加密凭证
	cipher, iv, err := service.EncryptCred(acc.ID, req.CredJSON)
	if err != nil {
		h.db.Delete(acc)
		response.Fail(c, http.StatusInternalServerError, "凭证加密失败: "+err.Error())
		return
	}
	h.db.Model(acc).Updates(map[string]any{"cred_cipher": cipher, "cred_iv": iv})

	// 记录审计
	operatorID := getOperatorID(c)
	h.credSvc.WriteAudit(acc.ID, operatorID, "create", "新增账号", c.ClientIP())

	response.OK(c, acc)
}

// Update 更新账号（可选更新凭证）
func (h *AccountHandler) Update(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var acc model.PlatformAccount
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&acc).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "账号不存在")
		return
	}

	var req struct {
		AccountName string `json:"account_name"`
		AccountUID  string `json:"account_uid"`
		Status      *int8  `json:"status"`
		Remark      string `json:"remark"`
		CredJSON    string `json:"cred_json"` // 可选，不传则不更新凭证
		ExpiresAt   string `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误")
		return
	}

	updates := map[string]any{}
	if req.AccountName != "" {
		updates["account_name"] = req.AccountName
	}
	if req.AccountUID != "" {
		updates["account_uid"] = req.AccountUID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.ExpiresAt != "" {
		t, _ := time.Parse("2006-01-02", req.ExpiresAt)
		updates["expires_at"] = t
	}

	// 更新凭证（需重新加密）
	if req.CredJSON != "" {
		if !json.Valid([]byte(req.CredJSON)) {
			response.Fail(c, http.StatusBadRequest, "cred_json 格式错误")
			return
		}
		cipher, iv, err := service.EncryptCred(acc.ID, req.CredJSON)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "凭证加密失败")
			return
		}
		updates["cred_cipher"] = cipher
		updates["cred_iv"] = iv
		updates["cred_version"] = acc.CredVersion + 1

		operatorID := getOperatorID(c)
		h.credSvc.WriteAudit(acc.ID, operatorID, "update", "更新凭证", c.ClientIP())
	}

	h.db.Model(&acc).Updates(updates)
	response.OK(c, acc)
}

// Delete 删除账号（软删除）
func (h *AccountHandler) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := h.db.Model(&model.PlatformAccount{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", gorm.Expr("NOW()")).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, "删除失败")
		return
	}
	operatorID := getOperatorID(c)
	h.credSvc.WriteAudit(id, operatorID, "revoke", "删除账号", c.ClientIP())
	response.OK(c, nil)
}

// AuditLogs 查询凭证审计日志
func (h *AccountHandler) AuditLogs(c *gin.Context) {
	accountID := parseUint(c.Param("id"))
	var req struct{ model.PageReq }
	c.ShouldBindQuery(&req)
	req.Normalize()

	var total int64
	h.db.Model(&model.CredentialAudit{}).Where("account_id = ?", accountID).Count(&total)

	var logs []model.CredentialAudit
	h.db.Where("account_id = ?", accountID).
		Order("id DESC").
		Limit(req.PageSize).Offset(req.Offset()).
		Find(&logs)

	response.OK(c, gin.H{"list": logs, "total": total})
}

// ─── 工具函数 ────────────────────────────────────────────────────────────────

func parseUint(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}

func getOperatorID(c *gin.Context) uint64 {
	if v, exists := c.Get("admin_id"); exists {
		if id, ok := v.(uint64); ok {
			return id
		}
		if id, ok := v.(float64); ok {
			return uint64(id)
		}
	}
	return 0
}

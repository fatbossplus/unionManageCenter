package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"unionManageCenter/cms/internal/model"
	"unionManageCenter/pkg/database"
	"unionManageCenter/pkg/response"
)

// DriverHandler 驱动配置管理（免费/付费切换）
type DriverHandler struct{ db *gorm.DB }

func NewDriverHandler() *DriverHandler { return &DriverHandler{db: database.Get()} }

// ListAll 列出所有驱动配置（附带元信息，供前端展示勾选）
func (h *DriverHandler) ListAll(c *gin.Context) {
	orgID := uint64(0) // 可从 query 读取

	var cfgs []model.DriverConfig
	h.db.Where("org_id = ? OR org_id = 0", orgID).Order("org_id DESC, config_key").Find(&cfgs)

	// 合并去重（组织级覆盖全局）
	merged := map[string]model.DriverConfig{}
	for _, cfg := range cfgs {
		if _, exist := merged[cfg.ConfigKey]; !exist {
			merged[cfg.ConfigKey] = cfg
		}
	}

	// 附加可选驱动元信息
	type Row struct {
		Config      model.DriverConfig   `json:"config"`
		Available   []model.DriverMeta   `json:"available"`
	}

	var rows []Row
	for key, cfg := range merged {
		rows = append(rows, Row{
			Config:    cfg,
			Available: model.AllDriverMeta[key],
		})
	}

	response.OK(c, rows)
}

// Update 更新驱动配置
func (h *DriverHandler) Update(c *gin.Context) {
	var req struct {
		OrgID      uint64 `json:"org_id"`
		ConfigKey  string `json:"config_key"  binding:"required"`
		DriverName string `json:"driver_name" binding:"required"`
		DriverType string `json:"driver_type"`
		ConfigJSON string `json:"config_json"` // 驱动参数JSON（含API Key，明文，前端应提示安全）
		Enabled    *int8  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	updates := map[string]any{
		"driver_name": req.DriverName,
	}
	if req.DriverType != "" {
		updates["driver_type"] = req.DriverType
	}
	if req.ConfigJSON != "" {
		updates["config_json"] = req.ConfigJSON
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	result := h.db.Model(&model.DriverConfig{}).
		Where("org_id = ? AND config_key = ?", req.OrgID, req.ConfigKey).
		Updates(updates)

	if result.RowsAffected == 0 {
		// 不存在则创建
		cfg := &model.DriverConfig{
			OrgID:      req.OrgID,
			ConfigKey:  req.ConfigKey,
			DriverName: req.DriverName,
			DriverType: req.DriverType,
			ConfigJSON: req.ConfigJSON,
		}
		if req.Enabled != nil {
			cfg.Enabled = *req.Enabled
		} else {
			cfg.Enabled = 1
		}
		if err := h.db.Create(cfg).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "保存配置失败: "+err.Error())
			return
		}
	}

	response.OK(c, nil)
}

// DriverMeta 获取所有可用驱动的元信息（前端展示用）
func (h *DriverHandler) DriverMeta(c *gin.Context) {
	response.OK(c, model.AllDriverMeta)
}

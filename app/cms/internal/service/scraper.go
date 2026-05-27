// Package service 采集服务：根据驱动配置调度采集适配器
package service

import (
	"crypto/sha256"
	"fmt"
	"time"

	"gorm.io/gorm"
	"unionManageCenter/cms/internal/adapter"
	"unionManageCenter/cms/internal/model"
	"unionManageCenter/pkg/database"
)

// ScraperService 采集服务
type ScraperService struct {
	db *gorm.DB
}

func NewScraperService() *ScraperService {
	return &ScraperService{db: database.Get()}
}

// RunTask 执行一次采集任务，返回新增内容数量
func (s *ScraperService) RunTask(task *model.ScrapeTask) (int, error) {
	// 获取该平台配置的采集驱动
	scraperName, err := s.getDriverName(task.OrgID, task.Platform)
	if err != nil {
		return 0, err
	}

	scraper, ok := adapter.GetScraper(scraperName)
	if !ok {
		return 0, fmt.Errorf("采集驱动 %q 未注册", scraperName)
	}

	// 执行采集
	var articles []adapter.Article
	switch task.TaskType {
	case model.TaskTypeSearchTitle:
		articles, err = scraper.SearchByTitle(task.TargetParam, task.FetchLimit)
	case model.TaskTypeFollowAuthor:
		articles, err = scraper.FetchByAuthor(task.TargetParam, task.FetchLimit)
	default:
		return 0, fmt.Errorf("未知任务类型: %s", task.TaskType)
	}
	if err != nil {
		return 0, err
	}

	// 去重写入
	newCount := 0
	for _, art := range articles {
		hash := articleHash(art.URL, art.Title)
		var exist model.RawContent
		if s.db.Where("source_hash = ?", hash).First(&exist).Error == nil {
			continue // 已存在
		}
		raw := &model.RawContent{
			TaskID:     task.ID,
			OrgID:      task.OrgID,
			Platform:   task.Platform,
			SourceURL:  art.URL,
			SourceHash: hash,
			Title:      art.Title,
			Author:     art.Author,
			BodyText:   art.BodyText,
			BodyImages: art.Images,
			FetchedAt:  time.Now(),
			ProcStatus: model.ProcStatusPending,
		}
		if err := s.db.Create(raw).Error; err == nil {
			newCount++
		}
	}

	// 更新任务最后执行时间
	now := time.Now()
	s.db.Model(task).Updates(map[string]any{
		"last_run_at": now,
		"last_error":  "",
	})

	return newCount, nil
}

// getDriverName 从驱动配置表读取启用的驱动名
func (s *ScraperService) getDriverName(orgID uint64, platform string) (string, error) {
	key := platform + ".scraper"
	var cfg model.DriverConfig
	// 先查组织级，再查全局默认
	err := s.db.Where("(org_id = ? OR org_id = 0) AND config_key = ? AND enabled = 1", orgID, key).
		Order("org_id DESC").
		First(&cfg).Error
	if err != nil {
		return "", fmt.Errorf("平台 %s 未配置采集驱动", platform)
	}
	return cfg.DriverName, nil
}

func articleHash(url, title string) string {
	h := sha256.Sum256([]byte(url + "|" + title))
	return fmt.Sprintf("%x", h[:])
}

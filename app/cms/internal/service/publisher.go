// Package service 发布服务
package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"unionManageCenter/cms/internal/adapter"
	"unionManageCenter/cms/internal/model"
	"unionManageCenter/pkg/database"
)

// PublisherService 发布服务
type PublisherService struct {
	db      *gorm.DB
	credSvc *CredentialService
}

func NewPublisherService() *PublisherService {
	return &PublisherService{
		db:      database.Get(),
		credSvc: NewCredentialService(),
	}
}

// Publish 执行一次发布
func (p *PublisherService) Publish(task *model.PublishTask, operatorID uint64, ip string) error {
	pub, ok := adapter.GetPublisher(task.TargetPlatform)
	if !ok {
		// 尝试按平台名找默认发布驱动
		pub, ok = p.findPublisherForPlatform(task.TargetPlatform)
		if !ok {
			return fmt.Errorf("平台 %s 无可用发布驱动", task.TargetPlatform)
		}
	}

	// 解密凭证
	credJSON, err := p.credSvc.GetDecrypted(task.AccountID, operatorID, "publish", ip)
	if err != nil {
		return fmt.Errorf("获取发布账号凭证失败: %w", err)
	}

	// 更新状态为发布中
	p.db.Model(task).Update("status", model.PublishStatusScheduled)

	result, err := pub.Publish(adapter.PublishRequest{
		Title:  task.FinalTitle,
		Text:   task.FinalText,
		Images: task.FinalImages,
		Tags:   task.FinalTags,
		Cred:   credJSON,
	})

	if err != nil {
		task.RetryCount++
		p.db.Model(task).Updates(map[string]any{
			"status":         model.PublishStatusFailed,
			"failure_reason": err.Error(),
			"retry_count":    task.RetryCount,
		})
		return err
	}

	now := time.Now()
	p.db.Model(task).Updates(map[string]any{
		"status":           model.PublishStatusPublished,
		"published_at":     now,
		"platform_post_id": result.PostID,
		"failure_reason":   "",
	})
	return nil
}

func (p *PublisherService) findPublisherForPlatform(platform string) (adapter.Publisher, bool) {
	// 从注册表遍历找第一个匹配平台的 publisher
	// 实际实现通过 adapter.GetPublisher(driverName) 配合 driver config 查询
	switch platform {
	case model.PlatformWechat:
		return adapter.GetPublisher("official_api")
	case model.PlatformCSdn:
		return adapter.GetPublisher("unofficial_api")
	case model.PlatformRednote:
		return adapter.GetPublisher("playwright_free")
	case model.PlatformDouyin:
		return adapter.GetPublisher("openapi_free")
	}
	return nil, false
}

// Package scraper 付费采集驱动桩（接口已定义，配置 API Key 后可接入）
package scraper

import (
	"errors"
	"unionManageCenter/cms/internal/adapter"
)

func init() {
	adapter.RegisterScraper(&NewrankScraper{})
	adapter.RegisterScraper(&QianguaScraper{})
	adapter.RegisterScraper(&PlaywrightScraper{})
}

// NewrankScraper 新榜 API（付费，¥500-2000/月）
type NewrankScraper struct{ APIKey string }

func (n *NewrankScraper) Name() string { return "newrank_paid" }
func (n *NewrankScraper) SearchByTitle(keyword string, limit int) ([]adapter.Article, error) {
	if n.APIKey == "" {
		return nil, errors.New("新榜 API：请先在驱动配置中填入 api_key")
	}
	// TODO: POST https://api.newrank.cn/api/...
	return nil, errors.New("新榜 API 驱动开发中，请联系服务商获取接口文档后实现")
}
func (n *NewrankScraper) FetchByAuthor(authorID string, limit int) ([]adapter.Article, error) {
	return n.SearchByTitle(authorID, limit)
}

// QianguaScraper 千瓜数据（付费，¥500-3000/月，小红书专用）
type QianguaScraper struct{ APIKey string }

func (q *QianguaScraper) Name() string { return "qiangua_paid" }
func (q *QianguaScraper) SearchByTitle(keyword string, limit int) ([]adapter.Article, error) {
	if q.APIKey == "" {
		return nil, errors.New("千瓜数据 API：请先在驱动配置中填入 api_key")
	}
	// TODO: POST https://api.qiangua.com/...
	return nil, errors.New("千瓜数据驱动开发中")
}
func (q *QianguaScraper) FetchByAuthor(authorID string, limit int) ([]adapter.Article, error) {
	return q.SearchByTitle(authorID, limit)
}

// PlaywrightScraper Playwright 浏览器自动化（免费，需部署浏览器服务）
// 通过外部 Playwright 微服务接口调用，避免在 Go 进程内嵌入浏览器
type PlaywrightScraper struct {
	ServiceURL string // Playwright 代理服务地址，如 http://localhost:3100
}

func (p *PlaywrightScraper) Name() string { return "playwright_free" }
func (p *PlaywrightScraper) SearchByTitle(keyword string, limit int) ([]adapter.Article, error) {
	if p.ServiceURL == "" {
		return nil, errors.New("Playwright 驱动：请先配置 service_url（Playwright 服务地址）")
	}
	// TODO: 调用外部 Playwright 微服务 POST /scrape/search
	return nil, errors.New("Playwright 服务未启动，请部署 playwright-service 容器后配置地址")
}
func (p *PlaywrightScraper) FetchByAuthor(authorID string, limit int) ([]adapter.Article, error) {
	if p.ServiceURL == "" {
		return nil, errors.New("Playwright 驱动：请先配置 service_url")
	}
	// TODO: POST /scrape/author
	return nil, errors.New("Playwright 服务未启动")
}

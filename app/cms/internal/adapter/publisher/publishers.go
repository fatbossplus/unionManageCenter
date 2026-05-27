// Package publisher 各平台发布适配器
package publisher

import (
	"encoding/json"
	"errors"

	"unionManageCenter/cms/internal/adapter"
)

func init() {
	adapter.RegisterPublisher(&WechatPublisher{})
	adapter.RegisterPublisher(&CsdnPublisher{})
	adapter.RegisterPublisher(&RednotePublisher{})
	adapter.RegisterPublisher(&DouyinPublisher{})
}

// ─── 微信公众号 ─────────────────────────────────────────────────────────────

// WechatPublisher 微信公众号发布（需申请公众号开发者权限）
// 凭证 JSON 格式：{"app_id":"xxx","app_secret":"xxx","author_name":"xxx"}
type WechatPublisher struct{}

func (w *WechatPublisher) Name() string     { return "official_api" }
func (w *WechatPublisher) Platform() string { return "wechat" }

func (w *WechatPublisher) Publish(req adapter.PublishRequest) (adapter.PublishResult, error) {
	var cred struct {
		AppID      string `json:"app_id"`
		AppSecret  string `json:"app_secret"`
		AuthorName string `json:"author_name"`
	}
	if err := json.Unmarshal([]byte(req.Cred), &cred); err != nil {
		return adapter.PublishResult{}, errors.New("微信凭证格式错误，需包含 app_id / app_secret")
	}
	if cred.AppID == "" || cred.AppSecret == "" {
		return adapter.PublishResult{}, errors.New("微信凭证缺少 app_id 或 app_secret")
	}

	// TODO: 完整实现步骤：
	// 1. POST https://api.weixin.qq.com/cgi-bin/token 获取 access_token
	// 2. POST https://api.weixin.qq.com/cgi-bin/media/uploadimg 上传图片
	// 3. POST https://api.weixin.qq.com/cgi-bin/draft/add 创建草稿
	// 4. POST https://api.weixin.qq.com/cgi-bin/freepublish/submit 发布
	return adapter.PublishResult{}, errors.New("微信发布驱动：接口框架已就绪，请补充 OAuth 实现")
}

// ─── CSDN ───────────────────────────────────────────────────────────────────

// CsdnPublisher CSDN 非官方API发布（通过 Cookie 模拟）
// 凭证 JSON 格式：{"cookie":"..."}
type CsdnPublisher struct{}

func (c *CsdnPublisher) Name() string     { return "unofficial_api" }
func (c *CsdnPublisher) Platform() string { return "csdn" }

func (c *CsdnPublisher) Publish(req adapter.PublishRequest) (adapter.PublishResult, error) {
	var cred struct {
		Cookie string `json:"cookie"`
	}
	if err := json.Unmarshal([]byte(req.Cred), &cred); err != nil || cred.Cookie == "" {
		return adapter.PublishResult{}, errors.New("CSDN凭证格式错误，需包含 cookie 字段")
	}

	// TODO: 调用 CSDN 内部 API（非官方，可能因页面改版失效）
	// POST https://bizapi.csdn.net/blog-console-api/v3/mdeditor/saveArticle
	// Header: Cookie: ...
	return adapter.PublishResult{}, errors.New("CSDN发布驱动：凭证格式验证通过，待补充发布实现")
}

// ─── 小红书 ─────────────────────────────────────────────────────────────────

// RednotePublisher 小红书发布（Playwright 自动化）
// 凭证 JSON 格式：{"phone":"xxx","cookie":"..."}
type RednotePublisher struct{}

func (r *RednotePublisher) Name() string     { return "playwright_free" }
func (r *RednotePublisher) Platform() string { return "rednote" }

func (r *RednotePublisher) Publish(req adapter.PublishRequest) (adapter.PublishResult, error) {
	// 小红书无开放发布API，依赖 Playwright 浏览器自动化
	// 需要部署外部 playwright-service 容器
	return adapter.PublishResult{}, errors.New("小红书发布：需先部署 Playwright 浏览器服务，详见 docs/playwright-service.md")
}

// ─── 抖音 ───────────────────────────────────────────────────────────────────

// DouyinPublisher 抖音图文发布（企业号开放API）
// 凭证 JSON 格式：{"client_key":"xxx","client_secret":"xxx","access_token":"..."}
type DouyinPublisher struct{}

func (d *DouyinPublisher) Name() string     { return "openapi_free" }
func (d *DouyinPublisher) Platform() string { return "douyin" }

func (d *DouyinPublisher) Publish(req adapter.PublishRequest) (adapter.PublishResult, error) {
	var cred struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal([]byte(req.Cred), &cred); err != nil || cred.AccessToken == "" {
		return adapter.PublishResult{}, errors.New("抖音凭证格式错误，需包含 access_token")
	}

	// TODO: POST https://open.douyin.com/api/douyin/v1/post/publish/content/init/
	return adapter.PublishResult{}, errors.New("抖音发布驱动：待补充 Open Platform API 实现")
}

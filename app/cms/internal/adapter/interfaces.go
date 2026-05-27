// Package adapter 定义所有平台适配器接口。
// 每个接口对应一个"能力维度"，驱动可热插拔替换，通过 DriverRegistry 按配置选择。
package adapter

import "time"

// ─── 采集接口 ────────────────────────────────────────────────────────────────

// Article 采集到的原始文章
type Article struct {
	Title    string
	Author   string
	URL      string
	BodyText string
	Images   []string
	PubTime  time.Time
}

// Scraper 采集适配器接口
type Scraper interface {
	// Name 返回驱动标识，与 model.DriverName 常量一致
	Name() string
	// SearchByTitle 按标题/关键词搜索文章，limit 为最大条数
	SearchByTitle(keyword string, limit int) ([]Article, error)
	// FetchByAuthor 抓取作者最新发布，limit 为最大条数
	FetchByAuthor(authorID string, limit int) ([]Article, error)
}

// ─── AI 改写接口 ──────────────────────────────────────────────────────────────

// RewriteRequest AI改写请求
type RewriteRequest struct {
	Text           string // 待改写文本
	Platform       string // 目标平台（影响风格）
	MaxLen         int    // 最大长度（0=不限）
	SystemPrompt   string // 自定义 system prompt（可选）
}

// RewriteResult AI改写结果
type RewriteResult struct {
	Text       string  // 改写后文本
	Confidence float64 // 置信度 0-1
	Model      string  // 使用的模型
	TokensUsed int
}

// AIRewriter AI改写适配器接口
type AIRewriter interface {
	Name() string
	Rewrite(req RewriteRequest) (RewriteResult, error)
	// SelfReview 让 AI 对内容进行6维度自审
	SelfReview(text, platform string) (SelfReviewResult, error)
}

// SelfReviewResult AI自审结果（第4轮）
type SelfReviewResult struct {
	Scores     map[string]int // 6个维度各自评分(0-10)
	Passed     bool
	Comment    string
	Issues     []string // 需要修改的问题列表
}

// ─── 合规检查接口 ─────────────────────────────────────────────────────────────

// ComplianceResult 合规检查结果
type ComplianceResult struct {
	Passed   bool
	Issues   []string // 问题描述列表
	HitWords []string // 命中的敏感词
	Score    int      // 综合合规分(0-100)
}

// ComplianceChecker 合规检查适配器接口
type ComplianceChecker interface {
	Name() string
	Check(text, platform string) (ComplianceResult, error)
}

// ─── 发布接口 ─────────────────────────────────────────────────────────────────

// PublishRequest 发布请求
type PublishRequest struct {
	Title    string
	Text     string
	Images   []string
	Tags     []string
	Cred     string // 解密后的凭证（JSON字符串）
}

// PublishResult 发布结果
type PublishResult struct {
	PostID  string // 平台返回的文章ID/URL
	PostURL string
}

// Publisher 发布适配器接口
type Publisher interface {
	Name() string
	Platform() string
	Publish(req PublishRequest) (PublishResult, error)
}

// ─── 注册表 ───────────────────────────────────────────────────────────────────

var (
	scrapers   = map[string]Scraper{}
	rewriters  = map[string]AIRewriter{}
	checkers   = map[string]ComplianceChecker{}
	publishers = map[string]Publisher{}
)

func RegisterScraper(s Scraper)           { scrapers[s.Name()] = s }
func RegisterRewriter(r AIRewriter)       { rewriters[r.Name()] = r }
func RegisterChecker(c ComplianceChecker) { checkers[c.Name()] = c }
func RegisterPublisher(p Publisher)       { publishers[p.Name()] = p }

func GetScraper(name string) (Scraper, bool)           { s, ok := scrapers[name]; return s, ok }
func GetRewriter(name string) (AIRewriter, bool)        { r, ok := rewriters[name]; return r, ok }
func GetChecker(name string) (ComplianceChecker, bool) { c, ok := checkers[name]; return c, ok }
func GetPublisher(name string) (Publisher, bool)        { p, ok := publishers[name]; return p, ok }

func AllScraperNames() []string {
	names := make([]string, 0, len(scrapers))
	for k := range scrapers { names = append(names, k) }
	return names
}

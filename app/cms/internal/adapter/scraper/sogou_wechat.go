// Package scraper 搜狗微信搜索采集驱动（免费，有反爬限制）
package scraper

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"unionManageCenter/cms/internal/adapter"
)

func init() {
	adapter.RegisterScraper(&SogouWechat{})
}

// SogouWechat 搜狗微信搜索驱动
// 注意：搜狗有反爬限制，生产建议配合代理IP使用
type SogouWechat struct {
	ProxyURL string // HTTP 代理地址（可选）
}

func (s *SogouWechat) Name() string { return "sogou_free" }

// SearchByTitle 通过搜狗微信搜索文章
func (s *SogouWechat) SearchByTitle(keyword string, limit int) ([]adapter.Article, error) {
	searchURL := fmt.Sprintf(
		"https://weixin.sogou.com/weixin?type=2&query=%s&ie=utf8&_sug_=n&_sug_type_=",
		url.QueryEscape(keyword),
	)

	body, err := s.fetch(searchURL)
	if err != nil {
		return nil, err
	}

	articles := parseWechatSearchResult(string(body), limit)
	if len(articles) == 0 {
		return nil, fmt.Errorf("搜狗微信搜索未返回结果（可能触发反爬，建议配置代理IP）")
	}
	return articles, nil
}

// FetchByAuthor 通过搜狗获取公众号最新文章
// authorID 为公众号名称（中文或英文ID均可）
func (s *SogouWechat) FetchByAuthor(authorID string, limit int) ([]adapter.Article, error) {
	// 搜狗公众号主页
	searchURL := fmt.Sprintf(
		"https://weixin.sogou.com/weixin?type=1&query=%s&ie=utf8",
		url.QueryEscape(authorID),
	)
	body, err := s.fetch(searchURL)
	if err != nil {
		return nil, err
	}

	// 从搜索结果中提取公众号链接，再抓取文章列表
	mpURL := extractMPURL(string(body))
	if mpURL == "" {
		return nil, fmt.Errorf("未找到公众号 %q，请确认公众号名称是否正确", authorID)
	}

	mpBody, err := s.fetch(mpURL)
	if err != nil {
		return nil, err
	}

	return parseWechatSearchResult(string(mpBody), limit), nil
}

func (s *SogouWechat) fetch(rawURL string) ([]byte, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	// 模拟真实浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Referer", "https://weixin.sogou.com/")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("搜狗请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 302 || resp.StatusCode == 301 {
		return nil, fmt.Errorf("搜狗触发反爬重定向（状态码%d），建议配置代理IP或降低请求频率", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("搜狗返回状态码 %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

var (
	// 匹配搜索结果中的文章链接和标题（正则解析 HTML，简化版）
	titleRe  = regexp.MustCompile(`<h3[^>]*><a[^>]*title="([^"]+)"`)
	urlRe    = regexp.MustCompile(`<h3[^>]*><a[^>]*href="([^"]+)"`)
	authorRe = regexp.MustCompile(`<span[^>]*account[^>]*>([^<]+)</span>`)
	mpURLRe  = regexp.MustCompile(`href="(https://mp\.weixin\.qq\.com/[^"]+)"`)
)

func parseWechatSearchResult(html string, limit int) []adapter.Article {
	titles := titleRe.FindAllStringSubmatch(html, limit)
	urls := urlRe.FindAllStringSubmatch(html, limit)
	authors := authorRe.FindAllStringSubmatch(html, limit)

	var articles []adapter.Article
	for i := range titles {
		if i >= limit {
			break
		}
		a := adapter.Article{
			Title:  strings.TrimSpace(titles[i][1]),
			PubTime: time.Now(),
		}
		if i < len(urls) {
			a.URL = urls[i][1]
		}
		if i < len(authors) {
			a.Author = strings.TrimSpace(authors[i][1])
		}
		if a.Title != "" {
			articles = append(articles, a)
		}
	}
	return articles
}

func extractMPURL(html string) string {
	matches := mpURLRe.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

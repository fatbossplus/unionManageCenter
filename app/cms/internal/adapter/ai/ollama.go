// Package ai Ollama 本地大模型适配器（免费，数据不出境）
package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"unionManageCenter/cms/internal/adapter"
)

func init() {
	adapter.RegisterRewriter(&OllamaRewriter{
		BaseURL: "http://localhost:11434",
		Model:   "qwen2.5:7b",
	})
}

// OllamaRewriter 调用本地 Ollama 服务进行内容改写
type OllamaRewriter struct {
	BaseURL string
	Model   string
}

func (o *OllamaRewriter) Name() string { return "ollama_free" }

func (o *OllamaRewriter) Rewrite(req adapter.RewriteRequest) (adapter.RewriteResult, error) {
	sysPrompt := req.SystemPrompt
	if sysPrompt == "" {
		sysPrompt = platformSystemPrompt(req.Platform)
	}

	userPrompt := fmt.Sprintf(`请按照以下要求改写这篇文章：
1. 保持核心内容和事实不变
2. 去除原作者署名、公众号名称、联系方式
3. 去除所有外部链接
4. 调整语言风格，使其更符合%s平台调性
5. 如有广告或引流内容，删除
6. 输出改写后的完整文章，不要包含任何解释

原文：
%s`, req.Platform, req.Text)

	body, err := o.call(sysPrompt, userPrompt)
	if err != nil {
		return adapter.RewriteResult{}, err
	}

	return adapter.RewriteResult{
		Text:  strings.TrimSpace(body),
		Model: o.Model,
	}, nil
}

func (o *OllamaRewriter) SelfReview(text, platform string) (adapter.SelfReviewResult, error) {
	prompt := fmt.Sprintf(`你是%s平台的内容审核专家。请从以下6个维度对内容进行评分（每项0-10分）：
1. 违规信息（是否含政治敏感/违法内容，10分=完全合规）
2. 广告导流（是否含营销/引流内容，10分=完全没有）
3. 语言规范（是否语言通顺规范，10分=非常规范）
4. 版权问题（是否明显抄袭或侵权，10分=无问题）
5. 平台调性（是否符合%s平台风格，10分=非常符合）
6. 歧视内容（是否含歧视性内容，10分=完全没有）

请用JSON格式输出，例如：
{"scores":{"违规信息":9,"广告导流":10,"语言规范":8,"版权问题":9,"平台调性":8,"歧视内容":10},"passed":true,"comment":"整体符合规范，建议优化语言表达","issues":[]}

待审核内容：
%s`, platform, platform, text)

	raw, err := o.call("你是专业内容审核AI", prompt)
	if err != nil {
		return adapter.SelfReviewResult{}, err
	}

	return parseSelfReview(raw), nil
}

func (o *OllamaRewriter) call(system, user string) (string, error) {
	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	payload := map[string]any{
		"model": o.Model,
		"messages": []Message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		"stream": false,
	}

	b, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", o.BaseURL+"/api/chat", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Ollama 服务调用失败（%s 是否已启动？）: %w", o.BaseURL, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Ollama 返回 %d: %s", resp.StatusCode, string(body))
	}

	var res struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return "", fmt.Errorf("Ollama 响应解析失败: %w", err)
	}
	return res.Message.Content, nil
}

func platformSystemPrompt(platform string) string {
	switch platform {
	case "wechat":
		return "你是一位资深微信公众号编辑，擅长写深度内容，文风严谨有深度，善用小标题和段落分隔，适合长文阅读。"
	case "rednote":
		return "你是小红书博主，文风轻松活泼，善用emoji，标题吸引眼球，内容短小精悍，适合移动端快速浏览。"
	case "csdn":
		return "你是技术博主，擅长写技术教程，文风严谨准确，善用代码块和技术术语，注重实用性和可操作性。"
	case "douyin":
		return "你是抖音图文创作者，内容简短有力，每段不超过2-3句话，善用数字和问句吸引互动。"
	default:
		return "你是专业的内容创作者，文风自然流畅，内容真实有价值。"
	}
}

func parseSelfReview(raw string) adapter.SelfReviewResult {
	// 提取 JSON 部分
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || end <= start {
		return adapter.SelfReviewResult{Passed: false, Comment: "AI自审响应格式异常", Issues: []string{"AI响应格式错误"}}
	}
	jsonStr := raw[start : end+1]

	var data struct {
		Scores  map[string]int `json:"scores"`
		Passed  bool           `json:"passed"`
		Comment string         `json:"comment"`
		Issues  []string       `json:"issues"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return adapter.SelfReviewResult{Passed: false, Comment: raw}
	}

	// 所有维度 >= 8 才算通过
	passed := true
	var issues []string
	for dim, score := range data.Scores {
		if score < 8 {
			passed = false
			issues = append(issues, fmt.Sprintf("%s评分偏低(%d/10)", dim, score))
		}
	}
	if len(data.Issues) > 0 {
		issues = append(issues, data.Issues...)
	}

	return adapter.SelfReviewResult{
		Scores:  data.Scores,
		Passed:  passed && data.Passed,
		Comment: data.Comment,
		Issues:  issues,
	}
}

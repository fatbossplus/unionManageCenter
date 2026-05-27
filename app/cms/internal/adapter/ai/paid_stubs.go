// Package ai 付费 AI 驱动桩（接口已定义，填入 API Key 后可接入）
package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"unionManageCenter/cms/internal/adapter"
)

func init() {
	adapter.RegisterRewriter(&TongyiRewriter{})
	adapter.RegisterRewriter(&OpenAIRewriter{})
}

// TongyiRewriter 通义千问（国内首推，合规优先）
type TongyiRewriter struct {
	APIKey string
	Model  string // qwen-max / qwen-plus / qwen-turbo
}

func (t *TongyiRewriter) Name() string { return "tongyi_paid" }

func (t *TongyiRewriter) Rewrite(req adapter.RewriteRequest) (adapter.RewriteResult, error) {
	if t.APIKey == "" {
		return adapter.RewriteResult{}, errors.New("通义千问：请先配置 api_key")
	}
	model := t.Model
	if model == "" {
		model = "qwen-max"
	}
	sysPrompt := req.SystemPrompt
	if sysPrompt == "" {
		sysPrompt = platformSystemPrompt(req.Platform)
	}

	payload := map[string]any{
		"model": model,
		"input": map[string]any{
			"messages": []map[string]string{
				{"role": "system", "content": sysPrompt},
				{"role": "user", "content": fmt.Sprintf("请改写以下内容（去除作者信息、链接、广告）：\n%s", req.Text)},
			},
		},
	}
	b, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequest("POST", "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation", bytes.NewReader(b))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+t.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return adapter.RewriteResult{}, fmt.Errorf("通义千问请求失败: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var res struct {
		Output struct {
			Text string `json:"text"`
		} `json:"output"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &res); err != nil || res.Output.Text == "" {
		return adapter.RewriteResult{}, fmt.Errorf("通义千问响应解析失败: %s", string(body))
	}

	return adapter.RewriteResult{
		Text:       strings.TrimSpace(res.Output.Text),
		Model:      model,
		TokensUsed: res.Usage.TotalTokens,
	}, nil
}

func (t *TongyiRewriter) SelfReview(text, platform string) (adapter.SelfReviewResult, error) {
	if t.APIKey == "" {
		return adapter.SelfReviewResult{}, errors.New("通义千问：请先配置 api_key")
	}
	// 复用同一逻辑，只是 prompt 不同
	req := adapter.RewriteRequest{
		Text:     text,
		Platform: platform,
		SystemPrompt: "你是内容合规审核专家，只输出JSON格式的审核报告",
	}
	result, err := t.Rewrite(req)
	if err != nil {
		return adapter.SelfReviewResult{}, err
	}
	return parseSelfReview(result.Text), nil
}

// OpenAIRewriter GPT-4o（质量最高，需境外访问）
type OpenAIRewriter struct {
	APIKey  string
	BaseURL string
	Model   string
}

func (o *OpenAIRewriter) Name() string { return "openai_paid" }

func (o *OpenAIRewriter) Rewrite(req adapter.RewriteRequest) (adapter.RewriteResult, error) {
	if o.APIKey == "" {
		return adapter.RewriteResult{}, errors.New("OpenAI：请先配置 api_key")
	}
	baseURL := o.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	model := o.Model
	if model == "" {
		model = "gpt-4o"
	}
	sysPrompt := req.SystemPrompt
	if sysPrompt == "" {
		sysPrompt = platformSystemPrompt(req.Platform)
	}

	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": sysPrompt},
			{"role": "user", "content": req.Text},
		},
	}
	b, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewReader(b))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.APIKey)

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return adapter.RewriteResult{}, fmt.Errorf("OpenAI 请求失败: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var res struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &res); err != nil || len(res.Choices) == 0 {
		return adapter.RewriteResult{}, fmt.Errorf("OpenAI 响应解析失败: %s", string(body))
	}

	return adapter.RewriteResult{
		Text:       strings.TrimSpace(res.Choices[0].Message.Content),
		Model:      model,
		TokensUsed: res.Usage.TotalTokens,
	}, nil
}

func (o *OpenAIRewriter) SelfReview(text, platform string) (adapter.SelfReviewResult, error) {
	if o.APIKey == "" {
		return adapter.SelfReviewResult{}, errors.New("OpenAI：请先配置 api_key")
	}
	req := adapter.RewriteRequest{
		Text:     text,
		Platform: platform,
		SystemPrompt: "你是内容合规审核专家，只输出JSON格式审核报告",
	}
	result, err := o.Rewrite(req)
	if err != nil {
		return adapter.SelfReviewResult{}, err
	}
	return parseSelfReview(result.Text), nil
}

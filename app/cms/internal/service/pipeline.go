// Package service AI 5轮合规处理流水线
//
// 流水线定义：
//  Round 1 - clean       机械清洗：去作者信息、外链、联系方式
//  Round 2 - scan        敏感词扫描：本地词库 + 可选云API
//  Round 3 - rewrite     AI 改写：按目标平台风格重写
//  Round 4 - self_review AI 自审：6维度评分，不合格则重写（最多重试2次）
//  Round 5 - format      格式化：字数/标签/图片规范
package service

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"
	"unionManageCenter/cms/internal/adapter"
	"unionManageCenter/cms/internal/model"
	"unionManageCenter/pkg/database"
)

// PipelineService 5轮流水线编排器
type PipelineService struct {
	db          *gorm.DB
	rewriter    adapter.AIRewriter
	checkers    []adapter.ComplianceChecker
	rewriterName string
}

// NewPipelineService 根据驱动配置初始化流水线
func NewPipelineService(rewriterName string, checkerNames []string) *PipelineService {
	svc := &PipelineService{db: database.Get(), rewriterName: rewriterName}

	if rw, ok := adapter.GetRewriter(rewriterName); ok {
		svc.rewriter = rw
	} else {
		// 降级到本地 wordlist
		svc.rewriter, _ = adapter.GetRewriter("ollama_free")
	}

	for _, name := range checkerNames {
		if ck, ok := adapter.GetChecker(name); ok {
			svc.checkers = append(svc.checkers, ck)
		}
	}
	if len(svc.checkers) == 0 {
		if ck, ok := adapter.GetChecker("wordlist_free"); ok {
			svc.checkers = []adapter.ComplianceChecker{ck}
		}
	}
	return svc
}

// ProcessRaw 对一条原始内容跑完整5轮流水线，返回最终文本
func (p *PipelineService) ProcessRaw(raw *model.RawContent, targetPlatform string) (string, error) {
	// 标记处理中
	p.db.Model(raw).Update("proc_status", model.ProcStatusProcessing)

	text := raw.BodyText
	var finalText string
	var finalErr error

	// 每轮执行，记录结果
	for round := 1; round <= 5; round++ {
		start := time.Now()
		roundName := model.RoundNames[round]
		var outText string
		var roundErr error
		var issues []string

		switch round {
		case model.RoundClean:
			outText = p.roundClean(text)

		case model.RoundScan:
			outText, issues, roundErr = p.roundScan(text, targetPlatform)

		case model.RoundRewrite:
			outText, roundErr = p.roundRewrite(text, targetPlatform)

		case model.RoundSelfReview:
			outText, issues, roundErr = p.roundSelfReview(text, targetPlatform)

		case model.RoundFormat:
			outText = p.roundFormat(text, targetPlatform)
		}

		dur := int(time.Since(start).Milliseconds())
		result := model.ProcResultPass
		if roundErr != nil {
			result = model.ProcResultFail
			issues = append(issues, roundErr.Error())
		}

		// 写入处理记录
		inputHash := hashText(text)
		rec := &model.ProcessRecord{
			RawID:      raw.ID,
			Round:      round,
			RoundName:  roundName,
			InputHash:  inputHash,
			InputText:  text,
			DriverUsed: p.rewriterName,
			Result:     result,
			DurationMs: dur,
		}
		if len(issues) > 0 {
			rec.IssuesJSON = issues
		}
		if outText != "" {
			rec.OutputText = outText
			text = outText // 下一轮使用本轮输出
		}
		p.db.Create(rec)

		if roundErr != nil {
			// 第2轮（scan）失败：中断，内容不合规
			if round == model.RoundScan {
				p.db.Model(raw).Updates(map[string]any{
					"proc_status": model.ProcStatusFailed,
					"proc_error":  roundErr.Error(),
				})
				return "", fmt.Errorf("合规扫描失败：%w", roundErr)
			}
			// 其他轮失败：记录但继续，使用原文
		}
	}

	finalText = text
	finalErr = nil

	p.db.Model(raw).Update("proc_status", model.ProcStatusDone)
	return finalText, finalErr
}

// ─── 各轮实现 ─────────────────────────────────────────────────────────────────

// roundClean 第1轮：机械清洗
func (p *PipelineService) roundClean(text string) string {
	// 1. 去除超链接
	urlRe := regexp.MustCompile(`https?://[^\s\u4e00-\u9fa5，。！？、；：""''（）【】]+`)
	text = urlRe.ReplaceAllString(text, "")

	// 2. 去除微信号/公众号/QQ等联系方式
	contactRe := regexp.MustCompile(`(?i)(微信号?|公众号|wx|qq号?|微博|抖音号?|小红书号?|联系方式)\s*[:：]?\s*[\w@_\-\.]+`)
	text = contactRe.ReplaceAllString(text, "")

	// 3. 去除"关注我"/"点赞收藏"类引流话术
	guideRe := regexp.MustCompile(`(关注|点赞|收藏|转发|分享).{0,10}(我|我们|公众号|账号|主页|主号)`)
	text = guideRe.ReplaceAllString(text, "")

	// 4. 去除"原创作者：xxx"/"文章来源：xxx"
	sourceRe := regexp.MustCompile(`(原创|作者|来源|转自|编辑)[：:]\s*[\S]{1,30}`)
	text = sourceRe.ReplaceAllString(text, "")

	// 5. 合并多余空行
	blankRe := regexp.MustCompile(`\n{3,}`)
	text = blankRe.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

// roundScan 第2轮：合规扫描（使用所有启用的 checker 串联）
func (p *PipelineService) roundScan(text, platform string) (string, []string, error) {
	var allIssues []string
	minScore := 100

	for _, ck := range p.checkers {
		result, err := ck.Check(text, platform)
		if err != nil {
			allIssues = append(allIssues, fmt.Sprintf("[%s] 检查失败: %s", ck.Name(), err.Error()))
			continue
		}
		allIssues = append(allIssues, result.Issues...)
		if result.Score < minScore {
			minScore = result.Score
		}
		if !result.Passed {
			return text, allIssues, fmt.Errorf("合规检查未通过（得分%d/100）：%s", minScore, strings.Join(result.HitWords, "、"))
		}
	}
	return text, allIssues, nil
}

// roundRewrite 第3轮：AI 改写
func (p *PipelineService) roundRewrite(text, platform string) (string, error) {
	if p.rewriter == nil {
		return text, fmt.Errorf("未配置AI改写驱动，跳过改写")
	}
	result, err := p.rewriter.Rewrite(adapter.RewriteRequest{
		Text:     text,
		Platform: platform,
	})
	if err != nil {
		return text, err
	}
	if result.Text == "" {
		return text, fmt.Errorf("AI改写返回空内容")
	}
	return result.Text, nil
}

// roundSelfReview 第4轮：AI 自审（不合格最多重试2次）
func (p *PipelineService) roundSelfReview(text, platform string) (string, []string, error) {
	if p.rewriter == nil {
		return text, nil, nil // 无AI驱动时跳过
	}

	const maxRetry = 2
	for attempt := 0; attempt <= maxRetry; attempt++ {
		review, err := p.rewriter.SelfReview(text, platform)
		if err != nil {
			return text, nil, err
		}
		if review.Passed {
			return text, nil, nil
		}
		if attempt < maxRetry {
			// 带着问题反馈重新改写
			fixPrompt := "请修复以下问题并重新改写：" + strings.Join(review.Issues, "；")
			result, err := p.rewriter.Rewrite(adapter.RewriteRequest{
				Text:         text,
				Platform:     platform,
				SystemPrompt: fixPrompt,
			})
			if err == nil && result.Text != "" {
				text = result.Text
			}
		} else {
			return text, review.Issues, fmt.Errorf("AI自审%d次仍未通过，建议人工审核", maxRetry+1)
		}
	}
	return text, nil, nil
}

// roundFormat 第5轮：格式规范
func (p *PipelineService) roundFormat(text, platform string) string {
	// 按平台限制字数
	maxLen := platformMaxLen(platform)
	if maxLen > 0 && utf8.RuneCountInString(text) > maxLen {
		runes := []rune(text)
		text = string(runes[:maxLen])
		// 在最后一个句号截断
		lastPeriod := strings.LastIndexAny(text, "。！？")
		if lastPeriod > len(text)/2 {
			text = text[:lastPeriod+3] // 3 bytes for 。
		}
	}
	// 去除首尾空白
	text = strings.TrimSpace(text)
	return text
}

func platformMaxLen(platform string) int {
	switch platform {
	case "rednote":
		return 1000
	case "douyin":
		return 500
	case "wechat", "csdn":
		return 0 // 不限
	default:
		return 0
	}
}

func hashText(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:8])
}

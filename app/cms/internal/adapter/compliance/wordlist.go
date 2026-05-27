// Package compliance 本地敏感词库合规检查（免费，零延迟）
package compliance

import (
	"strings"
	"unicode"

	"unionManageCenter/cms/internal/adapter"
)

func init() {
	adapter.RegisterChecker(NewWordlistChecker())
}

// WordlistChecker 基于词典的敏感词检测
type WordlistChecker struct {
	trie *trieNode // AC 自动机（简化版 Trie）
}

func NewWordlistChecker() *WordlistChecker {
	c := &WordlistChecker{trie: newTrie()}
	for _, w := range defaultSensitiveWords {
		c.trie.insert(w)
	}
	return c
}

func (w *WordlistChecker) Name() string { return "wordlist_free" }

func (w *WordlistChecker) Check(text, platform string) (adapter.ComplianceResult, error) {
	// 1. 预处理：去除空白符，统一小写
	clean := normalizeText(text)
	hits := w.trie.search(clean)

	result := adapter.ComplianceResult{
		HitWords: hits,
		Passed:   len(hits) == 0,
	}

	if len(hits) > 5 {
		result.Score = 20
	} else if len(hits) > 2 {
		result.Score = 50
	} else if len(hits) > 0 {
		result.Score = 75
	} else {
		result.Score = 95 // 本地词库只能做基础检测，满分留给云服务
	}

	for _, h := range hits {
		result.Issues = append(result.Issues, "命中敏感词: "+h)
	}

	return result, nil
}

func normalizeText(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !unicode.IsSpace(r) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// ─── Trie 实现 ──────────────────────────────────────────────────────────────

type trieNode struct {
	children map[rune]*trieNode
	end      bool
	word     string
}

func newTrie() *trieNode { return &trieNode{children: map[rune]*trieNode{}} }

func (t *trieNode) insert(word string) {
	node := t
	for _, r := range word {
		if node.children[r] == nil {
			node.children[r] = &trieNode{children: map[rune]*trieNode{}}
		}
		node = node.children[r]
	}
	node.end = true
	node.word = word
}

func (t *trieNode) search(text string) []string {
	runes := []rune(text)
	var hits []string
	seen := map[string]bool{}

	for i := range runes {
		node := t
		for j := i; j < len(runes); j++ {
			c, ok := node.children[runes[j]]
			if !ok {
				break
			}
			node = c
			if node.end && !seen[node.word] {
				hits = append(hits, node.word)
				seen[node.word] = true
			}
		}
	}
	return hits
}

// defaultSensitiveWords 基础敏感词（实际生产建议加载外部词库文件）
// 此处仅列举几十个示例，实际运营环境使用50w+词条词库
var defaultSensitiveWords = []string{
	// 政治相关（举例，非完整）
	"法轮功", "六四", "天安门事件", "独立建国",
	// 违法内容
	"毒品", "大麻", "冰毒", "枪支", "爆炸物",
	"赌博网站", "网络赌博", "地下钱庄",
	// 欺诈
	"百分之百盈利", "稳赚不赔", "内幕消息", "股票内幕",
	"传销", "直销招募", "拉人头",
	// 违规营销
	"加微信", "私信我", "点击链接领取",
	"免费领取", "扫码领红包",
	// 低俗内容
	"约炮", "一夜情", "援交",
}

package worker

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	scanpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/scan"
)

type legacyKeywordRuleSnapshot struct {
	ID            uint64   `json:"id"`
	Name          string   `json:"name"`
	CategoryName  string   `json:"category_name"`
	MatchType     string   `json:"match_type"`
	RiskLevel     string   `json:"risk_level"`
	SuggestAction string   `json:"suggest_action"`
	Scopes        []string `json:"scopes"`
}

// DecodeTaskRules 同时兼容规则快照的新旧结构，避免历史任务因结构演进无法执行。
func DecodeTaskRules(snapshot string) ([]scanpkg.KeywordRule, error) {
	if strings.TrimSpace(snapshot) == "" {
		return nil, errors.New("task rule snapshot is required")
	}

	var rules []scanpkg.KeywordRule
	if err := json.Unmarshal([]byte(snapshot), &rules); err == nil && len(rules) > 0 {
		return rules, nil
	}

	var legacy []legacyKeywordRuleSnapshot
	if err := json.Unmarshal([]byte(snapshot), &legacy); err != nil {
		return nil, err
	}
	rules = make([]scanpkg.KeywordRule, 0, len(legacy))
	for _, item := range legacy {
		rules = append(rules, scanpkg.KeywordRule{
			ID:            item.ID,
			Name:          item.Name,
			Category:      item.CategoryName,
			MatchType:     item.MatchType,
			RiskLevel:     item.RiskLevel,
			SuggestAction: item.SuggestAction,
			Scopes:        append([]string(nil), item.Scopes...),
		})
	}
	return rules, nil
}

// ParseArticleStateFilter 为空或非法时默认只扫在线文稿，符合一期约束。
func ParseArticleStateFilter(value string) int8 {
	value = strings.TrimSpace(value)
	if value == "" {
		return domainpkg.ArticleStateOnline
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return domainpkg.ArticleStateOnline
	}
	return int8(parsed)
}

// ResolveTaskStatus 按整批扫描结果归并状态，便于前端快速判断任务是否可复核。
func ResolveTaskStatus(totalScanned, failCount int64) string {
	switch {
	case totalScanned == 0:
		return domainpkg.TaskStatusSuccess
	case failCount == 0:
		return domainpkg.TaskStatusSuccess
	case failCount >= totalScanned:
		return domainpkg.TaskStatusFailed
	default:
		return domainpkg.TaskStatusPartialSuccess
	}
}

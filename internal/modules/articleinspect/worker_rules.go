package articleinspect

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
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

// decodeTaskRules 同时兼容规则快照的新旧结构，避免历史任务因结构演进无法执行。
func decodeTaskRules(snapshot string) ([]KeywordRule, error) {
	if strings.TrimSpace(snapshot) == "" {
		return nil, errors.New("task rule snapshot is required")
	}

	var rules []KeywordRule
	if err := json.Unmarshal([]byte(snapshot), &rules); err == nil && len(rules) > 0 {
		return rules, nil
	}

	var legacy []legacyKeywordRuleSnapshot
	if err := json.Unmarshal([]byte(snapshot), &legacy); err != nil {
		return nil, err
	}
	rules = make([]KeywordRule, 0, len(legacy))
	for _, item := range legacy {
		rules = append(rules, KeywordRule{
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

// parseArticleStateFilter 为空或非法时默认只扫在线文稿，符合一期约束。
func parseArticleStateFilter(value string) int8 {
	value = strings.TrimSpace(value)
	if value == "" {
		return ArticleStateOnline
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return ArticleStateOnline
	}
	return int8(parsed)
}

// resolveTaskStatus 按整批扫描结果归并状态，便于前端快速判断任务是否可复核。
func resolveTaskStatus(totalScanned, failCount int64) string {
	switch {
	case totalScanned == 0:
		return TaskStatusSuccess
	case failCount == 0:
		return TaskStatusSuccess
	case failCount >= totalScanned:
		return TaskStatusFailed
	default:
		return TaskStatusPartialSuccess
	}
}

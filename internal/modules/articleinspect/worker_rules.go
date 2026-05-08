package articleinspect

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// decodeTaskRules 同时兼容规则快照的新旧结构，避免历史任务因结构演进无法执行。
func decodeTaskRules(snapshot string) ([]KeywordRule, error) {
	if strings.TrimSpace(snapshot) == "" {
		return nil, errors.New("task rule snapshot is required")
	}

	var rules []KeywordRule
	if err := json.Unmarshal([]byte(snapshot), &rules); err == nil && len(rules) > 0 {
		return rules, nil
	}

	var dtos []KeywordDTO
	if err := json.Unmarshal([]byte(snapshot), &dtos); err != nil {
		return nil, err
	}
	rules = make([]KeywordRule, 0, len(dtos))
	for _, dto := range dtos {
		rules = append(rules, KeywordRule{
			ID:            dto.ID,
			Name:          dto.Name,
			Category:      dto.CategoryName,
			MatchType:     dto.MatchType,
			RiskLevel:     dto.RiskLevel,
			SuggestAction: dto.SuggestAction,
			Scopes:        append([]string(nil), dto.Scopes...),
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

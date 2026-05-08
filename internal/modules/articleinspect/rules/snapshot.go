package rules

import (
	"sort"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	scanpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/scan"
)

func BuildKeywordRuleSnapshots(keywords []KeywordRecord, scopesByKeyword map[uint64][]domainpkg.InspectionKeywordScope) []scanpkg.KeywordRule {
	items := make([]scanpkg.KeywordRule, 0, len(keywords))
	for _, keyword := range keywords {
		scopes := scopesByKeyword[keyword.ID]
		scopeValues := make([]string, 0, len(scopes))
		for _, scope := range scopes {
			scopeValues = append(scopeValues, scope.Scope)
		}
		sort.Strings(scopeValues)
		items = append(items, scanpkg.KeywordRule{
			ID:            keyword.ID,
			Name:          keyword.Name,
			Category:      keyword.CategoryName,
			MatchType:     keyword.MatchType,
			RiskLevel:     keyword.RiskLevel,
			SuggestAction: keyword.SuggestAction,
			Scopes:        scopeValues,
		})
	}
	return items
}

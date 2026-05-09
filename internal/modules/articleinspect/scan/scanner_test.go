package scan_test

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	scanpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/scan"
	"github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
)

func TestKeywordScanner(t *testing.T) {
	publishAt := testutil.MustTime(t, "2026-04-20T10:00:00Z")

	scanner := scanpkg.NewKeywordScanner()
	article := scanpkg.CandidateArticle{
		ID:            1,
		OrgID:         100,
		Title:         "Breaking spam headline",
		Keyword:       "alert",
		RichTitle:     "ref-42",
		Body:          "prefix text before spam appears in the body and more text after it for context",
		ShortTitle:    "spam short",
		PublishAtTime: &publishAt,
	}
	rules := []scanpkg.KeywordRule{
		{
			ID:            1,
			Name:          "spam",
			Category:      "policy",
			MatchType:     domainpkg.MatchTypeContains,
			RiskLevel:     domainpkg.RiskLevelHigh,
			SuggestAction: domainpkg.SuggestActionOffline,
			Scopes:        []string{domainpkg.KeywordScopeTitle, domainpkg.KeywordScopeBody},
		},
		{
			ID:            2,
			Name:          "alert",
			Category:      "policy",
			MatchType:     domainpkg.MatchTypeExact,
			RiskLevel:     domainpkg.RiskLevelMedium,
			SuggestAction: domainpkg.SuggestActionProcess,
			Scopes:        []string{domainpkg.KeywordScopeKeyword},
		},
		{
			ID:            3,
			Name:          `ref-\d+`,
			Category:      "policy",
			MatchType:     domainpkg.MatchTypeRegex,
			RiskLevel:     domainpkg.RiskLevelLow,
			SuggestAction: domainpkg.SuggestActionIgnore,
			Scopes:        []string{domainpkg.KeywordScopeRichTitle},
		},
	}

	hits, err := scanner.ScanArticle(context.Background(), article, rules)
	if err != nil {
		t.Fatalf("ScanArticle() error = %v", err)
	}
	if len(hits) != 4 {
		t.Fatalf("ScanArticle() hit count = %d, want %d", len(hits), 4)
	}

	fields := make([]string, 0, len(hits))
	for _, hit := range hits {
		fields = append(fields, hit.FieldName)
		if hit.FieldName == domainpkg.KeywordScopeBody {
			if hit.Snippet == "" || !strings.Contains(strings.ToLower(hit.Snippet), "spam") {
				t.Fatalf("body hit snippet = %q, want contains %q", hit.Snippet, "spam")
			}
		}
	}
	sort.Strings(fields)
	if !reflect.DeepEqual(fields, []string{domainpkg.KeywordScopeBody, domainpkg.KeywordScopeKeyword, domainpkg.KeywordScopeRichTitle, domainpkg.KeywordScopeTitle}) {
		t.Fatalf("ScanArticle() fields = %#v, want %#v", fields, []string{domainpkg.KeywordScopeBody, domainpkg.KeywordScopeKeyword, domainpkg.KeywordScopeRichTitle, domainpkg.KeywordScopeTitle})
	}
}

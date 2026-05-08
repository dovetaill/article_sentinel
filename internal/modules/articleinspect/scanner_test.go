package articleinspect

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestKeywordScanner(t *testing.T) {
	scanner := NewKeywordScanner()
	article := CandidateArticle{
		ID:         1,
		OrgID:      100,
		Title:      "Breaking spam headline",
		Keyword:    "alert",
		RichTitle:  "ref-42",
		Body:       "prefix text before spam appears in the body and more text after it for context",
		ShortTitle: "spam short",
	}
	rules := []KeywordRule{
		{
			ID:            1,
			Name:          "spam",
			Category:      "policy",
			MatchType:     MatchTypeContains,
			RiskLevel:     RiskLevelHigh,
			SuggestAction: SuggestActionOffline,
			Scopes:        []string{KeywordScopeTitle, KeywordScopeBody},
		},
		{
			ID:            2,
			Name:          "alert",
			Category:      "policy",
			MatchType:     MatchTypeExact,
			RiskLevel:     RiskLevelMedium,
			SuggestAction: SuggestActionProcess,
			Scopes:        []string{KeywordScopeKeyword},
		},
		{
			ID:            3,
			Name:          `ref-\d+`,
			Category:      "policy",
			MatchType:     MatchTypeRegex,
			RiskLevel:     RiskLevelLow,
			SuggestAction: SuggestActionIgnore,
			Scopes:        []string{KeywordScopeRichTitle},
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
		if hit.FieldName == KeywordScopeBody {
			if hit.Snippet == "" || !strings.Contains(strings.ToLower(hit.Snippet), "spam") {
				t.Fatalf("body hit snippet = %q, want contains %q", hit.Snippet, "spam")
			}
		}
	}
	sort.Strings(fields)
	if !reflect.DeepEqual(fields, []string{KeywordScopeBody, KeywordScopeKeyword, KeywordScopeRichTitle, KeywordScopeTitle}) {
		t.Fatalf("ScanArticle() fields = %#v, want %#v", fields, []string{KeywordScopeBody, KeywordScopeKeyword, KeywordScopeRichTitle, KeywordScopeTitle})
	}
}

func TestFieldDiff(t *testing.T) {
	before := EditableArticleFields{
		Title: "Old title",
		Body:  "old body content that is deliberately long so the diff summary needs truncation when rendered",
	}
	after := EditableArticleFields{
		Title: "New title",
		Body:  "new body content that is deliberately long so the diff summary needs truncation when rendered differently",
	}

	changes := DiffEditableFields(before, after)
	if len(changes) != 2 {
		t.Fatalf("DiffEditableFields() len = %d, want %d", len(changes), 2)
	}

	var bodyChange *FieldChange
	for index := range changes {
		if changes[index].FieldName == KeywordScopeBody {
			bodyChange = &changes[index]
		}
	}
	if bodyChange == nil {
		t.Fatal("DiffEditableFields() missing body change")
	}
	if len(bodyChange.DiffSummary) == 0 {
		t.Fatal("body diff summary = empty, want non-empty")
	}

	if got := DiffEditableFields(before, before); len(got) != 0 {
		t.Fatalf("DiffEditableFields(no-op) len = %d, want %d", len(got), 0)
	}
}

func TestCandidateArticleLoading(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedCandidateArticles(t, db)
	repo := NewArticleRepository(db)

	start := mustTime(t, "2026-04-20T09:00:00Z")
	end := mustTime(t, "2026-04-20T13:00:00Z")
	firstPage, nextCursor, err := repo.ListCandidateArticles(context.Background(), CandidateArticleFilter{
		OrgID:            100,
		ArticleState:     ArticleStateOnline,
		PublishTimeStart: &start,
		PublishTimeEnd:   &end,
		Limit:            1,
	})
	if err != nil {
		t.Fatalf("ListCandidateArticles(first page) error = %v", err)
	}
	if len(firstPage) != 1 || firstPage[0].ID != 1 {
		t.Fatalf("first page ids = %#v, want first article only", extractArticleIDs(firstPage))
	}
	if firstPage[0].Body != "body one" {
		t.Fatalf("first page body = %q, want %q", firstPage[0].Body, "body one")
	}

	secondPage, _, err := repo.ListCandidateArticles(context.Background(), CandidateArticleFilter{
		OrgID:            100,
		ArticleState:     ArticleStateOnline,
		PublishTimeStart: &start,
		PublishTimeEnd:   &end,
		AfterID:          nextCursor,
		Limit:            1,
	})
	if err != nil {
		t.Fatalf("ListCandidateArticles(second page) error = %v", err)
	}
	if len(secondPage) != 1 || secondPage[0].ID != 2 {
		t.Fatalf("second page ids = %#v, want second article only", extractArticleIDs(secondPage))
	}

	exact, _, err := repo.ListCandidateArticles(context.Background(), CandidateArticleFilter{
		OrgID:        100,
		ArticleState: ArticleStateOnline,
		ArticleID:    2,
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("ListCandidateArticles(exact id) error = %v", err)
	}
	if len(exact) != 1 || exact[0].ID != 2 {
		t.Fatalf("exact filter ids = %#v, want article 2 only", extractArticleIDs(exact))
	}

	fuzzy, _, err := repo.ListCandidateArticles(context.Background(), CandidateArticleFilter{
		OrgID:        100,
		ArticleState: ArticleStateOnline,
		TitleLike:    "Alpha",
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("ListCandidateArticles(title like) error = %v", err)
	}
	if len(fuzzy) != 1 || fuzzy[0].ID != 1 {
		t.Fatalf("title filter ids = %#v, want article 1 only", extractArticleIDs(fuzzy))
	}
}

func extractArticleIDs(items []CandidateArticle) []uint64 {
	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

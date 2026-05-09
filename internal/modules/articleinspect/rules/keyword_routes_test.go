package rules

import (
	"net/http"
	"testing"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	testutil "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
)

func TestKeywordRoutesCRUD(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	handler := newRulesHandler(t, db)

	created := testutil.SendJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/keywords", map[string]any{
		"orgid":          100,
		"name":           "spam",
		"category_id":    1001,
		"match_type":     domainpkg.MatchTypeContains,
		"risk_level":     domainpkg.RiskLevelHigh,
		"suggest_action": domainpkg.SuggestActionOffline,
		"enabled":        true,
		"scopes":         []string{domainpkg.KeywordScopeTitle, domainpkg.KeywordScopeBody},
	})
	if created.Status != http.StatusCreated {
		t.Fatalf("create keyword status = %d, want %d", created.Status, http.StatusCreated)
	}
	createdData := testutil.DataMap(t, created.Envelope.Data)
	keywordID := testutil.Uint64Field(t, createdData, "id")

	listed := testutil.SendRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/keywords?orgid=100&page=1&page_size=20", nil)
	if listed.Status != http.StatusOK {
		t.Fatalf("list keywords status = %d, want %d", listed.Status, http.StatusOK)
	}
	listData := testutil.DataMap(t, listed.Envelope.Data)
	if total := testutil.NumberField(t, listData, "total"); total != 1 {
		t.Fatalf("list keywords total = %v, want %d", total, 1)
	}

	detail := testutil.SendRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/keywords/"+testutil.Uint64String(t, createdData["id"])+"?orgid=100", nil)
	if detail.Status != http.StatusOK {
		t.Fatalf("get keyword status = %d, want %d", detail.Status, http.StatusOK)
	}

	updated := testutil.SendJSONRequest(t, handler, http.MethodPut, "/api/v1/article-inspect/keywords/"+testutil.Uint64String(t, createdData["id"]), map[string]any{
		"orgid":          100,
		"name":           "spam-updated",
		"category_id":    1002,
		"match_type":     domainpkg.MatchTypeContains,
		"risk_level":     domainpkg.RiskLevelHigh,
		"suggest_action": domainpkg.SuggestActionProcess,
		"enabled":        true,
		"remark":         "review immediately",
		"scopes":         []string{domainpkg.KeywordScopeTitle},
	})
	if updated.Status != http.StatusOK {
		t.Fatalf("update keyword status = %d, want %d", updated.Status, http.StatusOK)
	}

	patched := testutil.SendJSONRequest(t, handler, http.MethodPatch, "/api/v1/article-inspect/keywords/"+testutil.Uint64String(t, createdData["id"])+"/status", map[string]any{
		"orgid":   100,
		"enabled": false,
	})
	if patched.Status != http.StatusOK {
		t.Fatalf("patch keyword status = %d, want %d", patched.Status, http.StatusOK)
	}

	deleted := testutil.SendRequest(t, handler, http.MethodDelete, "/api/v1/article-inspect/keywords/"+testutil.Uint64String(t, createdData["id"])+"?orgid=100", nil)
	if deleted.Status != http.StatusOK {
		t.Fatalf("delete keyword status = %d, want %d", deleted.Status, http.StatusOK)
	}
	if keywordID == 0 {
		t.Fatal("keyword id = 0, want persisted id")
	}
}

func TestKeywordRoutesUseCategoryIDFields(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	handler := newRulesHandler(t, db)

	created := testutil.SendJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/keywords", map[string]any{
		"orgid":          29,
		"name":           "敏感词",
		"category_id":    501,
		"match_type":     domainpkg.MatchTypeContains,
		"risk_level":     domainpkg.RiskLevelHigh,
		"suggest_action": domainpkg.SuggestActionOffline,
		"enabled":        true,
		"scopes":         []string{domainpkg.KeywordScopeTitle},
	})
	if created.Status != http.StatusCreated {
		t.Fatalf("create keyword status = %d, want %d", created.Status, http.StatusCreated)
	}

	createdData := testutil.DataMap(t, created.Envelope.Data)
	if testutil.Uint64Field(t, createdData, "category_id") != 501 {
		t.Fatalf("create keyword category_id = %d, want %d", testutil.Uint64Field(t, createdData, "category_id"), 501)
	}
	if testutil.StringField(t, createdData, "category_name") != "政策红线" {
		t.Fatalf("create keyword category_name = %q, want %q", testutil.StringField(t, createdData, "category_name"), "政策红线")
	}

	updated := testutil.SendJSONRequest(t, handler, http.MethodPut, "/api/v1/article-inspect/keywords/"+testutil.Uint64String(t, createdData["id"]), map[string]any{
		"orgid":          29,
		"name":           "敏感词-更新",
		"category_id":    502,
		"match_type":     domainpkg.MatchTypeContains,
		"risk_level":     domainpkg.RiskLevelMedium,
		"suggest_action": domainpkg.SuggestActionProcess,
		"enabled":        true,
		"scopes":         []string{domainpkg.KeywordScopeBody},
	})
	if updated.Status != http.StatusOK {
		t.Fatalf("update keyword status = %d, want %d", updated.Status, http.StatusOK)
	}

	updatedData := testutil.DataMap(t, updated.Envelope.Data)
	if testutil.Uint64Field(t, updatedData, "category_id") != 502 {
		t.Fatalf("update keyword category_id = %d, want %d", testutil.Uint64Field(t, updatedData, "category_id"), 502)
	}
	if testutil.StringField(t, updatedData, "category_name") != "高频违规" {
		t.Fatalf("update keyword category_name = %q, want %q", testutil.StringField(t, updatedData, "category_name"), "高频违规")
	}
}

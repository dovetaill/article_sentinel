package results

import (
	"context"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	testutil "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
	"gorm.io/gorm"
)

func newResultsHandler(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	group := huma.NewGroup(api, "/api/v1/article-inspect")
	group.UseSimpleModifier(func(op *huma.Operation) {
		op.SkipValidateParams = true
	})
	RegisterResultRoutes(group, NewResultService(db))
	return mux
}

func TestResultQuery(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedQueryFixtures(t, db)
	service := NewResultService(db)

	listed, err := service.List(context.Background(), ResultListInput{
		OrgID:             100,
		TaskID:            501,
		RiskLevel:         domainpkg.RiskLevelHigh,
		DispositionStatus: domainpkg.ResultDispositionPending,
		TitleLike:         "Alpha",
		Page:              1,
		PageSize:          20,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if listed.Total != 1 || len(listed.Items) != 1 || listed.Items[0].ID != 1001 {
		t.Fatalf("List() = %+v, want one matching result 1001", listed)
	}
	if listed.Items[0].PreviewFieldName != domainpkg.KeywordScopeTitle || listed.Items[0].PreviewKeywordText != "alpha" {
		t.Fatalf("List() preview = %+v, want first title hit metadata", listed.Items[0])
	}
	if listed.Items[0].PreviewMatchedText != "Alpha" || listed.Items[0].PreviewSnippet != "Alpha news" {
		t.Fatalf("List() preview text = %+v, want title snippet preview", listed.Items[0])
	}
	if listed.Items[0].ExtraHitCount != 1 {
		t.Fatalf("List() extra hit count = %d, want %d", listed.Items[0].ExtraHitCount, 1)
	}

	byArticleID, err := service.List(context.Background(), ResultListInput{OrgID: 100, ArticleID: 2, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("List(article id) error = %v", err)
	}
	if byArticleID.Total != 1 || len(byArticleID.Items) != 1 || byArticleID.Items[0].ArticleID != 2 {
		t.Fatalf("List(article id) = %+v, want article 2 only", byArticleID)
	}

	detail, err := service.GetDetail(context.Background(), 100, 1001)
	if err != nil {
		t.Fatalf("GetDetail() error = %v", err)
	}
	if detail.Result.ID != 1001 || len(detail.Hits) != 2 || len(detail.OperationLogs) != 2 {
		t.Fatalf("GetDetail() = %+v, want result 1001 with 2 hits and 2 operation logs", detail)
	}
}

func TestResultRoutesListAndDetail(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedQueryFixtures(t, db)
	handler := newResultsHandler(t, db)

	listed := testutil.SendRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/results?orgid=100&task_id=501&risk_level=high&page=1&page_size=20", nil)
	if listed.Status != http.StatusOK {
		t.Fatalf("list results status = %d, want %d", listed.Status, http.StatusOK)
	}
	resultListData := testutil.DataMap(t, listed.Envelope.Data)
	if total := testutil.NumberField(t, resultListData, "total"); total != 1 {
		t.Fatalf("list results total = %v, want %d", total, 1)
	}
	resultItems := testutil.ListField(t, resultListData, "items")
	if len(resultItems) != 1 {
		t.Fatalf("result list items len = %d, want %d", len(resultItems), 1)
	}
	firstResult := testutil.DataMap(t, resultItems[0])
	if testutil.StringField(t, firstResult, "preview_field_name") != domainpkg.KeywordScopeTitle {
		t.Fatalf("preview_field_name = %q, want %q", testutil.StringField(t, firstResult, "preview_field_name"), domainpkg.KeywordScopeTitle)
	}
	if testutil.StringField(t, firstResult, "preview_keyword_text") != "alpha" {
		t.Fatalf("preview_keyword_text = %q, want %q", testutil.StringField(t, firstResult, "preview_keyword_text"), "alpha")
	}
	if testutil.StringField(t, firstResult, "preview_matched_text") != "Alpha" {
		t.Fatalf("preview_matched_text = %q, want %q", testutil.StringField(t, firstResult, "preview_matched_text"), "Alpha")
	}
	if testutil.StringField(t, firstResult, "preview_snippet") != "Alpha news" {
		t.Fatalf("preview_snippet = %q, want %q", testutil.StringField(t, firstResult, "preview_snippet"), "Alpha news")
	}
	if testutil.Uint64Field(t, firstResult, "extra_hit_count") != 1 {
		t.Fatalf("extra_hit_count = %d, want %d", testutil.Uint64Field(t, firstResult, "extra_hit_count"), 1)
	}

	detail := testutil.SendRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/results/1001?orgid=100", nil)
	if detail.Status != http.StatusOK {
		t.Fatalf("get result detail status = %d, want %d", detail.Status, http.StatusOK)
	}
	detailData := testutil.DataMap(t, detail.Envelope.Data)
	if _, ok := detailData["hits"].([]any); !ok {
		t.Fatalf("detail hits type = %T, want []any", detailData["hits"])
	}
	if _, ok := detailData["operation_logs"].([]any); !ok {
		t.Fatalf("detail operation_logs type = %T, want []any", detailData["operation_logs"])
	}
	if _, ok := detailData["field_change_logs"].([]any); !ok {
		t.Fatalf("detail field_change_logs type = %T, want []any", detailData["field_change_logs"])
	}
	if _, ok := detailData["field_changes"]; ok {
		t.Fatalf("detail keys = %#v, do not want field_changes alias", detailData)
	}
}

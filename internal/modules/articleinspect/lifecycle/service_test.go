package lifecycle

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/article-sentinel/internal/identity"
	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	testutil "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
	"gorm.io/gorm"
)

func newLifecycleHandler(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	group := huma.NewGroup(api, "/api/v1/article-inspect")
	group.UseSimpleModifier(func(op *huma.Operation) {
		op.SkipValidateParams = true
	})
	RegisterLifecycleRoutes(group, NewLifecycleService(db))
	return mux
}

func TestLifecycle(t *testing.T) {
	t.Run("offline transitions 9 to 8", func(t *testing.T) {
		db := testutil.NewArticleInspectTestDB(t)
		testutil.SeedLifecycleArticles(t, db)
		service := NewLifecycleService(db)

		result, err := service.OfflineArticle(context.Background(), OfflineArticleInput{OrgID: 100, ArticleID: 10, OperatorID: 7})
		if err != nil {
			t.Fatalf("OfflineArticle() error = %v", err)
		}
		if result.Status != domainpkg.ActionStatusSuccess || result.AfterState != domainpkg.ArticleStateOffline {
			t.Fatalf("OfflineArticle() = %+v, want success to state %d", result, domainpkg.ArticleStateOffline)
		}

		var article domainpkg.Article
		if err := db.First(&article, 10).Error; err != nil {
			t.Fatalf("load article error = %v", err)
		}
		if article.State != domainpkg.ArticleStateOffline {
			t.Fatalf("article.State = %d, want %d", article.State, domainpkg.ArticleStateOffline)
		}
	})

	t.Run("already offline records as skipped", func(t *testing.T) {
		db := testutil.NewArticleInspectTestDB(t)
		testutil.SeedLifecycleArticles(t, db)
		service := NewLifecycleService(db)

		result, err := service.OfflineArticle(context.Background(), OfflineArticleInput{OrgID: 100, ArticleID: 11, OperatorID: 7})
		if err != nil {
			t.Fatalf("OfflineArticle() error = %v", err)
		}
		if result.Status != domainpkg.ActionStatusSkipped || result.AfterState != domainpkg.ArticleStateOffline {
			t.Fatalf("OfflineArticle() = %+v, want skipped at state %d", result, domainpkg.ArticleStateOffline)
		}
	})

	t.Run("rectify updates article fields and writes change logs", func(t *testing.T) {
		db := testutil.NewArticleInspectTestDB(t)
		testutil.SeedLifecycleArticles(t, db)
		service := NewLifecycleService(db)

		changes, err := service.UpdateArticleFields(context.Background(), UpdateArticleFieldsInput{
			OrgID:      100,
			ArticleID:  12,
			OperatorID: 7,
			Fields:     EditableArticleFields{Title: "Rectified title", Body: "updated body for review"},
		})
		if err != nil {
			t.Fatalf("UpdateArticleFields() error = %v", err)
		}
		if len(changes) != 2 {
			t.Fatalf("UpdateArticleFields() change count = %d, want %d", len(changes), 2)
		}

		var logs []domainpkg.InspectionFieldChangeLog
		if err := db.Where("orgid = ? AND article_id = ?", 100, 12).Order("field_name ASC").Find(&logs).Error; err != nil {
			t.Fatalf("load change logs error = %v", err)
		}
		if len(logs) != 2 {
			t.Fatalf("field change logs len = %d, want %d", len(logs), 2)
		}
	})

	t.Run("republish defaults from 8 to 1 unless configured otherwise", func(t *testing.T) {
		db := testutil.NewArticleInspectTestDB(t)
		testutil.SeedLifecycleArticles(t, db)
		service := NewLifecycleService(db)

		result, err := service.RepublishArticle(context.Background(), RepublishArticleInput{OrgID: 100, ArticleID: 11, OperatorID: 7})
		if err != nil {
			t.Fatalf("RepublishArticle() error = %v", err)
		}
		if result.AfterState != domainpkg.ArticleStateAuditPending {
			t.Fatalf("RepublishArticle().AfterState = %d, want %d", result.AfterState, domainpkg.ArticleStateAuditPending)
		}
	})
}

func TestLifecycleRoutesUseSnakeCaseResponses(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedLifecycleArticles(t, db)
	handler := newLifecycleHandler(t, db)

	rectify := testutil.SendJSONRequest(t, handler, http.MethodPut, "/api/v1/article-inspect/articles/12/rectify", map[string]any{
		"orgid": 100,
		"title": "Updated title",
		"desc":  "Updated desc",
		"body":  "Updated body content",
	})
	if rectify.Status != http.StatusOK {
		t.Fatalf("rectify status = %d, want %d", rectify.Status, http.StatusOK)
	}

	rectifyChanges, ok := rectify.Envelope.Data.([]any)
	if !ok || len(rectifyChanges) == 0 {
		t.Fatalf("rectify data = %#v, want non-empty []any", rectify.Envelope.Data)
	}
	firstChange := testutil.DataMap(t, rectifyChanges[0])
	if testutil.StringField(t, firstChange, "field_name") == "" {
		t.Fatalf("rectify field_name = empty, want non-empty")
	}
	if _, ok := firstChange["FieldName"]; ok {
		t.Fatalf("rectify change keys = %#v, do not want FieldName", firstChange)
	}

	republish := testutil.SendJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/articles/11/republish", map[string]any{
		"orgid":  100,
		"reason": "send back to audit",
	})
	if republish.Status != http.StatusOK {
		t.Fatalf("republish status = %d, want %d", republish.Status, http.StatusOK)
	}

	republishData := testutil.DataMap(t, republish.Envelope.Data)
	if testutil.Uint64Field(t, republishData, "article_id") != 11 {
		t.Fatalf("republish article_id = %d, want %d", testutil.Uint64Field(t, republishData, "article_id"), 11)
	}
	if _, ok := republishData["ArticleID"]; ok {
		t.Fatalf("republish keys = %#v, do not want ArticleID", republishData)
	}
}

func TestOperatorResolverPreservesAuditMetadataOnLogs(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedLifecycleArticles(t, db)
	service := NewLifecycleService(db)

	actor := identity.NewActor(23, "jwt-user", "reviewer", "active")
	ctx := identity.ContextWithActor(context.Background(), actor)
	ctx = identity.ContextWithRequestMetadata(ctx, identity.RequestMetadata{RequestID: "req-456", SourceIP: "198.51.100.25"})

	changes, err := service.UpdateArticleFields(ctx, UpdateArticleFieldsInput{
		OrgID:        100,
		ArticleID:    12,
		OperatorID:   uint64(actor.ID),
		OperatorName: actor.Username,
		Fields:       EditableArticleFields{Title: "Updated title", Body: "Updated body content"},
	})
	if err != nil {
		t.Fatalf("UpdateArticleFields() error = %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("UpdateArticleFields() change count = %d, want %d", len(changes), 2)
	}

	var opLogs []domainpkg.InspectionOperationLog
	if err := db.Where("orgid = ? AND article_id = ?", 100, 12).Find(&opLogs).Error; err != nil {
		t.Fatalf("load operation logs error = %v", err)
	}
	if len(opLogs) != 1 {
		t.Fatalf("operation logs len = %d, want %d", len(opLogs), 1)
	}
	if opLogs[0].RequestID != "req-456" || opLogs[0].SourceIP != "198.51.100.25" {
		t.Fatalf("operation log audit metadata = %+v, want request id and source ip", opLogs[0])
	}
	if opLogs[0].Summary == "" {
		t.Fatal("operation log Summary = empty, want non-empty")
	}
	if opLogs[0].RequestSnapshot == "" || !strings.Contains(opLogs[0].RequestSnapshot, "\"article_id\":12") {
		t.Fatalf("operation log RequestSnapshot = %q, want article identifier", opLogs[0].RequestSnapshot)
	}

	var changeLogs []domainpkg.InspectionFieldChangeLog
	if err := db.Where("orgid = ? AND article_id = ?", 100, 12).Find(&changeLogs).Error; err != nil {
		t.Fatalf("load field change logs error = %v", err)
	}
	if len(changeLogs) != 2 {
		t.Fatalf("field change logs len = %d, want %d", len(changeLogs), 2)
	}
	for _, log := range changeLogs {
		if log.RequestID != "req-456" || log.SourceIP != "198.51.100.25" {
			t.Fatalf("field change log audit metadata = %+v, want request id and source ip", log)
		}
	}
}

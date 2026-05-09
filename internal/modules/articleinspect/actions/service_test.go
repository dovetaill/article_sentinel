package actions

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

func newActionsHandler(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	group := huma.NewGroup(api, "/api/v1/article-inspect")
	group.UseSimpleModifier(func(op *huma.Operation) {
		op.SkipValidateParams = true
	})
	RegisterActionRoutes(group, NewActionService(db, NewActionRepository(db), nil))
	return mux
}

func TestBatchAction(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedActionFixtures(t, db)
	service := NewActionService(db, NewActionRepository(db), nil)

	ignore, err := service.BatchIgnore(context.Background(), BatchActionInput{OrgID: 100, TaskID: 501, ResultIDs: []uint64{1001, 1002}, OperatorID: 7, Reason: "ignore duplicates"})
	if err != nil {
		t.Fatalf("BatchIgnore() error = %v", err)
	}
	if ignore.SuccessCount != 1 || ignore.SkipCount != 1 {
		t.Fatalf("BatchIgnore() summary = %+v, want success 1 skip 1", ignore)
	}

	processed, err := service.BatchProcess(context.Background(), BatchActionInput{OrgID: 100, TaskID: 501, ResultIDs: []uint64{1003, 1004}, OperatorID: 7, Reason: "mark done"})
	if err != nil {
		t.Fatalf("BatchProcess() error = %v", err)
	}
	if processed.SuccessCount != 1 || processed.SkipCount != 1 {
		t.Fatalf("BatchProcess() summary = %+v, want success 1 skip 1", processed)
	}

	t.Run("batch logs inherit task relation and audit snapshot from result rows", func(t *testing.T) {
		db := testutil.NewArticleInspectTestDB(t)
		testutil.SeedActionFixtures(t, db)
		service := NewActionService(db, NewActionRepository(db), nil)

		ctx := identity.ContextWithActor(context.Background(), identity.NewActor(17, "auditor", "reviewer", "active"))
		ctx = identity.ContextWithRequestMetadata(ctx, identity.RequestMetadata{RequestID: "req-batch-1", SourceIP: "203.0.113.20"})

		if _, err := service.BatchIgnore(ctx, BatchActionInput{OrgID: 100, ResultIDs: []uint64{1001}, Reason: "ignore duplicates"}); err != nil {
			t.Fatalf("BatchIgnore() error = %v", err)
		}

		var log domainpkg.InspectionOperationLog
		if err := db.Where("orgid = ? AND result_id = ? AND operation_type = ?", 100, 1001, domainpkg.ActionTypeBatchIgnore).Order("id DESC").First(&log).Error; err != nil {
			t.Fatalf("load batch operation log error = %v", err)
		}
		if log.TaskID != 501 {
			t.Fatalf("operation log TaskID = %d, want %d", log.TaskID, 501)
		}
		if log.OperatorName != "auditor" {
			t.Fatalf("operation log OperatorName = %q, want %q", log.OperatorName, "auditor")
		}
		if log.Summary == "" {
			t.Fatal("operation log Summary = empty, want non-empty")
		}
		if log.RequestSnapshot == "" || !strings.Contains(log.RequestSnapshot, "\"task_id\":501") || !strings.Contains(log.RequestSnapshot, "\"result_id\":1001") {
			t.Fatalf("operation log RequestSnapshot = %q, want task/result identifiers", log.RequestSnapshot)
		}
		if log.RequestID != "req-batch-1" || log.SourceIP != "203.0.113.20" {
			t.Fatalf("operation log audit metadata = %+v, want request id and source ip", log)
		}
	})

	t.Run("offline updates article and result state", func(t *testing.T) {
		db := testutil.NewArticleInspectTestDB(t)
		testutil.SeedLifecycleArticles(t, db)
		testutil.SeedBatchOfflineFixtures(t, db)
		service := NewActionService(db, NewActionRepository(db), nil)

		offline, err := service.BatchOffline(context.Background(), BatchActionInput{OrgID: 100, TaskID: 501, ResultIDs: []uint64{2001, 2002}, OperatorID: 7, Reason: "manual batch offline"})
		if err != nil {
			t.Fatalf("BatchOffline() error = %v", err)
		}
		if offline.SuccessCount != 1 || offline.SkipCount != 1 {
			t.Fatalf("BatchOffline() summary = %+v, want success 1 skip 1", offline)
		}

		var article domainpkg.Article
		if err := db.First(&article, 10).Error; err != nil {
			t.Fatalf("load article error = %v", err)
		}
		if article.State != domainpkg.ArticleStateOffline {
			t.Fatalf("article.State = %d, want %d", article.State, domainpkg.ArticleStateOffline)
		}

		var result domainpkg.InspectionResult
		if err := db.First(&result, 2001).Error; err != nil {
			t.Fatalf("load result error = %v", err)
		}
		if result.DispositionStatus != domainpkg.ResultDispositionOfflined {
			t.Fatalf("result.DispositionStatus = %q, want %q", result.DispositionStatus, domainpkg.ResultDispositionOfflined)
		}
		if result.ArticleState != domainpkg.ArticleStateOffline {
			t.Fatalf("result.ArticleState = %d, want %d", result.ArticleState, domainpkg.ArticleStateOffline)
		}
	})
}

func TestActionRoutesValidateTargets(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedActionFixtures(t, db)
	handler := newActionsHandler(t, db)

	tests := []struct {
		name string
		path string
	}{
		{name: "batch offline", path: "/api/v1/article-inspect/actions/batch-offline"},
		{name: "batch ignore", path: "/api/v1/article-inspect/actions/batch-ignore"},
		{name: "batch process", path: "/api/v1/article-inspect/actions/batch-process"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testutil.SendJSONRequest(t, handler, http.MethodPost, tt.path, map[string]any{"orgid": 100, "task_id": 501, "reason": "missing targets"})
			if result.Status != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", result.Status, http.StatusBadRequest)
			}
			if result.Envelope.Code != http.StatusBadRequest {
				t.Fatalf("envelope = %+v, want bad request code", result.Envelope)
			}
		})
	}
}

func TestActionRoutesUseSnakeCaseResponse(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedLifecycleArticles(t, db)
	testutil.SeedBatchOfflineFixtures(t, db)
	handler := newActionsHandler(t, db)

	batchOffline := testutil.SendJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/actions/batch-offline", map[string]any{
		"orgid":      100,
		"task_id":    501,
		"result_ids": []uint64{2001},
		"reason":     "manual batch offline",
	})
	if batchOffline.Status != http.StatusOK {
		t.Fatalf("batch offline status = %d, want %d", batchOffline.Status, http.StatusOK)
	}

	batchData := testutil.DataMap(t, batchOffline.Envelope.Data)
	if testutil.Uint64Field(t, batchData, "action_id") == 0 {
		t.Fatalf("batch action_id = 0, want non-zero")
	}
	if testutil.Uint64Field(t, batchData, "target_count") != 1 {
		t.Fatalf("batch target_count = %d, want %d", testutil.Uint64Field(t, batchData, "target_count"), 1)
	}
	if _, ok := batchData["ActionID"]; ok {
		t.Fatalf("batch action keys = %#v, do not want ActionID", batchData)
	}
}

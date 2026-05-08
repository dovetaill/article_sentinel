package articleinspect

import (
	"context"
	"strings"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/identity"
)

func TestLifecycle(t *testing.T) {
	t.Run("offline transitions 9 to 8", func(t *testing.T) {
		db := newArticleInspectTestDB(t)
		seedLifecycleArticles(t, db)
		service := NewLifecycleService(db)

		result, err := service.OfflineArticle(context.Background(), OfflineArticleInput{
			OrgID:      100,
			ArticleID:  10,
			OperatorID: 7,
		})
		if err != nil {
			t.Fatalf("OfflineArticle() error = %v", err)
		}
		if result.Status != ActionStatusSuccess || result.AfterState != ArticleStateOffline {
			t.Fatalf("OfflineArticle() = %+v, want success to state %d", result, ArticleStateOffline)
		}

		var article Article
		if err := db.First(&article, 10).Error; err != nil {
			t.Fatalf("load article error = %v", err)
		}
		if article.State != ArticleStateOffline {
			t.Fatalf("article.State = %d, want %d", article.State, ArticleStateOffline)
		}
	})

	t.Run("already offline records as skipped", func(t *testing.T) {
		db := newArticleInspectTestDB(t)
		seedLifecycleArticles(t, db)
		service := NewLifecycleService(db)

		result, err := service.OfflineArticle(context.Background(), OfflineArticleInput{
			OrgID:      100,
			ArticleID:  11,
			OperatorID: 7,
		})
		if err != nil {
			t.Fatalf("OfflineArticle() error = %v", err)
		}
		if result.Status != ActionStatusSkipped || result.AfterState != ArticleStateOffline {
			t.Fatalf("OfflineArticle() = %+v, want skipped at state %d", result, ArticleStateOffline)
		}
	})

	t.Run("rectify updates article fields and writes change logs", func(t *testing.T) {
		db := newArticleInspectTestDB(t)
		seedLifecycleArticles(t, db)
		service := NewLifecycleService(db)

		changes, err := service.UpdateArticleFields(context.Background(), UpdateArticleFieldsInput{
			OrgID:      100,
			ArticleID:  12,
			OperatorID: 7,
			Fields: EditableArticleFields{
				Title: "Rectified title",
				Body:  "updated body for review",
			},
		})
		if err != nil {
			t.Fatalf("UpdateArticleFields() error = %v", err)
		}
		if len(changes) != 2 {
			t.Fatalf("UpdateArticleFields() change count = %d, want %d", len(changes), 2)
		}

		var logs []InspectionFieldChangeLog
		if err := db.Where("orgid = ? AND article_id = ?", 100, 12).Order("field_name ASC").Find(&logs).Error; err != nil {
			t.Fatalf("load change logs error = %v", err)
		}
		if len(logs) != 2 {
			t.Fatalf("field change logs len = %d, want %d", len(logs), 2)
		}
	})

	t.Run("republish defaults from 8 to 1 unless configured otherwise", func(t *testing.T) {
		db := newArticleInspectTestDB(t)
		seedLifecycleArticles(t, db)
		service := NewLifecycleService(db)

		result, err := service.RepublishArticle(context.Background(), RepublishArticleInput{
			OrgID:      100,
			ArticleID:  11,
			OperatorID: 7,
		})
		if err != nil {
			t.Fatalf("RepublishArticle() error = %v", err)
		}
		if result.AfterState != ArticleStateAuditPending {
			t.Fatalf("RepublishArticle().AfterState = %d, want %d", result.AfterState, ArticleStateAuditPending)
		}
	})
}

func TestBatchAction(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedActionFixtures(t, db)
	service := NewActionService(db, NewActionRepository(db))

	ignore, err := service.BatchIgnore(context.Background(), BatchActionInput{
		OrgID:      100,
		TaskID:     501,
		ResultIDs:  []uint64{1001, 1002},
		OperatorID: 7,
		Reason:     "ignore duplicates",
	})
	if err != nil {
		t.Fatalf("BatchIgnore() error = %v", err)
	}
	if ignore.SuccessCount != 1 || ignore.SkipCount != 1 {
		t.Fatalf("BatchIgnore() summary = %+v, want success 1 skip 1", ignore)
	}

	processed, err := service.BatchProcess(context.Background(), BatchActionInput{
		OrgID:      100,
		TaskID:     501,
		ResultIDs:  []uint64{1003, 1004},
		OperatorID: 7,
		Reason:     "mark done",
	})
	if err != nil {
		t.Fatalf("BatchProcess() error = %v", err)
	}
	if processed.SuccessCount != 1 || processed.SkipCount != 1 {
		t.Fatalf("BatchProcess() summary = %+v, want success 1 skip 1", processed)
	}

	t.Run("batch logs inherit task relation and audit snapshot from result rows", func(t *testing.T) {
		db := newArticleInspectTestDB(t)
		seedActionFixtures(t, db)
		service := NewActionService(db, NewActionRepository(db))

		ctx := identity.ContextWithActor(context.Background(), identity.NewActor(17, "auditor", "reviewer", "active"))
		ctx = identity.ContextWithRequestMetadata(ctx, identity.RequestMetadata{
			RequestID: "req-batch-1",
			SourceIP:  "203.0.113.20",
		})

		if _, err := service.BatchIgnore(ctx, BatchActionInput{
			OrgID:     100,
			ResultIDs: []uint64{1001},
			Reason:    "ignore duplicates",
		}); err != nil {
			t.Fatalf("BatchIgnore() error = %v", err)
		}

		var log InspectionOperationLog
		if err := db.Where("orgid = ? AND result_id = ? AND operation_type = ?", 100, 1001, ActionTypeBatchIgnore).
			Order("id DESC").
			First(&log).Error; err != nil {
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
		db := newArticleInspectTestDB(t)
		seedLifecycleArticles(t, db)
		seedBatchOfflineFixtures(t, db)
		service := NewActionService(db, NewActionRepository(db))

		offline, err := service.BatchOffline(context.Background(), BatchActionInput{
			OrgID:      100,
			TaskID:     501,
			ResultIDs:  []uint64{2001, 2002},
			OperatorID: 7,
			Reason:     "manual batch offline",
		})
		if err != nil {
			t.Fatalf("BatchOffline() error = %v", err)
		}
		if offline.SuccessCount != 1 || offline.SkipCount != 1 {
			t.Fatalf("BatchOffline() summary = %+v, want success 1 skip 1", offline)
		}

		var article Article
		if err := db.First(&article, 10).Error; err != nil {
			t.Fatalf("load article error = %v", err)
		}
		if article.State != ArticleStateOffline {
			t.Fatalf("article.State = %d, want %d", article.State, ArticleStateOffline)
		}

		var result InspectionResult
		if err := db.First(&result, 2001).Error; err != nil {
			t.Fatalf("load result error = %v", err)
		}
		if result.DispositionStatus != ResultDispositionOfflined {
			t.Fatalf("result.DispositionStatus = %q, want %q", result.DispositionStatus, ResultDispositionOfflined)
		}
		if result.ArticleState != ArticleStateOffline {
			t.Fatalf("result.ArticleState = %d, want %d", result.ArticleState, ArticleStateOffline)
		}
	})
}

func TestResultQuery(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedQueryFixtures(t, db)
	service := NewResultService(db)

	listed, err := service.List(context.Background(), ResultListInput{
		OrgID:             100,
		TaskID:            501,
		RiskLevel:         RiskLevelHigh,
		DispositionStatus: ResultDispositionPending,
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
	if listed.Items[0].PreviewFieldName != KeywordScopeTitle || listed.Items[0].PreviewKeywordText != "alpha" {
		t.Fatalf("List() preview = %+v, want first title hit metadata", listed.Items[0])
	}
	if listed.Items[0].PreviewMatchedText != "Alpha" || listed.Items[0].PreviewSnippet != "Alpha news" {
		t.Fatalf("List() preview text = %+v, want title snippet preview", listed.Items[0])
	}
	if listed.Items[0].ExtraHitCount != 1 {
		t.Fatalf("List() extra hit count = %d, want %d", listed.Items[0].ExtraHitCount, 1)
	}

	byArticleID, err := service.List(context.Background(), ResultListInput{
		OrgID:     100,
		ArticleID: 2,
		Page:      1,
		PageSize:  20,
	})
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

func TestOperationLogQuery(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedQueryFixtures(t, db)
	service := NewLogService(db)

	start := mustTime(t, "2026-04-20T09:30:00Z")
	end := mustTime(t, "2026-04-20T12:30:00Z")
	result, err := service.ListOperationLogs(context.Background(), OperationLogListInput{
		OrgID:        100,
		ArticleID:    1,
		TaskID:       501,
		OperatorName: "alice",
		StartAt:      &start,
		EndAt:        &end,
		Page:         1,
		PageSize:     20,
	})
	if err != nil {
		t.Fatalf("ListOperationLogs() error = %v", err)
	}
	if result.Total != 2 || len(result.Items) != 2 {
		t.Fatalf("ListOperationLogs() = %+v, want 2 matching logs", result)
	}
}

func TestFieldChangeLogQuery(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedQueryFixtures(t, db)
	service := NewLogService(db)

	start := mustTime(t, "2026-04-20T09:30:00Z")
	end := mustTime(t, "2026-04-20T12:30:00Z")
	result, err := service.ListFieldChangeLogs(context.Background(), FieldChangeLogListInput{
		OrgID:     100,
		ArticleID: 1,
		FieldName: KeywordScopeBody,
		StartAt:   &start,
		EndAt:     &end,
		Page:      1,
		PageSize:  20,
	})
	if err != nil {
		t.Fatalf("ListFieldChangeLogs() error = %v", err)
	}
	if result.Total != 1 || len(result.Items) != 1 || result.Items[0].FieldName != KeywordScopeBody {
		t.Fatalf("ListFieldChangeLogs() = %+v, want one body change log", result)
	}
}

func TestOperatorResolverUsesActorAndRequestMetadata(t *testing.T) {
	actor := identity.NewActor(23, "jwt-user", "reviewer", "active")
	ctx := identity.ContextWithActor(context.Background(), actor)
	ctx = identity.ContextWithPrincipal(ctx, identity.PrincipalFromActor(actor))
	ctx = identity.ContextWithRequestMetadata(ctx, identity.RequestMetadata{
		RequestID: "req-123",
		SourceIP:  "203.0.113.10",
	})

	operator := ResolveOperator(ctx)
	if operator.ID != 23 || operator.Name != "jwt-user" || operator.Role != "reviewer" {
		t.Fatalf("ResolveOperator() identity = %+v, want actor fields", operator)
	}
	if operator.RequestID != "req-123" || operator.SourceIP != "203.0.113.10" {
		t.Fatalf("ResolveOperator() audit metadata = %+v, want request id and ip", operator)
	}
}

func TestOperatorResolverPreservesAuditMetadataOnLogs(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedLifecycleArticles(t, db)
	service := NewLifecycleService(db)

	actor := identity.NewActor(23, "jwt-user", "reviewer", "active")
	ctx := identity.ContextWithActor(context.Background(), actor)
	ctx = identity.ContextWithRequestMetadata(ctx, identity.RequestMetadata{
		RequestID: "req-456",
		SourceIP:  "198.51.100.25",
	})

	operator := ResolveOperator(ctx)
	changes, err := service.UpdateArticleFields(ctx, UpdateArticleFieldsInput{
		OrgID:        100,
		ArticleID:    12,
		OperatorID:   operator.ID,
		OperatorName: operator.Name,
		Fields: EditableArticleFields{
			Title: "Updated title",
			Body:  "Updated body content",
		},
	})
	if err != nil {
		t.Fatalf("UpdateArticleFields() error = %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("UpdateArticleFields() change count = %d, want %d", len(changes), 2)
	}

	var opLogs []InspectionOperationLog
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

	var changeLogs []InspectionFieldChangeLog
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

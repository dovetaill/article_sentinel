package audit

import (
	"context"
	"testing"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	testutil "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
)

func TestOperationLogQuery(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedQueryFixtures(t, db)
	service := NewLogService(db)

	start := testutil.MustTime(t, "2026-04-20T09:30:00Z")
	end := testutil.MustTime(t, "2026-04-20T12:30:00Z")
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
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedQueryFixtures(t, db)
	service := NewLogService(db)

	start := testutil.MustTime(t, "2026-04-20T09:30:00Z")
	end := testutil.MustTime(t, "2026-04-20T12:30:00Z")
	result, err := service.ListFieldChangeLogs(context.Background(), FieldChangeLogListInput{
		OrgID:     100,
		ArticleID: 1,
		FieldName: domainpkg.KeywordScopeBody,
		StartAt:   &start,
		EndAt:     &end,
		Page:      1,
		PageSize:  20,
	})
	if err != nil {
		t.Fatalf("ListFieldChangeLogs() error = %v", err)
	}
	if result.Total != 1 || len(result.Items) != 1 || result.Items[0].FieldName != domainpkg.KeywordScopeBody {
		t.Fatalf("ListFieldChangeLogs() = %+v, want one body change log", result)
	}
}

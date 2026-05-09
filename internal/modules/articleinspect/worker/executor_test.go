package worker

import (
	"context"
	"errors"
	"testing"

	articlespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/articles"
	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	scanpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/scan"
	"github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
)

func TestExecutorExecuteTask(t *testing.T) {
	t.Run("successful batch updates task counters and status", func(t *testing.T) {
		db := testutil.NewArticleInspectTestDB(t)
		testutil.SeedCandidateArticles(t, db)
		task := testutil.SeedInspectionTaskForWorker(t, db, []scanpkg.KeywordRule{
			{
				ID:            1,
				Name:          "Alpha",
				Category:      "policy",
				MatchType:     domainpkg.MatchTypeContains,
				RiskLevel:     domainpkg.RiskLevelHigh,
				SuggestAction: domainpkg.SuggestActionOffline,
				Scopes:        []string{domainpkg.KeywordScopeTitle},
			},
		})

		executor := NewExecutorWithDeps(db, scanpkg.NewKeywordScanner(), articlespkg.NewArticleRepository(db), 100)
		err := executor.ExecuteTask(context.Background(), queuetasks.ArticleInspectTaskPayload{
			TaskID: task.ID,
			OrgID:  task.OrgID,
		})
		if err != nil {
			t.Fatalf("ExecuteTask() error = %v", err)
		}

		var stored domainpkg.InspectionTask
		if err := db.First(&stored, task.ID).Error; err != nil {
			t.Fatalf("load task error = %v", err)
		}
		if stored.Status != domainpkg.TaskStatusSuccess {
			t.Fatalf("task.Status = %q, want %q", stored.Status, domainpkg.TaskStatusSuccess)
		}
		if stored.TotalScanned != 2 {
			t.Fatalf("task.TotalScanned = %d, want %d", stored.TotalScanned, 2)
		}
		if stored.HitArticles != 1 || stored.HitCount != 1 || stored.FailCount != 0 {
			t.Fatalf("task counters = hits:%d hit_count:%d fail:%d, want 1/1/0", stored.HitArticles, stored.HitCount, stored.FailCount)
		}

		var results []domainpkg.InspectionResult
		if err := db.Where("orgid = ? AND task_id = ?", task.OrgID, task.ID).Find(&results).Error; err != nil {
			t.Fatalf("load results error = %v", err)
		}
		if len(results) != 1 || results[0].ArticleID != 1 {
			t.Fatalf("results = %#v, want one result for article 1", results)
		}

		var hits []domainpkg.InspectionResultHit
		if err := db.Where("orgid = ? AND task_id = ?", task.OrgID, task.ID).Find(&hits).Error; err != nil {
			t.Fatalf("load hits error = %v", err)
		}
		if len(hits) != 1 || hits[0].FieldName != domainpkg.KeywordScopeTitle {
			t.Fatalf("hits = %#v, want one title hit", hits)
		}
	})

	t.Run("mixed batch failures end in partial_success", func(t *testing.T) {
		db := testutil.NewArticleInspectTestDB(t)
		testutil.SeedCandidateArticles(t, db)
		task := testutil.SeedInspectionTaskForWorker(t, db, []scanpkg.KeywordRule{
			{
				ID:            1,
				Name:          "a",
				Category:      "policy",
				MatchType:     domainpkg.MatchTypeContains,
				RiskLevel:     domainpkg.RiskLevelLow,
				SuggestAction: domainpkg.SuggestActionProcess,
				Scopes:        []string{domainpkg.KeywordScopeTitle},
			},
		})

		executor := &Executor{
			db: db,
			scanner: scannerFunc(func(ctx context.Context, article scanpkg.CandidateArticle, rules []scanpkg.KeywordRule) ([]scanpkg.Hit, error) {
				if article.ID == 2 {
					return nil, errors.New("scan failed")
				}
				return []scanpkg.Hit{{
					KeywordID:     rules[0].ID,
					KeywordText:   rules[0].Name,
					Category:      rules[0].Category,
					FieldName:     domainpkg.KeywordScopeTitle,
					MatchType:     rules[0].MatchType,
					RiskLevel:     rules[0].RiskLevel,
					SuggestAction: rules[0].SuggestAction,
					MatchedText:   "A",
					Snippet:       "Alpha",
				}}, nil
			}),
			articleRepo: articlespkg.NewArticleRepository(db),
		}

		err := executor.ExecuteTask(context.Background(), queuetasks.ArticleInspectTaskPayload{
			TaskID: task.ID,
			OrgID:  task.OrgID,
		})
		if err != nil {
			t.Fatalf("ExecuteTask() error = %v", err)
		}

		var stored domainpkg.InspectionTask
		if err := db.First(&stored, task.ID).Error; err != nil {
			t.Fatalf("load task error = %v", err)
		}
		if stored.Status != domainpkg.TaskStatusPartialSuccess {
			t.Fatalf("task.Status = %q, want %q", stored.Status, domainpkg.TaskStatusPartialSuccess)
		}
		if stored.FailCount != 1 {
			t.Fatalf("task.FailCount = %d, want %d", stored.FailCount, 1)
		}
	})
}

type scannerFunc func(ctx context.Context, article scanpkg.CandidateArticle, rules []scanpkg.KeywordRule) ([]scanpkg.Hit, error)

func (fn scannerFunc) ScanArticle(ctx context.Context, article scanpkg.CandidateArticle, rules []scanpkg.KeywordRule) ([]scanpkg.Hit, error) {
	return fn(ctx, article, rules)
}

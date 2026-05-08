package articleinspect

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
)

func TestArticleInspectWorker(t *testing.T) {
	t.Run("successful batch updates task counters and status", func(t *testing.T) {
		db := newArticleInspectTestDB(t)
		seedCandidateArticles(t, db)
		task := seedInspectionTaskForWorker(t, db, []KeywordRule{
			{
				ID:            1,
				Name:          "Alpha",
				Category:      "policy",
				MatchType:     MatchTypeContains,
				RiskLevel:     RiskLevelHigh,
				SuggestAction: SuggestActionOffline,
				Scopes:        []string{KeywordScopeTitle},
			},
		})

		worker := NewWorker(db)
		err := worker.ExecuteTask(context.Background(), queuetasks.ArticleInspectTaskPayload{
			TaskID: task.ID,
			OrgID:  task.OrgID,
		})
		if err != nil {
			t.Fatalf("ExecuteTask() error = %v", err)
		}

		var stored InspectionTask
		if err := db.First(&stored, task.ID).Error; err != nil {
			t.Fatalf("load task error = %v", err)
		}
		if stored.Status != TaskStatusSuccess {
			t.Fatalf("task.Status = %q, want %q", stored.Status, TaskStatusSuccess)
		}
		if stored.TotalScanned != 2 {
			t.Fatalf("task.TotalScanned = %d, want %d", stored.TotalScanned, 2)
		}
		if stored.HitArticles != 1 || stored.HitCount != 1 || stored.FailCount != 0 {
			t.Fatalf("task counters = hits:%d hit_count:%d fail:%d, want 1/1/0", stored.HitArticles, stored.HitCount, stored.FailCount)
		}

		var results []InspectionResult
		if err := db.Where("orgid = ? AND task_id = ?", task.OrgID, task.ID).Find(&results).Error; err != nil {
			t.Fatalf("load results error = %v", err)
		}
		if len(results) != 1 || results[0].ArticleID != 1 {
			t.Fatalf("results = %#v, want one result for article 1", results)
		}

		var hits []InspectionResultHit
		if err := db.Where("orgid = ? AND task_id = ?", task.OrgID, task.ID).Find(&hits).Error; err != nil {
			t.Fatalf("load hits error = %v", err)
		}
		if len(hits) != 1 || hits[0].FieldName != KeywordScopeTitle {
			t.Fatalf("hits = %#v, want one title hit", hits)
		}
	})

	t.Run("mixed batch failures end in partial_success", func(t *testing.T) {
		db := newArticleInspectTestDB(t)
		seedCandidateArticles(t, db)
		task := seedInspectionTaskForWorker(t, db, []KeywordRule{
			{
				ID:            1,
				Name:          "a",
				Category:      "policy",
				MatchType:     MatchTypeContains,
				RiskLevel:     RiskLevelLow,
				SuggestAction: SuggestActionProcess,
				Scopes:        []string{KeywordScopeTitle},
			},
		})

		worker := &Worker{
			db: db,
			scanner: scannerFunc(func(ctx context.Context, article CandidateArticle, rules []KeywordRule) ([]Hit, error) {
				if article.ID == 2 {
					return nil, errors.New("scan failed")
				}
				return []Hit{{
					KeywordID:     rules[0].ID,
					KeywordText:   rules[0].Name,
					Category:      rules[0].Category,
					FieldName:     KeywordScopeTitle,
					MatchType:     rules[0].MatchType,
					RiskLevel:     rules[0].RiskLevel,
					SuggestAction: rules[0].SuggestAction,
					MatchedText:   "A",
					Snippet:       "Alpha",
				}}, nil
			}),
			articleRepo: NewArticleRepository(db),
		}

		err := worker.ExecuteTask(context.Background(), queuetasks.ArticleInspectTaskPayload{
			TaskID: task.ID,
			OrgID:  task.OrgID,
		})
		if err != nil {
			t.Fatalf("ExecuteTask() error = %v", err)
		}

		var stored InspectionTask
		if err := db.First(&stored, task.ID).Error; err != nil {
			t.Fatalf("load task error = %v", err)
		}
		if stored.Status != TaskStatusPartialSuccess {
			t.Fatalf("task.Status = %q, want %q", stored.Status, TaskStatusPartialSuccess)
		}
		if stored.FailCount != 1 {
			t.Fatalf("task.FailCount = %d, want %d", stored.FailCount, 1)
		}
	})
}

func TestDecodeTaskRulesFromKeywordDTOJSON(t *testing.T) {
	snapshot, err := json.Marshal([]KeywordDTO{
		{
			ID:            9101013,
			OrgID:         29,
			Name:          "svg",
			CategoryID:    502,
			CategoryName:  "高频违规",
			MatchType:     MatchTypeContains,
			RiskLevel:     RiskLevelHigh,
			SuggestAction: SuggestActionOffline,
			Enabled:       true,
			Scopes:        []string{KeywordScopeTitle},
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	rules, err := decodeTaskRules(string(snapshot))
	if err != nil {
		t.Fatalf("decodeTaskRules() error = %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("decodeTaskRules() len = %d, want %d", len(rules), 1)
	}
	if rules[0].Name != "svg" || rules[0].MatchType != MatchTypeContains || rules[0].RiskLevel != RiskLevelHigh || rules[0].SuggestAction != SuggestActionOffline {
		t.Fatalf("decodeTaskRules() rule = %+v, want populated match metadata", rules[0])
	}
	if len(rules[0].Scopes) != 1 || rules[0].Scopes[0] != KeywordScopeTitle {
		t.Fatalf("decodeTaskRules() scopes = %#v, want %#v", rules[0].Scopes, []string{KeywordScopeTitle})
	}
}

func timePointer(value time.Time) *time.Time {
	return &value
}

type scannerFunc func(ctx context.Context, article CandidateArticle, rules []KeywordRule) ([]Hit, error)

func (fn scannerFunc) ScanArticle(ctx context.Context, article CandidateArticle, rules []KeywordRule) ([]Hit, error) {
	return fn(ctx, article, rules)
}

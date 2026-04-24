package articleinspect

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/article-sentinel/internal/api/response"
	"github.com/dovetaill/article-sentinel/internal/identity"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInspectionModelMetadata(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "org table", got: (ChuangqiOrg{}).TableName(), want: "xt_chuangqi_org"},
		{name: "category table", got: (InspectionCategory{}).TableName(), want: "xt_article_inspect_categories"},
		{name: "keyword table", got: (InspectionKeyword{}).TableName(), want: "xt_article_inspect_keywords"},
		{name: "keyword scope table", got: (InspectionKeywordScope{}).TableName(), want: "xt_article_inspect_keyword_scopes"},
		{name: "task table", got: (InspectionTask{}).TableName(), want: "xt_article_inspect_tasks"},
		{name: "task keyword table", got: (InspectionTaskKeyword{}).TableName(), want: "xt_article_inspect_task_keywords"},
		{name: "result table", got: (InspectionResult{}).TableName(), want: "xt_article_inspect_results"},
		{name: "result hit table", got: (InspectionResultHit{}).TableName(), want: "xt_article_inspect_result_hits"},
		{name: "action table", got: (InspectionAction{}).TableName(), want: "xt_article_inspect_actions"},
		{name: "field change log table", got: (InspectionFieldChangeLog{}).TableName(), want: "xt_article_inspect_field_change_logs"},
		{name: "operation log table", got: (InspectionOperationLog{}).TableName(), want: "xt_article_inspect_operation_logs"},
		{name: "article table", got: (Article{}).TableName(), want: "xt_article"},
		{name: "article info table", got: (ArticleInfo{}).TableName(), want: "xt_article_info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("TableName() = %q, want %q", tt.got, tt.want)
			}
		})
	}

	wantArticleStates := map[int8]string{
		ArticleStateDeleted:      "del",
		ArticleStateAuditPending: "audit",
		ArticleStateAuditBack:    "back",
		ArticleStateDraft:        "draft",
		ArticleStateStep:         "step",
		ArticleStateOfflineSync:  "offline_sync",
		ArticleStateOffline:      "offline",
		ArticleStateOnline:       "online",
	}
	if len(ArticleLifecycleStates) != len(wantArticleStates) {
		t.Fatalf("ArticleLifecycleStates len = %d, want %d", len(ArticleLifecycleStates), len(wantArticleStates))
	}
	for state, want := range wantArticleStates {
		if got := ArticleLifecycleStates[state]; got != want {
			t.Fatalf("ArticleLifecycleStates[%d] = %q, want %q", state, got, want)
		}
	}

	wantTaskStatuses := []string{
		TaskStatusPending,
		TaskStatusRunning,
		TaskStatusSuccess,
		TaskStatusFailed,
		TaskStatusPartialSuccess,
	}
	assertExactStringSet(t, InspectionTaskStatuses(), wantTaskStatuses)

	wantResultStatuses := []string{
		ResultDispositionPending,
		ResultDispositionIgnored,
		ResultDispositionProcessed,
		ResultDispositionOfflined,
		ResultDispositionRepublished,
		ResultDispositionFailed,
	}
	assertExactStringSet(t, InspectionResultDispositionStatuses(), wantResultStatuses)
}

func assertExactStringSet(t *testing.T, got []string, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("status count = %d, want %d", len(got), len(want))
	}

	seen := make(map[string]int, len(got))
	for _, item := range got {
		seen[item]++
	}
	for _, item := range want {
		if seen[item] != 1 {
			t.Fatalf("status %q count = %d, want %d", item, seen[item], 1)
		}
		delete(seen, item)
	}
	if len(seen) != 0 {
		t.Fatalf("unexpected extra statuses: %v", seen)
	}
}

func TestMigrationFileContainsInspectionTables(t *testing.T) {
	path := filepath.Join("..", "..", "..", "migrations", "20260420_01_article_inspection.sql")

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	text := string(content)
	requiredTables := []string{
		"xt_chuangqi_org",
		"xt_article_inspect_categories",
		"xt_article_inspect_keywords",
		"xt_article_inspect_keyword_scopes",
		"xt_article_inspect_tasks",
		"xt_article_inspect_task_keywords",
		"xt_article_inspect_results",
		"xt_article_inspect_result_hits",
		"xt_article_inspect_actions",
		"xt_article_inspect_field_change_logs",
		"xt_article_inspect_operation_logs",
	}

	for _, table := range requiredTables {
		if !strings.Contains(text, table) {
			t.Fatalf("migration missing table %q", table)
		}
	}
}

func TestInspectionDocsArtifactsExist(t *testing.T) {
	requiredArtifacts := []string{
		filepath.Join("..", "..", "..", "docs", "article-inspection-api.md"),
		filepath.Join("..", "..", "..", "docs", "article-inspection-pages.md"),
		filepath.Join("..", "..", "..", "scripts", "article_inspection_seed.sql"),
	}

	for _, artifact := range requiredArtifacts {
		info, err := os.Stat(artifact)
		if err != nil {
			t.Fatalf("Stat(%q) error = %v", artifact, err)
		}
		if info.Size() == 0 {
			t.Fatalf("artifact %q is empty", artifact)
		}
	}
}

func TestKeywordService(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedOrgCategoryFixtures(t, db)
	repo := NewKeywordRepository(db)
	service := NewKeywordService(repo)
	ctx := identity.ContextWithActor(context.Background(), identity.NewActor(7, "alice", "ops", "active"))

	created, err := service.Create(ctx, CreateKeywordInput{
		OrgID:         100,
		Name:          "spam",
		CategoryID:    1001,
		MatchType:     MatchTypeContains,
		RiskLevel:     RiskLevelHigh,
		SuggestAction: SuggestActionOffline,
		Enabled:       true,
		Remark:        "watch closely",
		Scopes:        []string{KeywordScopeBody, KeywordScopeTitle, KeywordScopeTitle},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.OrgID != 100 {
		t.Fatalf("Create().OrgID = %d, want %d", created.OrgID, 100)
	}
	if created.CategoryID != 1001 || created.CategoryName != "政策红线" {
		t.Fatalf("Create().Category = %d/%q, want %d/%q", created.CategoryID, created.CategoryName, 1001, "政策红线")
	}
	if !created.Enabled {
		t.Fatal("Create().Enabled = false, want true")
	}
	if created.CreatorID != 7 || created.UpdaterID != 7 {
		t.Fatalf("creator/updater ids = %d/%d, want %d/%d", created.CreatorID, created.UpdaterID, 7, 7)
	}
	if created.CreatorName != "alice" || created.UpdaterName != "alice" {
		t.Fatalf("creator/updater names = %q/%q, want %q/%q", created.CreatorName, created.UpdaterName, "alice", "alice")
	}
	if !reflect.DeepEqual(created.Scopes, []string{KeywordScopeBody, KeywordScopeTitle}) {
		t.Fatalf("Create().Scopes = %#v, want %#v", created.Scopes, []string{KeywordScopeBody, KeywordScopeTitle})
	}

	if _, err := service.Create(ctx, CreateKeywordInput{
		OrgID:         200,
		Name:          "spam",
		CategoryID:    2001,
		MatchType:     MatchTypeContains,
		RiskLevel:     RiskLevelLow,
		SuggestAction: SuggestActionIgnore,
		Enabled:       true,
		Scopes:        []string{KeywordScopeTitle},
	}); err != nil {
		t.Fatalf("Create(other org) error = %v", err)
	}

	listed, err := service.List(ctx, KeywordListInput{OrgID: 100, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if listed.Total != 1 {
		t.Fatalf("List().Total = %d, want %d", listed.Total, 1)
	}
	if len(listed.Items) != 1 {
		t.Fatalf("List().Items len = %d, want %d", len(listed.Items), 1)
	}
	if listed.Items[0].Name != "spam" {
		t.Fatalf("List().Items[0].Name = %q, want %q", listed.Items[0].Name, "spam")
	}

	disabled, err := service.PatchEnabled(ctx, PatchKeywordStatusInput{
		OrgID:     100,
		KeywordID: created.ID,
		Enabled:   false,
	})
	if err != nil {
		t.Fatalf("PatchEnabled(false) error = %v", err)
	}
	if disabled.Enabled {
		t.Fatal("PatchEnabled(false).Enabled = true, want false")
	}

	enabled, err := service.PatchEnabled(ctx, PatchKeywordStatusInput{
		OrgID:     100,
		KeywordID: created.ID,
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("PatchEnabled(true) error = %v", err)
	}
	if !enabled.Enabled {
		t.Fatal("PatchEnabled(true).Enabled = false, want true")
	}
}

func TestKeywordValidation(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedOrgCategoryFixtures(t, db)
	service := NewKeywordService(NewKeywordRepository(db))

	tests := []struct {
		name  string
		input CreateKeywordInput
	}{
		{
			name: "missing orgid",
			input: CreateKeywordInput{
				Name:          "spam",
				CategoryID:    1001,
				MatchType:     MatchTypeContains,
				RiskLevel:     RiskLevelHigh,
				SuggestAction: SuggestActionOffline,
				Enabled:       true,
				Scopes:        []string{KeywordScopeTitle},
			},
		},
		{
			name: "missing category id",
			input: CreateKeywordInput{
				OrgID:         100,
				Name:          "spam",
				MatchType:     MatchTypeContains,
				RiskLevel:     RiskLevelHigh,
				SuggestAction: SuggestActionOffline,
				Enabled:       true,
				Scopes:        []string{KeywordScopeTitle},
			},
		},
		{
			name: "unsupported scope",
			input: CreateKeywordInput{
				OrgID:         100,
				Name:          "spam",
				CategoryID:    1001,
				MatchType:     MatchTypeContains,
				RiskLevel:     RiskLevelHigh,
				SuggestAction: SuggestActionOffline,
				Enabled:       true,
				Scopes:        []string{"summary"},
			},
		},
		{
			name: "unsupported risk",
			input: CreateKeywordInput{
				OrgID:         100,
				Name:          "spam",
				CategoryID:    1001,
				MatchType:     MatchTypeContains,
				RiskLevel:     "critical",
				SuggestAction: SuggestActionOffline,
				Enabled:       true,
				Scopes:        []string{KeywordScopeTitle},
			},
		},
		{
			name: "unsupported action",
			input: CreateKeywordInput{
				OrgID:         100,
				Name:          "spam",
				CategoryID:    1001,
				MatchType:     MatchTypeContains,
				RiskLevel:     RiskLevelHigh,
				SuggestAction: "ban",
				Enabled:       true,
				Scopes:        []string{KeywordScopeTitle},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Create(context.Background(), tt.input)
			if !errors.Is(err, ErrInvalidKeywordInput) {
				t.Fatalf("Create() error = %v, want %v", err, ErrInvalidKeywordInput)
			}
		})
	}
}

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

func TestTaskCreation(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedOrgCategoryFixtures(t, db)
	keywordService := NewKeywordService(NewKeywordRepository(db))
	taskService := NewTaskService(db, NewKeywordRepository(db), NewArticleRepository(db))
	ctx := identity.ContextWithActor(context.Background(), identity.NewActor(9, "operator", "ops", "active"))

	keyword, err := keywordService.Create(ctx, CreateKeywordInput{
		OrgID:         100,
		Name:          "spam",
		CategoryID:    1001,
		MatchType:     MatchTypeContains,
		RiskLevel:     RiskLevelHigh,
		SuggestAction: SuggestActionOffline,
		Enabled:       true,
		Scopes:        []string{KeywordScopeTitle, KeywordScopeBody},
	})
	if err != nil {
		t.Fatalf("Create keyword error = %v", err)
	}

	start := mustTime(t, "2026-04-20T09:00:00Z")
	end := mustTime(t, "2026-04-20T13:00:00Z")
	created, err := taskService.Create(ctx, CreateInspectionTaskInput{
		OrgID:            100,
		KeywordIDs:       []uint64{keyword.ID},
		PublishTimeStart: &start,
		PublishTimeEnd:   &end,
		IncludeBody:      true,
	})
	if err != nil {
		t.Fatalf("Create task error = %v", err)
	}
	if created.OrgID != 100 {
		t.Fatalf("Create().OrgID = %d, want %d", created.OrgID, 100)
	}
	if created.Status != TaskStatusPending {
		t.Fatalf("Create().Status = %q, want %q", created.Status, TaskStatusPending)
	}
	if created.RuleSnapshot == "" || !strings.Contains(created.RuleSnapshot, "spam") {
		t.Fatalf("Create().RuleSnapshot = %q, want contains %q", created.RuleSnapshot, "spam")
	}
	if created.RequestSnapshot == "" || !strings.Contains(created.RequestSnapshot, "\"orgid\":100") {
		t.Fatalf("Create().RequestSnapshot = %q, want contains %q", created.RequestSnapshot, "\"orgid\":100")
	}

	var taskKeywords []InspectionTaskKeyword
	if err := db.Where("orgid = ? AND task_id = ?", 100, created.ID).Find(&taskKeywords).Error; err != nil {
		t.Fatalf("Find task keywords error = %v", err)
	}
	if len(taskKeywords) != 1 || taskKeywords[0].KeywordID != keyword.ID {
		t.Fatalf("task keywords = %#v, want keyword %d linked once", taskKeywords, keyword.ID)
	}

	_, err = taskService.Create(ctx, CreateInspectionTaskInput{
		KeywordIDs: []uint64{keyword.ID},
	})
	if !errors.Is(err, ErrInvalidTaskInput) {
		t.Fatalf("Create(missing orgid) error = %v, want %v", err, ErrInvalidTaskInput)
	}
}

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

func newArticleInspectTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "articleinspect.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	if err := db.AutoMigrate(
		&ChuangqiOrg{},
		&InspectionCategory{},
		&InspectionKeyword{},
		&InspectionKeywordScope{},
		&InspectionTask{},
		&InspectionTaskKeyword{},
		&InspectionResult{},
		&InspectionResultHit{},
		&InspectionAction{},
		&InspectionOperationLog{},
		&InspectionFieldChangeLog{},
		&Article{},
		&ArticleInfo{},
	); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	return db
}

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("time.Parse(%q) error = %v", value, err)
	}
	return parsed
}

func sortedStrings(values []string) []string {
	cloned := append([]string(nil), values...)
	sort.Strings(cloned)
	return cloned
}

func seedCandidateArticles(t *testing.T, db *gorm.DB) {
	t.Helper()

	articles := []Article{
		{ID: 1, OrgID: 100, Title: "Alpha news", State: ArticleStateOnline, PublishAtUnix: mustTime(t, "2026-04-20T10:00:00Z").Unix()},
		{ID: 2, OrgID: 100, Title: "Beta update", State: ArticleStateOnline, PublishAtUnix: mustTime(t, "2026-04-20T11:00:00Z").Unix()},
		{ID: 3, OrgID: 100, Title: "Gamma draft", State: ArticleStateDraft, PublishAtUnix: mustTime(t, "2026-04-20T12:00:00Z").Unix()},
		{ID: 4, OrgID: 200, Title: "Other org", State: ArticleStateOnline, PublishAtUnix: mustTime(t, "2026-04-20T10:30:00Z").Unix()},
	}
	if err := db.Create(&articles).Error; err != nil {
		t.Fatalf("seed articles error = %v", err)
	}

	infos := []ArticleInfo{
		{ID: 1, OrgID: 100, Body: "body one"},
		{ID: 2, OrgID: 100, Body: "body two"},
		{ID: 3, OrgID: 100, Body: "body three"},
		{ID: 4, OrgID: 200, Body: "body four"},
	}
	if err := db.Create(&infos).Error; err != nil {
		t.Fatalf("seed article infos error = %v", err)
	}
}

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

func TestHandlerKeywordTaskAndResultsRoutes(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedOrgCategoryFixtures(t, db)
	seedQueryFixtures(t, db)
	dispatcher := &articleInspectTaskDispatcherStub{}
	handler := newArticleInspectHandler(t, db, dispatcher)

	createdKeyword := sendArticleInspectJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/keywords", map[string]any{
		"orgid":          100,
		"name":           "spam",
		"category_id":    1001,
		"match_type":     MatchTypeContains,
		"risk_level":     RiskLevelHigh,
		"suggest_action": SuggestActionOffline,
		"enabled":        true,
		"scopes":         []string{KeywordScopeTitle, KeywordScopeBody},
	})
	if createdKeyword.status != http.StatusCreated {
		t.Fatalf("create keyword status = %d, want %d", createdKeyword.status, http.StatusCreated)
	}
	if createdKeyword.envelope.Code != 0 {
		t.Fatalf("create keyword envelope = %+v, want success", createdKeyword.envelope)
	}
	createdKeywordData := articleInspectDataMap(t, createdKeyword.envelope.Data)
	keywordID := articleInspectUint64Field(t, createdKeywordData, "id")

	listedKeywords := sendArticleInspectRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/keywords?orgid=100&page=1&page_size=20", nil)
	if listedKeywords.status != http.StatusOK {
		t.Fatalf("list keywords status = %d, want %d", listedKeywords.status, http.StatusOK)
	}
	listData := articleInspectDataMap(t, listedKeywords.envelope.Data)
	if total := articleInspectNumberField(t, listData, "total"); total != 1 {
		t.Fatalf("list keywords total = %v, want %d", total, 1)
	}

	detailKeyword := sendArticleInspectRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/keywords/"+articleInspectUint64String(t, createdKeywordData["id"])+"?orgid=100", nil)
	if detailKeyword.status != http.StatusOK {
		t.Fatalf("get keyword status = %d, want %d", detailKeyword.status, http.StatusOK)
	}

	updatedKeyword := sendArticleInspectJSONRequest(t, handler, http.MethodPut, "/api/v1/article-inspect/keywords/"+articleInspectUint64String(t, createdKeywordData["id"]), map[string]any{
		"orgid":          100,
		"name":           "spam-updated",
		"category_id":    1002,
		"match_type":     MatchTypeContains,
		"risk_level":     RiskLevelHigh,
		"suggest_action": SuggestActionProcess,
		"enabled":        true,
		"remark":         "review immediately",
		"scopes":         []string{KeywordScopeTitle},
	})
	if updatedKeyword.status != http.StatusOK {
		t.Fatalf("update keyword status = %d, want %d", updatedKeyword.status, http.StatusOK)
	}

	patchedKeyword := sendArticleInspectJSONRequest(t, handler, http.MethodPatch, "/api/v1/article-inspect/keywords/"+articleInspectUint64String(t, createdKeywordData["id"])+"/status", map[string]any{
		"orgid":   100,
		"enabled": false,
	})
	if patchedKeyword.status != http.StatusOK {
		t.Fatalf("patch keyword status = %d, want %d", patchedKeyword.status, http.StatusOK)
	}

	createdTask := sendArticleInspectJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/tasks", map[string]any{
		"orgid":         100,
		"keyword_ids":   []uint64{keywordID},
		"include_body":  true,
		"article_state": ArticleStateOnline,
	})
	if createdTask.status != http.StatusCreated {
		t.Fatalf("create task status = %d, want %d", createdTask.status, http.StatusCreated)
	}
	createdTaskData := articleInspectDataMap(t, createdTask.envelope.Data)
	if len(dispatcher.payloads) != 1 {
		t.Fatalf("dispatcher payloads len = %d, want %d", len(dispatcher.payloads), 1)
	}
	if dispatcher.payloads[0].OrgID != 100 || dispatcher.payloads[0].TaskID == 0 {
		t.Fatalf("dispatcher payload = %+v, want orgid and task id", dispatcher.payloads[0])
	}
	listedTasks := sendArticleInspectRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/tasks?orgid=100&page=1&page_size=20", nil)
	if listedTasks.status != http.StatusOK {
		t.Fatalf("list tasks status = %d, want %d", listedTasks.status, http.StatusOK)
	}
	taskListData := articleInspectDataMap(t, listedTasks.envelope.Data)
	if total := articleInspectNumberField(t, taskListData, "total"); total != 1 {
		t.Fatalf("list tasks total = %v, want %d", total, 1)
	}
	taskItems := articleInspectListField(t, taskListData, "items")
	if len(taskItems) != 1 {
		t.Fatalf("list tasks items len = %d, want %d", len(taskItems), 1)
	}
	listedTask := articleInspectDataMap(t, taskItems[0])
	if articleInspectStringField(t, listedTask, "status") != TaskStatusPending {
		t.Fatalf("listed task status = %q, want %q", articleInspectStringField(t, listedTask, "status"), TaskStatusPending)
	}
	taskID := articleInspectUint64Field(t, createdTaskData, "id")
	detailTask := sendArticleInspectRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/tasks/"+strconv.FormatUint(taskID, 10)+"?orgid=100", nil)
	if detailTask.status != http.StatusOK {
		t.Fatalf("get task status = %d, want %d", detailTask.status, http.StatusOK)
	}
	detailTaskData := articleInspectDataMap(t, detailTask.envelope.Data)
	if articleInspectStringField(t, detailTaskData, "task_no") != articleInspectStringField(t, createdTaskData, "task_no") {
		t.Fatalf("task detail task_no = %q, want %q", articleInspectStringField(t, detailTaskData, "task_no"), articleInspectStringField(t, createdTaskData, "task_no"))
	}

	listedResults := sendArticleInspectRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/results?orgid=100&task_id=501&risk_level=high&page=1&page_size=20", nil)
	if listedResults.status != http.StatusOK {
		t.Fatalf("list results status = %d, want %d", listedResults.status, http.StatusOK)
	}
	resultListData := articleInspectDataMap(t, listedResults.envelope.Data)
	if total := articleInspectNumberField(t, resultListData, "total"); total != 1 {
		t.Fatalf("list results total = %v, want %d", total, 1)
	}

	detailResult := sendArticleInspectRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/results/1001?orgid=100", nil)
	if detailResult.status != http.StatusOK {
		t.Fatalf("get result detail status = %d, want %d", detailResult.status, http.StatusOK)
	}
	detailData := articleInspectDataMap(t, detailResult.envelope.Data)
	if _, ok := detailData["hits"].([]any); !ok {
		t.Fatalf("detail hits type = %T, want []any", detailData["hits"])
	}

	deletedKeyword := sendArticleInspectRequest(t, handler, http.MethodDelete, "/api/v1/article-inspect/keywords/"+articleInspectUint64String(t, createdKeywordData["id"])+"?orgid=100", nil)
	if deletedKeyword.status != http.StatusOK {
		t.Fatalf("delete keyword status = %d, want %d", deletedKeyword.status, http.StatusOK)
	}
}

func TestHandlerBatchActionsValidateTargets(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedActionFixtures(t, db)
	handler := newArticleInspectHandler(t, db, &articleInspectTaskDispatcherStub{})

	tests := []struct {
		name string
		path string
	}{
		{name: "batch ignore", path: "/api/v1/article-inspect/actions/batch-ignore"},
		{name: "batch process", path: "/api/v1/article-inspect/actions/batch-process"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sendArticleInspectJSONRequest(t, handler, http.MethodPost, tt.path, map[string]any{
				"orgid":   100,
				"task_id": 501,
				"reason":  "missing targets",
			})
			if result.status != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", result.status, http.StatusBadRequest)
			}
			if result.envelope.Code != http.StatusBadRequest {
				t.Fatalf("envelope = %+v, want bad request code", result.envelope)
			}
		})
	}
}

func TestHandlerRectifyAndRepublishRequireOrgID(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedLifecycleArticles(t, db)
	handler := newArticleInspectHandler(t, db, &articleInspectTaskDispatcherStub{})

	rectify := sendArticleInspectJSONRequest(t, handler, http.MethodPut, "/api/v1/article-inspect/articles/12/rectify", map[string]any{
		"title": "Updated title",
		"body":  "Updated body content",
	})
	if rectify.status != http.StatusBadRequest {
		t.Fatalf("rectify status = %d, want %d", rectify.status, http.StatusBadRequest)
	}

	republish := sendArticleInspectJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/articles/11/republish", map[string]any{
		"reason": "try republish",
	})
	if republish.status != http.StatusBadRequest {
		t.Fatalf("republish status = %d, want %d", republish.status, http.StatusBadRequest)
	}
}

func TestHandlerOrgCategoryAndArticleCenterContracts(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedOrgCategoryFixtures(t, db)
	seedArticleCenterFixtures(t, db)
	handler := newArticleInspectHandler(t, db, &articleInspectTaskDispatcherStub{})

	t.Run("listing organizations returns seeded org", func(t *testing.T) {
		result := sendArticleInspectRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/orgs", nil)
		if result.status != http.StatusOK {
			t.Fatalf("list orgs status = %d, want %d", result.status, http.StatusOK)
		}

		data := articleInspectDataMap(t, result.envelope.Data)
		items := articleInspectListField(t, data, "items")
		if len(items) == 0 {
			t.Fatal("list orgs items = empty, want seeded org")
		}

		first := articleInspectDataMap(t, items[0])
		if articleInspectUint64Field(t, first, "id") != 29 {
			t.Fatalf("first org id = %d, want %d", articleInspectUint64Field(t, first, "id"), 29)
		}
		if articleInspectStringField(t, first, "name") != "一县一端" {
			t.Fatalf("first org name = %q, want %q", articleInspectStringField(t, first, "name"), "一县一端")
		}
	})

	t.Run("categories are scoped by orgid", func(t *testing.T) {
		result := sendArticleInspectRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/categories?orgid=29&page=1&page_size=20", nil)
		if result.status != http.StatusOK {
			t.Fatalf("list categories status = %d, want %d", result.status, http.StatusOK)
		}

		data := articleInspectDataMap(t, result.envelope.Data)
		if total := articleInspectNumberField(t, data, "total"); total != 2 {
			t.Fatalf("list categories total = %v, want %d", total, 2)
		}

		items := articleInspectListField(t, data, "items")
		gotIDs := make([]uint64, 0, len(items))
		for _, raw := range items {
			item := articleInspectDataMap(t, raw)
			if articleInspectUint64Field(t, item, "orgid") != 29 {
				t.Fatalf("category orgid = %d, want %d", articleInspectUint64Field(t, item, "orgid"), 29)
			}
			gotIDs = append(gotIDs, articleInspectUint64Field(t, item, "id"))
		}
		sort.Slice(gotIDs, func(i, j int) bool { return gotIDs[i] < gotIDs[j] })
		if !reflect.DeepEqual(gotIDs, []uint64{501, 502}) {
			t.Fatalf("category ids = %#v, want %#v", gotIDs, []uint64{501, 502})
		}
	})

	t.Run("category CRUD rejects missing orgid", func(t *testing.T) {
		tests := []struct {
			name   string
			method string
			path   string
			body   any
		}{
			{
				name:   "create",
				method: http.MethodPost,
				path:   "/api/v1/article-inspect/categories",
				body: map[string]any{
					"name":    "新增分类",
					"code":    "new-category",
					"enabled": true,
				},
			},
			{
				name:   "detail",
				method: http.MethodGet,
				path:   "/api/v1/article-inspect/categories/501",
			},
			{
				name:   "update",
				method: http.MethodPut,
				path:   "/api/v1/article-inspect/categories/501",
				body: map[string]any{
					"name":    "分类更新",
					"code":    "policy-updated",
					"enabled": true,
				},
			},
			{
				name:   "delete",
				method: http.MethodDelete,
				path:   "/api/v1/article-inspect/categories/501",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var result articleInspectHTTPResult
				if tt.body != nil {
					result = sendArticleInspectJSONRequest(t, handler, tt.method, tt.path, tt.body)
				} else {
					result = sendArticleInspectRequest(t, handler, tt.method, tt.path, nil)
				}
				if result.status != http.StatusBadRequest {
					t.Fatalf("%s status = %d, want %d", tt.name, result.status, http.StatusBadRequest)
				}
				if result.envelope.Code != http.StatusBadRequest {
					t.Fatalf("%s envelope = %+v, want bad request code", tt.name, result.envelope)
				}
			})
		}
	})

	t.Run("keyword create and update payloads use category_id", func(t *testing.T) {
		created := sendArticleInspectJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/keywords", map[string]any{
			"orgid":          29,
			"name":           "敏感词",
			"category_id":    501,
			"match_type":     MatchTypeContains,
			"risk_level":     RiskLevelHigh,
			"suggest_action": SuggestActionOffline,
			"enabled":        true,
			"scopes":         []string{KeywordScopeTitle},
		})
		if created.status != http.StatusCreated {
			t.Fatalf("create keyword status = %d, want %d", created.status, http.StatusCreated)
		}

		createdData := articleInspectDataMap(t, created.envelope.Data)
		if articleInspectUint64Field(t, createdData, "category_id") != 501 {
			t.Fatalf("create keyword category_id = %d, want %d", articleInspectUint64Field(t, createdData, "category_id"), 501)
		}
		if articleInspectStringField(t, createdData, "category_name") != "政策红线" {
			t.Fatalf("create keyword category_name = %q, want %q", articleInspectStringField(t, createdData, "category_name"), "政策红线")
		}

		updated := sendArticleInspectJSONRequest(t, handler, http.MethodPut, "/api/v1/article-inspect/keywords/"+articleInspectUint64String(t, createdData["id"]), map[string]any{
			"orgid":          29,
			"name":           "敏感词-更新",
			"category_id":    502,
			"match_type":     MatchTypeContains,
			"risk_level":     RiskLevelMedium,
			"suggest_action": SuggestActionProcess,
			"enabled":        true,
			"scopes":         []string{KeywordScopeBody},
		})
		if updated.status != http.StatusOK {
			t.Fatalf("update keyword status = %d, want %d", updated.status, http.StatusOK)
		}

		updatedData := articleInspectDataMap(t, updated.envelope.Data)
		if articleInspectUint64Field(t, updatedData, "category_id") != 502 {
			t.Fatalf("update keyword category_id = %d, want %d", articleInspectUint64Field(t, updatedData, "category_id"), 502)
		}
		if articleInspectStringField(t, updatedData, "category_name") != "高频违规" {
			t.Fatalf("update keyword category_name = %q, want %q", articleInspectStringField(t, updatedData, "category_name"), "高频违规")
		}
	})

	t.Run("article list endpoint reads real articles", func(t *testing.T) {
		result := sendArticleInspectRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/articles?orgid=29&page=1&page_size=20", nil)
		if result.status != http.StatusOK {
			t.Fatalf("list articles status = %d, want %d", result.status, http.StatusOK)
		}

		data := articleInspectDataMap(t, result.envelope.Data)
		if total := articleInspectNumberField(t, data, "total"); total != 2 {
			t.Fatalf("list articles total = %v, want %d", total, 2)
		}

		items := articleInspectListField(t, data, "items")
		gotIDs := make([]uint64, 0, len(items))
		for _, raw := range items {
			item := articleInspectDataMap(t, raw)
			gotIDs = append(gotIDs, articleInspectUint64Field(t, item, "id"))
		}
		sort.Slice(gotIDs, func(i, j int) bool { return gotIDs[i] < gotIDs[j] })
		if !reflect.DeepEqual(gotIDs, []uint64{9001, 9002}) {
			t.Fatalf("article ids = %#v, want %#v", gotIDs, []uint64{9001, 9002})
		}
	})

	t.Run("article detail endpoint includes article data and latest inspect summary", func(t *testing.T) {
		result := sendArticleInspectRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/articles/9001?orgid=29", nil)
		if result.status != http.StatusOK {
			t.Fatalf("get article detail status = %d, want %d", result.status, http.StatusOK)
		}

		data := articleInspectDataMap(t, result.envelope.Data)
		if articleInspectUint64Field(t, data, "id") != 9001 {
			t.Fatalf("article detail id = %d, want %d", articleInspectUint64Field(t, data, "id"), 9001)
		}
		if articleInspectStringField(t, data, "title") != "县域要闻一" {
			t.Fatalf("article detail title = %q, want %q", articleInspectStringField(t, data, "title"), "县域要闻一")
		}
		if articleInspectStringField(t, data, "body") != "<p>real body one</p>" {
			t.Fatalf("article detail body = %q, want %q", articleInspectStringField(t, data, "body"), "<p>real body one</p>")
		}
		if articleInspectStringField(t, data, "thumbnail") != "https://example.com/article-9001.png" {
			t.Fatalf("article detail thumbnail = %q, want %q", articleInspectStringField(t, data, "thumbnail"), "https://example.com/article-9001.png")
		}
		if articleInspectUint64Field(t, data, "latest_task_id") != 702 {
			t.Fatalf("article detail latest_task_id = %d, want %d", articleInspectUint64Field(t, data, "latest_task_id"), 702)
		}
		if articleInspectStringField(t, data, "latest_risk_level") != RiskLevelHigh {
			t.Fatalf("article detail latest_risk_level = %q, want %q", articleInspectStringField(t, data, "latest_risk_level"), RiskLevelHigh)
		}
	})
}

func TestRouteRegistrationRegistersArticleInspectPaths(t *testing.T) {
	db := newArticleInspectTestDB(t)
	dispatcher := &articleInspectTaskDispatcherStub{}

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterRoutes(api, Routes{
		Categories: NewCategoryService(NewCategoryRepository(db)),
		Keywords:   NewKeywordService(NewKeywordRepository(db)),
		Tasks:      NewTaskService(db, NewKeywordRepository(db), NewArticleRepository(db)),
		Results:    NewResultService(db),
		Actions:    NewActionService(db, NewActionRepository(db)),
		Lifecycle:  NewLifecycleService(db),
		Logs:       NewLogService(db),
		Articles:   NewArticleService(NewArticleRepository(db)),
		Dispatcher: dispatcher,
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("openapi status = %d, want %d", rec.Code, http.StatusOK)
	}

	var doc struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("decode openapi: %v", err)
	}

	requiredPaths := []string{
		"/api/v1/article-inspect/orgs",
		"/api/v1/article-inspect/categories",
		"/api/v1/article-inspect/categories/{id}",
		"/api/v1/article-inspect/categories/{id}/status",
		"/api/v1/article-inspect/articles",
		"/api/v1/article-inspect/articles/{article_id}",
		"/api/v1/article-inspect/keywords",
		"/api/v1/article-inspect/keywords/{id}",
		"/api/v1/article-inspect/tasks",
		"/api/v1/article-inspect/tasks/{id}",
		"/api/v1/article-inspect/results",
		"/api/v1/article-inspect/results/{id}",
		"/api/v1/article-inspect/actions/batch-ignore",
		"/api/v1/article-inspect/actions/batch-process",
		"/api/v1/article-inspect/articles/{article_id}/rectify",
		"/api/v1/article-inspect/articles/{article_id}/republish",
		"/api/v1/article-inspect/logs/operations",
		"/api/v1/article-inspect/logs/field-changes",
	}

	for _, path := range requiredPaths {
		if _, ok := doc.Paths[path]; !ok {
			t.Fatalf("openapi missing path %s", path)
		}
	}
}

func extractArticleIDs(items []CandidateArticle) []uint64 {
	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

func timePointer(value time.Time) *time.Time {
	return &value
}

type articleInspectHTTPResult struct {
	status   int
	envelope response.Envelope
}

type articleInspectTaskDispatcherStub struct {
	payloads []queuetasks.ArticleInspectTaskPayload
	err      error
}

func (s *articleInspectTaskDispatcherStub) DispatchArticleInspectTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) error {
	_ = ctx
	s.payloads = append(s.payloads, payload)
	return s.err
}

func newArticleInspectHandler(t *testing.T, db *gorm.DB, dispatcher *articleInspectTaskDispatcherStub) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterRoutes(api, Routes{
		Categories: NewCategoryService(NewCategoryRepository(db)),
		Keywords:   NewKeywordService(NewKeywordRepository(db)),
		Tasks:      NewTaskService(db, NewKeywordRepository(db), NewArticleRepository(db)),
		Results:    NewResultService(db),
		Actions:    NewActionService(db, NewActionRepository(db)),
		Lifecycle:  NewLifecycleService(db),
		Logs:       NewLogService(db),
		Articles:   NewArticleService(NewArticleRepository(db)),
		Dispatcher: dispatcher,
	})
	return mux
}

func sendArticleInspectJSONRequest(t *testing.T, handler http.Handler, method, path string, body any) articleInspectHTTPResult {
	t.Helper()

	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	return sendArticleInspectRequest(t, handler, method, path, bytes.NewReader(encoded))
}

func sendArticleInspectRequest(t *testing.T, handler http.Handler, method, path string, body *bytes.Reader) articleInspectHTTPResult {
	t.Helper()

	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, body)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return articleInspectHTTPResult{status: rec.Code, envelope: decodeArticleInspectEnvelope(t, rec)}
}

func decodeArticleInspectEnvelope(t *testing.T, rec *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	body := bytes.TrimSpace(rec.Body.Bytes())
	if len(body) == 0 || body[0] != '{' {
		return response.Envelope{}
	}
	var got response.Envelope
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return got
}

func articleInspectDataMap(t *testing.T, value any) map[string]any {
	t.Helper()
	data, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("data type = %T, want map[string]any", value)
	}
	return data
}

func articleInspectListField(t *testing.T, m map[string]any, key string) []any {
	t.Helper()
	value, ok := m[key]
	if !ok {
		t.Fatalf("missing key %q", key)
	}
	items, ok := value.([]any)
	if !ok {
		t.Fatalf("key %q type = %T, want []any", key, value)
	}
	return items
}

func articleInspectNumberField(t *testing.T, m map[string]any, key string) float64 {
	t.Helper()
	v, ok := m[key]
	if !ok {
		t.Fatalf("missing key %q", key)
	}
	switch value := v.(type) {
	case float64:
		return value
	case json.Number:
		number, err := value.Float64()
		if err != nil {
			t.Fatalf("json number %q: %v", value, err)
		}
		return number
	default:
		t.Fatalf("key %q type = %T, want numeric", key, v)
		return 0
	}
}

func articleInspectStringField(t *testing.T, m map[string]any, key string) string {
	t.Helper()
	value, ok := m[key]
	if !ok {
		t.Fatalf("missing key %q", key)
	}
	text, ok := value.(string)
	if !ok {
		t.Fatalf("key %q type = %T, want string", key, value)
	}
	return text
}

func articleInspectUint64Field(t *testing.T, m map[string]any, key string) uint64 {
	t.Helper()
	return uint64(articleInspectNumberField(t, m, key))
}

func articleInspectUint64String(t *testing.T, value any) string {
	t.Helper()
	switch item := value.(type) {
	case float64:
		return strconv.FormatUint(uint64(item), 10)
	case json.Number:
		return item.String()
	default:
		t.Fatalf("id type = %T, want numeric", value)
		return ""
	}
}

func seedInspectionTaskForWorker(t *testing.T, db *gorm.DB, rules []KeywordRule) *InspectionTask {
	t.Helper()

	start := mustTime(t, "2026-04-20T09:00:00Z")
	end := mustTime(t, "2026-04-20T13:00:00Z")
	ruleSnapshot, err := marshalJSON(rules)
	if err != nil {
		t.Fatalf("marshal rule snapshot error = %v", err)
	}
	requestSnapshot, err := marshalJSON(map[string]any{
		"orgid":              uint64(100),
		"publish_time_start": start,
		"publish_time_end":   end,
		"include_body":       true,
	})
	if err != nil {
		t.Fatalf("marshal request snapshot error = %v", err)
	}

	task := &InspectionTask{
		OrgID:              100,
		TaskNo:             "inspect-test",
		Status:             TaskStatusPending,
		ArticleStateFilter: "9",
		PublishTimeStart:   timePointer(start),
		PublishTimeEnd:     timePointer(end),
		IncludeBody:        true,
		RequestSnapshot:    requestSnapshot,
		RuleSnapshot:       ruleSnapshot,
	}
	if err := db.Create(task).Error; err != nil {
		t.Fatalf("create task error = %v", err)
	}
	return task
}

func marshalJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type scannerFunc func(ctx context.Context, article CandidateArticle, rules []KeywordRule) ([]Hit, error)

func (fn scannerFunc) ScanArticle(ctx context.Context, article CandidateArticle, rules []KeywordRule) ([]Hit, error) {
	return fn(ctx, article, rules)
}

func seedLifecycleArticles(t *testing.T, db *gorm.DB) {
	t.Helper()

	articles := []Article{
		{ID: 10, OrgID: 100, Title: "Online article", State: ArticleStateOnline},
		{ID: 11, OrgID: 100, Title: "Offline article", State: ArticleStateOffline},
		{ID: 12, OrgID: 100, Title: "Needs rectify", State: ArticleStateOffline},
	}
	if err := db.Create(&articles).Error; err != nil {
		t.Fatalf("seed lifecycle articles error = %v", err)
	}
	if err := db.Create(&[]ArticleInfo{
		{ID: 10, OrgID: 100, Body: "body a"},
		{ID: 11, OrgID: 100, Body: "body b"},
		{ID: 12, OrgID: 100, Body: "body c"},
	}).Error; err != nil {
		t.Fatalf("seed lifecycle article info error = %v", err)
	}
}

func seedOrgCategoryFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS xt_chuangqi_org (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			cateid INTEGER NOT NULL DEFAULT 0,
			enabled INTEGER NOT NULL DEFAULT 1,
			sort INTEGER NOT NULL DEFAULT 0,
			create_at DATETIME NOT NULL,
			update_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS xt_article_inspect_categories (
			id INTEGER PRIMARY KEY,
			orgid INTEGER NOT NULL,
			name TEXT NOT NULL,
			code TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			sort INTEGER NOT NULL DEFAULT 0,
			creator_id INTEGER NOT NULL DEFAULT 0,
			creator_name TEXT NOT NULL DEFAULT '',
			updater_id INTEGER NOT NULL DEFAULT 0,
			updater_name TEXT NOT NULL DEFAULT '',
			create_at DATETIME NOT NULL,
			update_at DATETIME NOT NULL
		)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("prepare org/category tables error = %v", err)
		}
	}

	timestamp := mustTime(t, "2026-04-20T08:00:00Z")
	if err := db.Exec(
		`INSERT INTO xt_chuangqi_org (id, name, cateid, enabled, sort, create_at, update_at) VALUES
			(29, '一县一端', 0, 1, 10, ?, ?),
			(30, '其他组织', 0, 1, 20, ?, ?),
			(100, '测试组织A', 0, 1, 30, ?, ?),
			(200, '测试组织B', 0, 1, 40, ?, ?)`,
		timestamp, timestamp, timestamp, timestamp, timestamp, timestamp, timestamp, timestamp,
	).Error; err != nil {
		t.Fatalf("seed orgs error = %v", err)
	}

	if err := db.Exec(
		`INSERT INTO xt_article_inspect_categories (id, orgid, name, code, enabled, sort, creator_id, creator_name, updater_id, updater_name, create_at, update_at) VALUES
			(501, 29, '政策红线', 'policy', 1, 10, 7, 'alice', 7, 'alice', ?, ?),
			(502, 29, '高频违规', 'risk', 1, 20, 7, 'alice', 7, 'alice', ?, ?),
			(601, 30, '外部分类', 'external', 1, 10, 8, 'bob', 8, 'bob', ?, ?),
			(1001, 100, '政策红线', 'policy', 1, 10, 7, 'alice', 7, 'alice', ?, ?),
			(1002, 100, '高频违规', 'risk', 1, 20, 7, 'alice', 7, 'alice', ?, ?),
			(2001, 200, '其他组织分类', 'other-org', 1, 10, 8, 'bob', 8, 'bob', ?, ?)`,
		timestamp, timestamp, timestamp, timestamp, timestamp, timestamp, timestamp, timestamp, timestamp, timestamp, timestamp, timestamp,
	).Error; err != nil {
		t.Fatalf("seed categories error = %v", err)
	}
}

func seedArticleCenterFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()

	publishAt := mustTime(t, "2026-04-21T09:00:00Z")
	laterPublishAt := mustTime(t, "2026-04-21T10:30:00Z")
	latestActionAt := mustTime(t, "2026-04-21T12:30:00Z")
	olderActionAt := mustTime(t, "2026-04-21T11:00:00Z")

	articles := []Article{
		{ID: 9001, OrgID: 29, Title: "县域要闻一", Thumbnail: "https://example.com/article-9001.png", State: ArticleStateOnline, PublishAtUnix: publishAt.Unix(), UpdateAtUnix: latestActionAt.Unix()},
		{ID: 9002, OrgID: 29, Title: "县域要闻二", State: ArticleStateOffline, PublishAtUnix: laterPublishAt.Unix(), UpdateAtUnix: laterPublishAt.Unix()},
		{ID: 9901, OrgID: 30, Title: "外部组织稿件", State: ArticleStateOnline, PublishAtUnix: publishAt.Unix(), UpdateAtUnix: publishAt.Unix()},
	}
	if err := db.Create(&articles).Error; err != nil {
		t.Fatalf("seed article center articles error = %v", err)
	}

	infos := []ArticleInfo{
		{ID: 9001, OrgID: 29, Body: "<p>real body one</p>"},
		{ID: 9002, OrgID: 29, Body: "<p>real body two</p>"},
		{ID: 9901, OrgID: 30, Body: "<p>other org body</p>"},
	}
	if err := db.Create(&infos).Error; err != nil {
		t.Fatalf("seed article center infos error = %v", err)
	}

	results := []InspectionResult{
		{
			ID:                 7101,
			OrgID:              29,
			TaskID:             701,
			ArticleID:          9001,
			ArticleTitle:       "县域要闻一",
			ArticleState:       ArticleStateOnline,
			PublishAtTime:      &publishAt,
			RiskLevel:          RiskLevelLow,
			SuggestAction:      SuggestActionIgnore,
			HitCount:           1,
			DispositionStatus:  ResultDispositionPending,
			LatestActionAt:     &olderActionAt,
			LatestOperatorID:   7,
			LatestOperatorName: "alice",
		},
		{
			ID:                 7102,
			OrgID:              29,
			TaskID:             702,
			ArticleID:          9001,
			ArticleTitle:       "县域要闻一",
			ArticleState:       ArticleStateOnline,
			PublishAtTime:      &publishAt,
			RiskLevel:          RiskLevelHigh,
			SuggestAction:      SuggestActionOffline,
			HitCount:           2,
			DispositionStatus:  ResultDispositionProcessed,
			LatestActionAt:     &latestActionAt,
			LatestOperatorID:   8,
			LatestOperatorName: "bob",
		},
	}
	if err := db.Create(&results).Error; err != nil {
		t.Fatalf("seed article center results error = %v", err)
	}
}

func seedActionFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()

	results := []InspectionResult{
		{ID: 1001, OrgID: 100, TaskID: 501, ArticleID: 1, DispositionStatus: ResultDispositionPending},
		{ID: 1002, OrgID: 100, TaskID: 501, ArticleID: 2, DispositionStatus: ResultDispositionIgnored},
		{ID: 1003, OrgID: 100, TaskID: 501, ArticleID: 3, DispositionStatus: ResultDispositionPending},
		{ID: 1004, OrgID: 100, TaskID: 501, ArticleID: 4, DispositionStatus: ResultDispositionProcessed},
	}
	if err := db.Create(&results).Error; err != nil {
		t.Fatalf("seed action results error = %v", err)
	}
}

func seedQueryFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()

	publishAt := mustTime(t, "2026-04-20T10:00:00Z")
	later := mustTime(t, "2026-04-20T11:00:00Z")
	createAt := mustTime(t, "2026-04-20T10:30:00Z")
	updateAt := mustTime(t, "2026-04-20T12:00:00Z")

	results := []InspectionResult{
		{ID: 1001, OrgID: 100, TaskID: 501, ArticleID: 1, ArticleTitle: "Alpha news", ArticleState: ArticleStateOnline, PublishAtTime: &publishAt, RiskLevel: RiskLevelHigh, SuggestAction: SuggestActionOffline, HitFieldsCount: 2, HitKeywordsCount: 2, HitCount: 2, DispositionStatus: ResultDispositionPending},
		{ID: 1002, OrgID: 100, TaskID: 501, ArticleID: 2, ArticleTitle: "Beta update", ArticleState: ArticleStateOnline, PublishAtTime: &later, RiskLevel: RiskLevelLow, SuggestAction: SuggestActionProcess, HitFieldsCount: 1, HitKeywordsCount: 1, HitCount: 1, DispositionStatus: ResultDispositionProcessed},
		{ID: 2001, OrgID: 200, TaskID: 601, ArticleID: 9, ArticleTitle: "Other org", ArticleState: ArticleStateOnline, PublishAtTime: &later, RiskLevel: RiskLevelHigh, SuggestAction: SuggestActionOffline, HitFieldsCount: 1, HitKeywordsCount: 1, HitCount: 1, DispositionStatus: ResultDispositionPending},
	}
	if err := db.Create(&results).Error; err != nil {
		t.Fatalf("seed query results error = %v", err)
	}

	hits := []InspectionResultHit{
		{ID: 1, OrgID: 100, TaskID: 501, ResultID: 1001, ArticleID: 1, KeywordID: 1, KeywordText: "alpha", FieldName: KeywordScopeTitle, MatchType: MatchTypeContains, RiskLevel: RiskLevelHigh, SuggestAction: SuggestActionOffline, MatchedText: "Alpha", Snippet: "Alpha news"},
		{ID: 2, OrgID: 100, TaskID: 501, ResultID: 1001, ArticleID: 1, KeywordID: 2, KeywordText: "body", FieldName: KeywordScopeBody, MatchType: MatchTypeContains, RiskLevel: RiskLevelHigh, SuggestAction: SuggestActionOffline, MatchedText: "body", Snippet: "body snippet"},
		{ID: 3, OrgID: 100, TaskID: 501, ResultID: 1002, ArticleID: 2, KeywordID: 3, KeywordText: "beta", FieldName: KeywordScopeTitle, MatchType: MatchTypeContains, RiskLevel: RiskLevelLow, SuggestAction: SuggestActionProcess, MatchedText: "Beta", Snippet: "Beta update"},
	}
	if err := db.Create(&hits).Error; err != nil {
		t.Fatalf("seed query hits error = %v", err)
	}

	opLogs := []InspectionOperationLog{
		{ID: 1, OrgID: 100, TaskID: 501, ResultID: 1001, ArticleID: 1, OperationType: ActionTypeOffline, BeforeState: "9", AfterState: "8", Status: ActionStatusSuccess, OperatorID: 7, OperatorName: "alice", InspectionTimestamps: InspectionTimestamps{CreateAt: createAt, UpdateAt: createAt}},
		{ID: 2, OrgID: 100, TaskID: 501, ResultID: 1001, ArticleID: 1, OperationType: ActionTypeRectify, BeforeState: "8", AfterState: "8", Status: ActionStatusSuccess, OperatorID: 7, OperatorName: "alice", InspectionTimestamps: InspectionTimestamps{CreateAt: updateAt, UpdateAt: updateAt}},
		{ID: 3, OrgID: 100, TaskID: 501, ResultID: 1002, ArticleID: 2, OperationType: ActionTypeBatchProcess, BeforeState: ResultDispositionPending, AfterState: ResultDispositionProcessed, Status: ActionStatusSuccess, OperatorID: 8, OperatorName: "bob", InspectionTimestamps: InspectionTimestamps{CreateAt: updateAt, UpdateAt: updateAt}},
	}
	if err := db.Create(&opLogs).Error; err != nil {
		t.Fatalf("seed operation logs error = %v", err)
	}

	changeLogs := []InspectionFieldChangeLog{
		{ID: 1, OrgID: 100, TaskID: 501, ResultID: 1001, ArticleID: 1, FieldName: KeywordScopeBody, BeforeValue: "old", AfterValue: "new", DiffSummary: "body: old -> new", OperatorID: 7, OperatorName: "alice", InspectionTimestamps: InspectionTimestamps{CreateAt: updateAt, UpdateAt: updateAt}},
		{ID: 2, OrgID: 100, TaskID: 501, ResultID: 1001, ArticleID: 1, FieldName: KeywordScopeTitle, BeforeValue: "old title", AfterValue: "new title", DiffSummary: "title: old title -> new title", OperatorID: 7, OperatorName: "alice", InspectionTimestamps: InspectionTimestamps{CreateAt: createAt, UpdateAt: createAt}},
	}
	if err := db.Create(&changeLogs).Error; err != nil {
		t.Fatalf("seed field change logs error = %v", err)
	}
}

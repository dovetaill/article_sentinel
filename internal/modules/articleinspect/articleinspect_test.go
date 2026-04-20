package articleinspect

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

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

func TestKeywordService(t *testing.T) {
	db := newArticleInspectTestDB(t)
	repo := NewKeywordRepository(db)
	service := NewKeywordService(repo)
	ctx := identity.ContextWithActor(context.Background(), identity.NewActor(7, "alice", "ops", "active"))

	created, err := service.Create(ctx, CreateKeywordInput{
		OrgID:         100,
		Name:          "spam",
		Category:      "policy",
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
		Category:      "policy",
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
	service := NewKeywordService(NewKeywordRepository(db))

	tests := []struct {
		name  string
		input CreateKeywordInput
	}{
		{
			name: "missing orgid",
			input: CreateKeywordInput{
				Name:          "spam",
				Category:      "policy",
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
				Category:      "policy",
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
				Category:      "policy",
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
				Category:      "policy",
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
	keywordService := NewKeywordService(NewKeywordRepository(db))
	taskService := NewTaskService(db, NewKeywordRepository(db), NewArticleRepository(db))
	ctx := identity.ContextWithActor(context.Background(), identity.NewActor(9, "operator", "ops", "active"))

	keyword, err := keywordService.Create(ctx, CreateKeywordInput{
		OrgID:         100,
		Name:          "spam",
		Category:      "policy",
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

func newArticleInspectTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "articleinspect.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	if err := db.AutoMigrate(
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
		{ID: 1, OrgID: 100, Title: "Alpha news", State: ArticleStateOnline, PublishAtTime: timePointer(mustTime(t, "2026-04-20T10:00:00Z"))},
		{ID: 2, OrgID: 100, Title: "Beta update", State: ArticleStateOnline, PublishAtTime: timePointer(mustTime(t, "2026-04-20T11:00:00Z"))},
		{ID: 3, OrgID: 100, Title: "Gamma draft", State: ArticleStateDraft, PublishAtTime: timePointer(mustTime(t, "2026-04-20T12:00:00Z"))},
		{ID: 4, OrgID: 200, Title: "Other org", State: ArticleStateOnline, PublishAtTime: timePointer(mustTime(t, "2026-04-20T10:30:00Z"))},
	}
	if err := db.Create(&articles).Error; err != nil {
		t.Fatalf("seed articles error = %v", err)
	}

	infos := []ArticleInfo{
		{ArticleID: 1, Body: "body one"},
		{ArticleID: 2, Body: "body two"},
		{ArticleID: 3, Body: "body three"},
		{ArticleID: 4, Body: "body four"},
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
	if err := db.Create(&[]ArticleInfo{{ArticleID: 10, Body: "body a"}, {ArticleID: 11, Body: "body b"}, {ArticleID: 12, Body: "body c"}}).Error; err != nil {
		t.Fatalf("seed lifecycle article info error = %v", err)
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

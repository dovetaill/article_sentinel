package articleinspect

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		{name: "task outbox table", got: (InspectionTaskOutboxMessage{}).TableName(), want: "xt_article_inspect_task_outbox"},
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
	dropPath := filepath.Join("..", "..", "..", "migrations", "20260428_01_drop_category_code.sql")

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	dropContent, err := os.ReadFile(dropPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", dropPath, err)
	}

	text := string(content)
	dropText := string(dropContent)
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
	if strings.Contains(text, "`code` VARCHAR(64) NOT NULL") {
		t.Fatalf("base migration still contains category code column")
	}
	if strings.Contains(text, "uk_org_code") {
		t.Fatalf("base migration still contains uk_org_code")
	}
	requiredDropSnippets := []string{
		"information_schema.statistics",
		"index_name = 'uk_org_code'",
		"DROP INDEX `uk_org_code`",
		"information_schema.columns",
		"column_name = 'code'",
		"DROP COLUMN `code`",
	}
	for _, snippet := range requiredDropSnippets {
		if !strings.Contains(dropText, snippet) {
			t.Fatalf("drop migration missing snippet %q", snippet)
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

func TestTaskCreationWithOutbox(t *testing.T) {
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
		Scopes:        []string{KeywordScopeTitle},
	})
	if err != nil {
		t.Fatalf("Create keyword error = %v", err)
	}

	created, outbox, err := taskService.CreateWithOutbox(ctx, CreateInspectionTaskInput{
		OrgID:          100,
		KeywordIDs:     []uint64{keyword.ID},
		IncludeBody:    true,
		ArticleState:   ArticleStateOnline,
		PublishTimeEnd: timePointer(mustTime(t, "2026-04-20T13:00:00Z")),
	})
	if err != nil {
		t.Fatalf("CreateWithOutbox() error = %v", err)
	}
	if created.ID == 0 {
		t.Fatal("CreateWithOutbox().Task.ID = 0, want persisted task id")
	}
	if outbox.ID == 0 {
		t.Fatal("CreateWithOutbox().Outbox.ID = 0, want persisted outbox id")
	}
	if outbox.Status != TaskOutboxStatusPending {
		t.Fatalf("CreateWithOutbox().Outbox.Status = %q, want %q", outbox.Status, TaskOutboxStatusPending)
	}
	if outbox.MessageType != TaskOutboxMessageTypeRunTask {
		t.Fatalf("CreateWithOutbox().Outbox.MessageType = %q, want %q", outbox.MessageType, TaskOutboxMessageTypeRunTask)
	}
	if !strings.Contains(outbox.Payload, fmt.Sprintf("\"task_id\":%d", created.ID)) {
		t.Fatalf("CreateWithOutbox().Outbox.Payload = %q, want contains task id %d", outbox.Payload, created.ID)
	}

	var stored InspectionTaskOutboxMessage
	if err := db.Where("orgid = ? AND task_id = ?", 100, created.ID).First(&stored).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	if stored.Status != TaskOutboxStatusPending {
		t.Fatalf("stored outbox status = %q, want %q", stored.Status, TaskOutboxStatusPending)
	}
}

func TestTaskOutboxRelayDispatchesPendingMessage(t *testing.T) {
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
		Scopes:        []string{KeywordScopeTitle},
	})
	if err != nil {
		t.Fatalf("Create keyword error = %v", err)
	}

	task, outbox, err := taskService.CreateWithOutbox(ctx, CreateInspectionTaskInput{
		OrgID:        100,
		KeywordIDs:   []uint64{keyword.ID},
		IncludeBody:  true,
		ArticleState: ArticleStateOnline,
	})
	if err != nil {
		t.Fatalf("CreateWithOutbox() error = %v", err)
	}

	dispatcher := &articleInspectTaskDispatcherStub{}
	relay := NewTaskOutboxRelay(db, dispatcher, nil)
	if err := relay.DispatchMessage(context.Background(), outbox.ID); err != nil {
		t.Fatalf("DispatchMessage() error = %v", err)
	}
	if len(dispatcher.payloads) != 1 {
		t.Fatalf("dispatcher payloads len = %d, want %d", len(dispatcher.payloads), 1)
	}
	if dispatcher.payloads[0].TaskID != task.ID || dispatcher.payloads[0].OrgID != task.OrgID {
		t.Fatalf("dispatcher payload = %+v, want task=%d org=%d", dispatcher.payloads[0], task.ID, task.OrgID)
	}

	var stored InspectionTaskOutboxMessage
	if err := db.First(&stored, outbox.ID).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	if stored.Status != TaskOutboxStatusDispatched {
		t.Fatalf("outbox status = %q, want %q", stored.Status, TaskOutboxStatusDispatched)
	}
	if stored.AttemptCount != 1 {
		t.Fatalf("outbox attempt_count = %d, want %d", stored.AttemptCount, 1)
	}
	if stored.DispatchedAt == nil {
		t.Fatal("outbox dispatched_at = nil, want timestamp")
	}
}

func TestTaskOutboxRelayReclaimsExpiredLease(t *testing.T) {
	db := newArticleInspectTestDB(t)
	payload, err := json.Marshal(queuetasks.ArticleInspectTaskPayload{
		TaskID:        88,
		OrgID:         100,
		TriggerSource: "api",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	expiredAt := mustTime(t, "2026-04-20T13:00:00Z")
	claimedAt := expiredAt.Add(-time.Minute)
	if err := db.Exec(
		`INSERT INTO xt_article_inspect_task_outbox
		(orgid, task_id, message_type, status, payload, claimed_by, claimed_at, claim_until, create_at, update_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		100, 88, TaskOutboxMessageTypeRunTask, "claimed", string(payload), "scheduler@test", claimedAt, expiredAt, claimedAt, claimedAt,
	).Error; err != nil {
		t.Fatalf("insert claimed outbox row: %v", err)
	}

	dispatcher := &articleInspectTaskDispatcherStub{}
	relay := NewTaskOutboxRelay(db, dispatcher, nil)

	report, err := relay.DispatchPending(context.Background(), 1)
	if err != nil {
		t.Fatalf("DispatchPending() error = %v", err)
	}
	if report.Dispatched != 1 {
		t.Fatalf("DispatchPending().Dispatched = %d, want %d", report.Dispatched, 1)
	}
	if len(dispatcher.payloads) != 1 {
		t.Fatalf("dispatcher payload count = %d, want %d", len(dispatcher.payloads), 1)
	}

	row := loadOutboxPhase3Row(t, db, 1)
	if row.Status != TaskOutboxStatusDispatched {
		t.Fatalf("reclaimed outbox status = %q, want %q", row.Status, TaskOutboxStatusDispatched)
	}
	if !row.DispatchedAt.Valid {
		t.Fatal("reclaimed outbox dispatched_at = NULL, want timestamp")
	}
}

func TestTaskOutboxRelayRetryableFailureSchedulesNextAttempt(t *testing.T) {
	db := newArticleInspectTestDB(t)
	payload, err := json.Marshal(queuetasks.ArticleInspectTaskPayload{
		TaskID:        99,
		OrgID:         100,
		TriggerSource: "api",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	message := InspectionTaskOutboxMessage{
		OrgID:       100,
		TaskID:      99,
		MessageType: TaskOutboxMessageTypeRunTask,
		Status:      TaskOutboxStatusPending,
		Payload:     string(payload),
	}
	if err := db.Create(&message).Error; err != nil {
		t.Fatalf("create outbox row: %v", err)
	}

	dispatcher := &articleInspectTaskDispatcherStub{err: errors.New("queue down")}
	relay := NewTaskOutboxRelay(db, dispatcher, nil)

	report, err := relay.DispatchPending(context.Background(), 1)
	if err != nil {
		t.Fatalf("DispatchPending() error = %v", err)
	}
	if report.Failed != 1 {
		t.Fatalf("DispatchPending().Failed = %d, want %d", report.Failed, 1)
	}

	row := loadOutboxPhase3Row(t, db, message.ID)
	if row.Status != TaskOutboxStatusPending {
		t.Fatalf("retryable failure status = %q, want %q", row.Status, TaskOutboxStatusPending)
	}
	if !row.NextAttemptAt.Valid {
		t.Fatal("retryable failure next_attempt_at = NULL, want timestamp")
	}
	if row.LastErrorCode.String != "dispatch_error" {
		t.Fatalf("retryable failure last_error_code = %q, want %q", row.LastErrorCode.String, "dispatch_error")
	}
}

func TestTaskOutboxRelayMovesPoisonMessageToDeadLetter(t *testing.T) {
	db := newArticleInspectTestDB(t)
	message := InspectionTaskOutboxMessage{
		OrgID:       100,
		TaskID:      77,
		MessageType: TaskOutboxMessageTypeRunTask,
		Status:      TaskOutboxStatusPending,
		Payload:     "{bad-json",
	}
	if err := db.Create(&message).Error; err != nil {
		t.Fatalf("create poison outbox row: %v", err)
	}

	dispatcher := &articleInspectTaskDispatcherStub{}
	relay := NewTaskOutboxRelay(db, dispatcher, nil)

	report, err := relay.DispatchPending(context.Background(), 1)
	if err != nil {
		t.Fatalf("DispatchPending() error = %v", err)
	}
	if report.Failed != 1 {
		t.Fatalf("DispatchPending().Failed = %d, want %d", report.Failed, 1)
	}
	if len(dispatcher.payloads) != 0 {
		t.Fatalf("dispatcher payload count = %d, want %d", len(dispatcher.payloads), 0)
	}

	row := loadOutboxPhase3Row(t, db, message.ID)
	if row.Status != "dead_letter" {
		t.Fatalf("poison message status = %q, want %q", row.Status, "dead_letter")
	}
	if !row.DeadLetteredAt.Valid {
		t.Fatal("poison message dead_lettered_at = NULL, want timestamp")
	}
	if row.LastErrorCode.String != "payload_decode_error" {
		t.Fatalf("poison message last_error_code = %q, want %q", row.LastErrorCode.String, "payload_decode_error")
	}
}

func TestTaskOutboxRelayImplementsCleanerContract(t *testing.T) {
	db := newArticleInspectTestDB(t)
	relay := NewTaskOutboxRelay(db, nil, nil)
	if relay == nil {
		t.Fatal("NewTaskOutboxRelay() = nil, want relay")
	}

	type outboxCleaner interface {
		CleanupArticleInspectTaskOutbox(ctx context.Context, limit int) (int, error)
	}

	if _, ok := any(relay).(outboxCleaner); !ok {
		t.Fatal("TaskOutboxRelay does not implement CleanupArticleInspectTaskOutbox")
	}
}

func TestTaskOutboxRelayDeadLettersPoisonRowWithoutBlockingLaterMessages(t *testing.T) {
	db := newArticleInspectTestDB(t)
	dispatcher := &articleInspectTaskDispatcherStub{}
	relay := NewTaskOutboxRelay(db, dispatcher, nil)

	bad := InspectionTaskOutboxMessage{
		OrgID:       100,
		TaskID:      1,
		MessageType: TaskOutboxMessageTypeRunTask,
		Status:      TaskOutboxStatusPending,
		Payload:     "{not-json",
	}
	if err := db.Create(&bad).Error; err != nil {
		t.Fatalf("create bad outbox row: %v", err)
	}

	goodPayload, err := json.Marshal(queuetasks.ArticleInspectTaskPayload{
		TaskID:        2,
		OrgID:         100,
		TriggerSource: "api",
	})
	if err != nil {
		t.Fatalf("marshal good payload: %v", err)
	}
	good := InspectionTaskOutboxMessage{
		OrgID:       100,
		TaskID:      2,
		MessageType: TaskOutboxMessageTypeRunTask,
		Status:      TaskOutboxStatusPending,
		Payload:     string(goodPayload),
	}
	if err := db.Create(&good).Error; err != nil {
		t.Fatalf("create good outbox row: %v", err)
	}

	first, err := relay.DispatchPending(context.Background(), 1)
	if err != nil {
		t.Fatalf("DispatchPending(first) error = %v", err)
	}
	if first.Scanned != 1 || first.Failed != 1 || first.Dispatched != 0 {
		t.Fatalf("first DispatchPending() report = %+v, want scanned=1 failed=1 dispatched=0", first)
	}

	var badStored InspectionTaskOutboxMessage
	if err := db.First(&badStored, bad.ID).Error; err != nil {
		t.Fatalf("load bad outbox row: %v", err)
	}
	badPhase3 := loadOutboxPhase3Row(t, db, bad.ID)
	if badStored.Status != TaskOutboxStatusDeadLetter {
		t.Fatalf("bad outbox status = %q, want %q", badStored.Status, TaskOutboxStatusDeadLetter)
	}
	if badPhase3.AttemptCount != 1 {
		t.Fatalf("bad outbox attempt_count = %d, want %d", badPhase3.AttemptCount, 1)
	}
	if !badPhase3.DeadLetteredAt.Valid {
		t.Fatal("bad outbox dead_lettered_at = NULL, want timestamp")
	}
	if badPhase3.LastErrorCode.String != TaskOutboxErrorPayloadDecode {
		t.Fatalf("bad outbox last_error_code = %q, want %q", badPhase3.LastErrorCode.String, TaskOutboxErrorPayloadDecode)
	}

	second, err := relay.DispatchPending(context.Background(), 1)
	if err != nil {
		t.Fatalf("DispatchPending(second) error = %v", err)
	}
	if second.Scanned != 1 || second.Dispatched != 1 || second.Failed != 0 {
		t.Fatalf("second DispatchPending() report = %+v, want scanned=1 dispatched=1 failed=0", second)
	}
	if len(dispatcher.payloads) != 1 {
		t.Fatalf("dispatcher payloads len after second relay = %d, want %d", len(dispatcher.payloads), 1)
	}
	if dispatcher.payloads[0].TaskID != 2 {
		t.Fatalf("dispatcher payload task id = %d, want %d", dispatcher.payloads[0].TaskID, 2)
	}

	var goodStored InspectionTaskOutboxMessage
	if err := db.First(&goodStored, good.ID).Error; err != nil {
		t.Fatalf("load good outbox row: %v", err)
	}
	if goodStored.Status != TaskOutboxStatusDispatched {
		t.Fatalf("good outbox status = %q, want %q", goodStored.Status, TaskOutboxStatusDispatched)
	}
}

func TestTaskDelete(t *testing.T) {
	db := newArticleInspectTestDB(t)
	service := NewTaskService(db, NewKeywordRepository(db), NewArticleRepository(db))

	t.Run("deletes pending task and dependent rows", func(t *testing.T) {
		task := seedTaskForDeletion(t, db, 100, 901, TaskStatusPending)

		if err := service.Delete(context.Background(), 100, task.ID); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		assertTaskOwnedRowsDeleted(t, db, 100, task.ID)
	})

	t.Run("deletes failed task and dependent rows", func(t *testing.T) {
		task := seedTaskForDeletion(t, db, 100, 902, TaskStatusFailed)

		if err := service.Delete(context.Background(), 100, task.ID); err != nil {
			t.Fatalf("Delete(failed) error = %v", err)
		}

		assertTaskOwnedRowsDeleted(t, db, 100, task.ID)
	})

	for index, status := range []string{TaskStatusRunning, TaskStatusSuccess, TaskStatusPartialSuccess} {
		t.Run("rejects "+status+" task deletion", func(t *testing.T) {
			task := seedTaskForDeletion(t, db, 100, uint64(1001+index), status)

			err := service.Delete(context.Background(), 100, task.ID)
			if !errors.Is(err, ErrTaskDeleteNotAllowed) {
				t.Fatalf("Delete(%s) error = %v, want %v", status, err, ErrTaskDeleteNotAllowed)
			}

			var count int64
			if err := db.Model(&InspectionTask{}).Where("orgid = ? AND id = ?", 100, task.ID).Count(&count).Error; err != nil {
				t.Fatalf("count task error = %v", err)
			}
			if count != 1 {
				t.Fatalf("task count after rejected delete = %d, want %d", count, 1)
			}
		})
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
		&InspectionTaskOutboxMessage{},
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
	if _, ok := listedTask["created_at"]; !ok {
		t.Fatalf("listed task keys = %#v, want created_at", listedTask)
	}
	if _, ok := listedTask["create_at"]; ok {
		t.Fatalf("listed task keys = %#v, do not want create_at", listedTask)
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
	deletedTask := sendArticleInspectRequest(t, handler, http.MethodDelete, "/api/v1/article-inspect/tasks/"+strconv.FormatUint(taskID, 10)+"?orgid=100", nil)
	if deletedTask.status != http.StatusOK {
		t.Fatalf("delete task status = %d, want %d", deletedTask.status, http.StatusOK)
	}

	listedResults := sendArticleInspectRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/results?orgid=100&task_id=501&risk_level=high&page=1&page_size=20", nil)
	if listedResults.status != http.StatusOK {
		t.Fatalf("list results status = %d, want %d", listedResults.status, http.StatusOK)
	}
	resultListData := articleInspectDataMap(t, listedResults.envelope.Data)
	if total := articleInspectNumberField(t, resultListData, "total"); total != 1 {
		t.Fatalf("list results total = %v, want %d", total, 1)
	}
	resultItems := articleInspectListField(t, resultListData, "items")
	if len(resultItems) != 1 {
		t.Fatalf("result list items len = %d, want %d", len(resultItems), 1)
	}
	firstResult := articleInspectDataMap(t, resultItems[0])
	if articleInspectStringField(t, firstResult, "preview_field_name") != KeywordScopeTitle {
		t.Fatalf("preview_field_name = %q, want %q", articleInspectStringField(t, firstResult, "preview_field_name"), KeywordScopeTitle)
	}
	if articleInspectStringField(t, firstResult, "preview_keyword_text") != "alpha" {
		t.Fatalf("preview_keyword_text = %q, want %q", articleInspectStringField(t, firstResult, "preview_keyword_text"), "alpha")
	}
	if articleInspectStringField(t, firstResult, "preview_matched_text") != "Alpha" {
		t.Fatalf("preview_matched_text = %q, want %q", articleInspectStringField(t, firstResult, "preview_matched_text"), "Alpha")
	}
	if articleInspectStringField(t, firstResult, "preview_snippet") != "Alpha news" {
		t.Fatalf("preview_snippet = %q, want %q", articleInspectStringField(t, firstResult, "preview_snippet"), "Alpha news")
	}
	if articleInspectUint64Field(t, firstResult, "extra_hit_count") != 1 {
		t.Fatalf("extra_hit_count = %d, want %d", articleInspectUint64Field(t, firstResult, "extra_hit_count"), 1)
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
		{name: "batch offline", path: "/api/v1/article-inspect/actions/batch-offline"},
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

func TestHandlerTenantScopedRoutesRequireSession(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedOrgCategoryFixtures(t, db)
	seedLifecycleArticles(t, db)
	handler := newArticleInspectHandler(t, db, &articleInspectTaskDispatcherStub{})

	tests := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{
			name:   "category list",
			method: http.MethodGet,
			path:   "/api/v1/article-inspect/categories?page=1&page_size=20",
		},
		{
			name:   "article rectify",
			method: http.MethodPut,
			path:   "/api/v1/article-inspect/articles/12/rectify",
			body: map[string]any{
				"title": "Updated title",
				"body":  "Updated body content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result articleInspectHTTPResult
			if tt.body != nil {
				result = sendArticleInspectJSONRequestWithOptions(t, handler, tt.method, tt.path, tt.body, articleInspectRequestOptions{Anonymous: true})
			} else {
				result = sendArticleInspectRequestWithOptions(t, handler, tt.method, tt.path, nil, articleInspectRequestOptions{Anonymous: true})
			}
			if result.status != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", result.status, http.StatusUnauthorized)
			}
			if result.envelope.Code != http.StatusUnauthorized {
				t.Fatalf("envelope = %+v, want unauthorized", result.envelope)
			}
		})
	}
}

func TestArticleInspectRoutesPreferSessionOrgIDOverRequestOrgID(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedOrgCategoryFixtures(t, db)
	seedQueryFixtures(t, db)
	seedLifecycleArticles(t, db)
	seedArticleCenterFixtures(t, db)
	dispatcher := &articleInspectTaskDispatcherStub{}
	handler := newArticleInspectHandler(t, db, dispatcher)
	if err := db.Create(&InspectionResult{
		ID:                3001,
		OrgID:             100,
		TaskID:            501,
		ArticleID:         10,
		ArticleState:      ArticleStateOnline,
		DispositionStatus: ResultDispositionPending,
	}).Error; err != nil {
		t.Fatalf("seed session-scoped batch result error = %v", err)
	}

	session100 := articleInspectRequestOptions{
		Session: &identity.AdminSession{
			UserID:   7,
			OrgID:    100,
			OrgName:  "测试组织A",
			Nickname: "alice",
			Priv:     "admin",
			Status:   "active",
		},
	}
	session29 := articleInspectRequestOptions{
		Session: &identity.AdminSession{
			UserID:   7,
			OrgID:    29,
			OrgName:  "一县一端",
			Nickname: "alice",
			Priv:     "admin",
			Status:   "active",
		},
	}

	createdCategory := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPost, "/api/v1/article-inspect/categories", map[string]any{
		"orgid":   100,
		"name":    "会话机构分类",
		"enabled": true,
		"sort":    15,
	}, session29)
	if createdCategory.status != http.StatusCreated {
		t.Fatalf("create category status = %d, want %d", createdCategory.status, http.StatusCreated)
	}
	categoryData := articleInspectDataMap(t, createdCategory.envelope.Data)
	if articleInspectUint64Field(t, categoryData, "orgid") != 29 {
		t.Fatalf("created category orgid = %d, want session org %d", articleInspectUint64Field(t, categoryData, "orgid"), 29)
	}

	createdKeyword := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPost, "/api/v1/article-inspect/keywords", map[string]any{
		"orgid":          200,
		"name":           "session-keyword",
		"category_id":    1001,
		"match_type":     MatchTypeContains,
		"risk_level":     RiskLevelHigh,
		"suggest_action": SuggestActionOffline,
		"enabled":        true,
		"scopes":         []string{KeywordScopeTitle},
	}, session100)
	if createdKeyword.status != http.StatusCreated {
		t.Fatalf("create keyword status = %d, want %d", createdKeyword.status, http.StatusCreated)
	}
	createdKeywordData := articleInspectDataMap(t, createdKeyword.envelope.Data)
	keywordID := articleInspectUint64Field(t, createdKeywordData, "id")
	if articleInspectUint64Field(t, createdKeywordData, "orgid") != 100 {
		t.Fatalf("created keyword orgid = %d, want session org %d", articleInspectUint64Field(t, createdKeywordData, "orgid"), 100)
	}

	listedKeywords := sendArticleInspectRequestWithOptions(t, handler, http.MethodGet, "/api/v1/article-inspect/keywords?orgid=200&page=1&page_size=20", nil, session100)
	if listedKeywords.status != http.StatusOK {
		t.Fatalf("list keywords status = %d, want %d", listedKeywords.status, http.StatusOK)
	}
	keywordListData := articleInspectDataMap(t, listedKeywords.envelope.Data)
	keywordItems := articleInspectListField(t, keywordListData, "items")
	if len(keywordItems) != 1 {
		t.Fatalf("keyword list items len = %d, want %d", len(keywordItems), 1)
	}
	keywordItem := articleInspectDataMap(t, keywordItems[0])
	if articleInspectUint64Field(t, keywordItem, "orgid") != 100 {
		t.Fatalf("listed keyword orgid = %d, want session org %d", articleInspectUint64Field(t, keywordItem, "orgid"), 100)
	}

	createdTask := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPost, "/api/v1/article-inspect/tasks", map[string]any{
		"orgid":         200,
		"keyword_ids":   []uint64{keywordID},
		"include_body":  true,
		"article_state": ArticleStateOnline,
	}, session100)
	if createdTask.status != http.StatusCreated {
		t.Fatalf("create task status = %d, want %d", createdTask.status, http.StatusCreated)
	}
	taskData := articleInspectDataMap(t, createdTask.envelope.Data)
	taskID := articleInspectUint64Field(t, taskData, "id")
	if len(dispatcher.payloads) != 1 || dispatcher.payloads[0].OrgID != 100 {
		t.Fatalf("dispatcher payloads = %+v, want session org 100", dispatcher.payloads)
	}

	detailTask := sendArticleInspectRequestWithOptions(t, handler, http.MethodGet, "/api/v1/article-inspect/tasks/"+strconv.FormatUint(taskID, 10)+"?orgid=200", nil, session100)
	if detailTask.status != http.StatusOK {
		t.Fatalf("task detail status = %d, want %d", detailTask.status, http.StatusOK)
	}
	detailTaskData := articleInspectDataMap(t, detailTask.envelope.Data)
	if articleInspectUint64Field(t, detailTaskData, "orgid") != 100 {
		t.Fatalf("task detail orgid = %d, want session org %d", articleInspectUint64Field(t, detailTaskData, "orgid"), 100)
	}

	detailResult := sendArticleInspectRequestWithOptions(t, handler, http.MethodGet, "/api/v1/article-inspect/results/1001?orgid=200", nil, session100)
	if detailResult.status != http.StatusOK {
		t.Fatalf("result detail status = %d, want %d", detailResult.status, http.StatusOK)
	}
	detailResultData := articleInspectDataMap(t, detailResult.envelope.Data)
	resultRecord := articleInspectDataMap(t, detailResultData["result"])
	if articleInspectUint64Field(t, resultRecord, "orgid") != 100 {
		t.Fatalf("result detail orgid = %d, want session org %d", articleInspectUint64Field(t, resultRecord, "orgid"), 100)
	}

	batchOffline := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPost, "/api/v1/article-inspect/actions/batch-offline", map[string]any{
		"orgid":      200,
		"task_id":    501,
		"result_ids": []uint64{3001},
		"reason":     "session org should win",
	}, session100)
	if batchOffline.status != http.StatusOK {
		t.Fatalf("batch offline status = %d, want %d", batchOffline.status, http.StatusOK)
	}

	rectify := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPut, "/api/v1/article-inspect/articles/12/rectify", map[string]any{
		"orgid": 200,
		"title": "Session updated title",
		"body":  "Session updated body",
	}, session100)
	if rectify.status != http.StatusOK {
		t.Fatalf("rectify status = %d, want %d", rectify.status, http.StatusOK)
	}

	republish := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPost, "/api/v1/article-inspect/articles/11/republish", map[string]any{
		"orgid":  200,
		"reason": "session org should win",
	}, session100)
	if republish.status != http.StatusOK {
		t.Fatalf("republish status = %d, want %d", republish.status, http.StatusOK)
	}

	listLogs := sendArticleInspectRequestWithOptions(t, handler, http.MethodGet, "/api/v1/article-inspect/logs/operations?orgid=200&page=1&page_size=20", nil, session100)
	if listLogs.status != http.StatusOK {
		t.Fatalf("operation logs status = %d, want %d", listLogs.status, http.StatusOK)
	}
	logData := articleInspectDataMap(t, listLogs.envelope.Data)
	if total := articleInspectNumberField(t, logData, "total"); total == 0 {
		t.Fatalf("operation logs total = %v, want > 0", total)
	}

	listArticles := sendArticleInspectRequestWithOptions(t, handler, http.MethodGet, "/api/v1/article-inspect/articles?orgid=30&page=1&page_size=20", nil, session29)
	if listArticles.status != http.StatusOK {
		t.Fatalf("article list status = %d, want %d", listArticles.status, http.StatusOK)
	}
	articleData := articleInspectDataMap(t, listArticles.envelope.Data)
	articleItems := articleInspectListField(t, articleData, "items")
	if len(articleItems) == 0 {
		t.Fatal("article list items = empty, want session org records")
	}
	for _, raw := range articleItems {
		item := articleInspectDataMap(t, raw)
		if articleInspectUint64Field(t, item, "orgid") != 29 {
			t.Fatalf("article item orgid = %d, want session org %d", articleInspectUint64Field(t, item, "orgid"), 29)
		}
	}
}

func TestHandlerActionAndLifecycleResponsesUseSnakeCase(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedLifecycleArticles(t, db)
	seedBatchOfflineFixtures(t, db)
	handler := newArticleInspectHandler(t, db, &articleInspectTaskDispatcherStub{})

	batchOffline := sendArticleInspectJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/actions/batch-offline", map[string]any{
		"orgid":      100,
		"task_id":    501,
		"result_ids": []uint64{2001},
		"reason":     "manual batch offline",
	})
	if batchOffline.status != http.StatusOK {
		t.Fatalf("batch offline status = %d, want %d", batchOffline.status, http.StatusOK)
	}

	batchData := articleInspectDataMap(t, batchOffline.envelope.Data)
	if articleInspectUint64Field(t, batchData, "action_id") == 0 {
		t.Fatalf("batch action_id = 0, want non-zero")
	}
	if articleInspectUint64Field(t, batchData, "target_count") != 1 {
		t.Fatalf("batch target_count = %d, want %d", articleInspectUint64Field(t, batchData, "target_count"), 1)
	}
	if _, ok := batchData["ActionID"]; ok {
		t.Fatalf("batch action keys = %#v, do not want ActionID", batchData)
	}

	rectify := sendArticleInspectJSONRequest(t, handler, http.MethodPut, "/api/v1/article-inspect/articles/12/rectify", map[string]any{
		"orgid": 100,
		"title": "Updated title",
		"desc":  "Updated desc",
		"body":  "Updated body content",
	})
	if rectify.status != http.StatusOK {
		t.Fatalf("rectify status = %d, want %d", rectify.status, http.StatusOK)
	}

	rectifyChanges, ok := rectify.envelope.Data.([]any)
	if !ok || len(rectifyChanges) == 0 {
		t.Fatalf("rectify data = %#v, want non-empty []any", rectify.envelope.Data)
	}
	firstChange := articleInspectDataMap(t, rectifyChanges[0])
	if articleInspectStringField(t, firstChange, "field_name") == "" {
		t.Fatalf("rectify field_name = empty, want non-empty")
	}
	if _, ok := firstChange["FieldName"]; ok {
		t.Fatalf("rectify change keys = %#v, do not want FieldName", firstChange)
	}

	republish := sendArticleInspectJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/articles/11/republish", map[string]any{
		"orgid":  100,
		"reason": "send back to audit",
	})
	if republish.status != http.StatusOK {
		t.Fatalf("republish status = %d, want %d", republish.status, http.StatusOK)
	}

	republishData := articleInspectDataMap(t, republish.envelope.Data)
	if articleInspectUint64Field(t, republishData, "article_id") != 11 {
		t.Fatalf("republish article_id = %d, want %d", articleInspectUint64Field(t, republishData, "article_id"), 11)
	}
	if _, ok := republishData["ArticleID"]; ok {
		t.Fatalf("republish keys = %#v, do not want ArticleID", republishData)
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
		if articleInspectUint64Field(t, first, "cate_id") != 0 {
			t.Fatalf("first org cate_id = %d, want %d", articleInspectUint64Field(t, first, "cate_id"), 0)
		}
		if _, ok := first["cateid"]; ok {
			t.Fatalf("first org keys = %#v, do not want cateid", first)
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
			if _, ok := item["code"]; ok {
				t.Fatalf("category payload = %#v, do not want code", item)
			}
			if _, ok := item["created_at"]; !ok {
				t.Fatalf("category keys = %#v, want created_at", item)
			}
			if _, ok := item["create_at"]; ok {
				t.Fatalf("category keys = %#v, do not want create_at", item)
			}
			gotIDs = append(gotIDs, articleInspectUint64Field(t, item, "id"))
		}
		sort.Slice(gotIDs, func(i, j int) bool { return gotIDs[i] < gotIDs[j] })
		if !reflect.DeepEqual(gotIDs, []uint64{501, 502}) {
			t.Fatalf("category ids = %#v, want %#v", gotIDs, []uint64{501, 502})
		}
	})

	t.Run("category CRUD rejects anonymous request", func(t *testing.T) {
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
					result = sendArticleInspectJSONRequestWithOptions(t, handler, tt.method, tt.path, tt.body, articleInspectRequestOptions{Anonymous: true})
				} else {
					result = sendArticleInspectRequestWithOptions(t, handler, tt.method, tt.path, nil, articleInspectRequestOptions{Anonymous: true})
				}
				if result.status != http.StatusUnauthorized {
					t.Fatalf("%s status = %d, want %d", tt.name, result.status, http.StatusUnauthorized)
				}
				if result.envelope.Code != http.StatusUnauthorized {
					t.Fatalf("%s envelope = %+v, want unauthorized code", tt.name, result.envelope)
				}
			})
		}
	})

	t.Run("category create and update payloads do not use code", func(t *testing.T) {
		created := sendArticleInspectJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/categories", map[string]any{
			"orgid":   29,
			"name":    "新增分类",
			"enabled": true,
			"sort":    10,
		})
		if created.status != http.StatusCreated {
			t.Fatalf("create category status = %d, want %d", created.status, http.StatusCreated)
		}

		createdData := articleInspectDataMap(t, created.envelope.Data)
		if articleInspectStringField(t, createdData, "name") != "新增分类" {
			t.Fatalf("create category name = %q, want %q", articleInspectStringField(t, createdData, "name"), "新增分类")
		}
		if _, ok := createdData["code"]; ok {
			t.Fatalf("category payload = %#v, do not want code", createdData)
		}

		categoryID := articleInspectUint64String(t, createdData["id"])
		updated := sendArticleInspectJSONRequest(t, handler, http.MethodPut, "/api/v1/article-inspect/categories/"+categoryID, map[string]any{
			"orgid":   29,
			"name":    "新增分类-更新",
			"enabled": false,
			"sort":    15,
		})
		if updated.status != http.StatusOK {
			t.Fatalf("update category status = %d, want %d", updated.status, http.StatusOK)
		}

		updatedData := articleInspectDataMap(t, updated.envelope.Data)
		if articleInspectStringField(t, updatedData, "name") != "新增分类-更新" {
			t.Fatalf("update category name = %q, want %q", articleInspectStringField(t, updatedData, "name"), "新增分类-更新")
		}
		if _, ok := updatedData["code"]; ok {
			t.Fatalf("category payload = %#v, do not want code", updatedData)
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

	t.Run("article list endpoint filters by title like", func(t *testing.T) {
		result := sendArticleInspectRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/articles?orgid=29&page=1&page_size=20&title=%E8%A6%81%E9%97%BB%E4%B8%80", nil)
		if result.status != http.StatusOK {
			t.Fatalf("list articles by title status = %d, want %d", result.status, http.StatusOK)
		}

		data := articleInspectDataMap(t, result.envelope.Data)
		items := articleInspectListField(t, data, "items")
		if len(items) != 1 {
			t.Fatalf("title filtered items len = %d, want %d", len(items), 1)
		}

		item := articleInspectDataMap(t, items[0])
		if articleInspectUint64Field(t, item, "id") != 9001 {
			t.Fatalf("title filtered id = %d, want %d", articleInspectUint64Field(t, item, "id"), 9001)
		}
	})

	t.Run("article list endpoint filters by article id", func(t *testing.T) {
		result := sendArticleInspectRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/articles?orgid=29&page=1&page_size=20&article_id=9002", nil)
		if result.status != http.StatusOK {
			t.Fatalf("list articles by article id status = %d, want %d", result.status, http.StatusOK)
		}

		data := articleInspectDataMap(t, result.envelope.Data)
		items := articleInspectListField(t, data, "items")
		if len(items) != 1 {
			t.Fatalf("article id filtered items len = %d, want %d", len(items), 1)
		}

		item := articleInspectDataMap(t, items[0])
		if articleInspectUint64Field(t, item, "id") != 9002 {
			t.Fatalf("article id filtered id = %d, want %d", articleInspectUint64Field(t, item, "id"), 9002)
		}
	})

	t.Run("article list endpoint still supports exact state filtering", func(t *testing.T) {
		result := sendArticleInspectRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/articles?orgid=29&page=1&page_size=20&state=9", nil)
		if result.status != http.StatusOK {
			t.Fatalf("list articles by state status = %d, want %d", result.status, http.StatusOK)
		}

		data := articleInspectDataMap(t, result.envelope.Data)
		items := articleInspectListField(t, data, "items")
		if len(items) != 1 {
			t.Fatalf("state filtered items len = %d, want %d", len(items), 1)
		}

		item := articleInspectDataMap(t, items[0])
		if articleInspectUint64Field(t, item, "id") != 9001 {
			t.Fatalf("state filtered id = %d, want %d", articleInspectUint64Field(t, item, "id"), 9001)
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
	RegisterRoutes(api, NewRoutes(db, dispatcher))

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
		"/api/v1/article-inspect/actions/batch-offline",
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

	assertArticleInspectOperationID(t, doc.Paths, "/api/v1/article-inspect/categories", http.MethodPost, "article-inspect-category-create")
	assertArticleInspectOperationID(t, doc.Paths, "/api/v1/article-inspect/tasks", http.MethodPost, "article-inspect-task-create")
	assertArticleInspectOperationID(t, doc.Paths, "/api/v1/article-inspect/keywords/{id}", http.MethodGet, "article-inspect-keyword-detail")

	if !articleInspectHasResponseStatus(doc.Paths, "/api/v1/article-inspect/categories", http.MethodPost, "201") {
		t.Fatal("category create must document 201 response")
	}
	if !articleInspectHasResponseStatus(doc.Paths, "/api/v1/article-inspect/tasks", http.MethodPost, "201") {
		t.Fatal("task create must document 201 response")
	}

	if got := articleInspectParameterSchemaType(t, doc.Paths, "/api/v1/article-inspect/keywords/{id}", http.MethodGet, "id"); got != "integer" {
		t.Fatalf("keyword detail path id schema type = %q, want %q", got, "integer")
	}
	if got := articleInspectParameterSchemaType(t, doc.Paths, "/api/v1/article-inspect/categories", http.MethodGet, "enabled"); got != "boolean" {
		t.Fatalf("category list enabled schema type = %q, want %q", got, "boolean")
	}
	if got := articleInspectParameterSchemaType(t, doc.Paths, "/api/v1/article-inspect/results", http.MethodGet, "orgid"); got != "integer" {
		t.Fatalf("result list orgid schema type = %q, want %q", got, "integer")
	}
}

func TestNewRoutesBuildsModuleDependencies(t *testing.T) {
	db := newArticleInspectTestDB(t)
	dispatcher := &articleInspectTaskDispatcherStub{}

	routes := NewRoutes(db, dispatcher)
	if routes.Categories == nil {
		t.Fatal("NewRoutes().Categories = nil")
	}
	if routes.Keywords == nil {
		t.Fatal("NewRoutes().Keywords = nil")
	}
	if routes.Tasks == nil {
		t.Fatal("NewRoutes().Tasks = nil")
	}
	if routes.Results == nil {
		t.Fatal("NewRoutes().Results = nil")
	}
	if routes.Actions == nil {
		t.Fatal("NewRoutes().Actions = nil")
	}
	if routes.Lifecycle == nil {
		t.Fatal("NewRoutes().Lifecycle = nil")
	}
	if routes.Logs == nil {
		t.Fatal("NewRoutes().Logs = nil")
	}
	if routes.Articles == nil {
		t.Fatal("NewRoutes().Articles = nil")
	}
	if routes.Dispatcher != dispatcher {
		t.Fatalf("NewRoutes().Dispatcher = %#v, want %#v", routes.Dispatcher, dispatcher)
	}
}

func TestHandlerInvalidRouteInputsStillUseEnvelopeBadRequest(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedOrgCategoryFixtures(t, db)
	seedArticleCenterFixtures(t, db)
	handler := newArticleInspectHandler(t, db, &articleInspectTaskDispatcherStub{})

	tests := []struct {
		name string
		path string
	}{
		{name: "invalid category id", path: "/api/v1/article-inspect/categories/not-a-number?orgid=29"},
		{name: "invalid keyword enabled", path: "/api/v1/article-inspect/keywords?orgid=29&page=1&page_size=20&enabled=not-bool"},
		{name: "invalid article state", path: "/api/v1/article-inspect/articles?orgid=29&page=1&page_size=20&state=oops"},
		{name: "invalid log start_at", path: "/api/v1/article-inspect/logs/operations?orgid=29&page=1&page_size=20&start_at=bad-time"},
		{name: "invalid result orgid", path: "/api/v1/article-inspect/results?orgid=abc&page=1&page_size=20"},
		{name: "invalid task page", path: "/api/v1/article-inspect/tasks?orgid=100&page=bad&page_size=20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sendArticleInspectRequest(t, handler, http.MethodGet, tt.path, nil)
			if result.status != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", result.status, http.StatusBadRequest)
			}
			if result.envelope.Code != http.StatusBadRequest {
				t.Fatalf("envelope = %+v, want bad request envelope", result.envelope)
			}
		})
	}
}

func TestTaskCreateEnqueueFailureLeavesPendingOutbox(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedOrgCategoryFixtures(t, db)
	dispatcher := &articleInspectTaskDispatcherStub{err: errors.New("queue down")}
	handler := newArticleInspectHandler(t, db, dispatcher)

	keywordService := NewKeywordService(NewKeywordRepository(db))
	createdKeyword, err := keywordService.Create(context.Background(), CreateKeywordInput{
		OrgID:         100,
		Name:          "spam",
		CategoryID:    1001,
		MatchType:     MatchTypeContains,
		RiskLevel:     RiskLevelHigh,
		SuggestAction: SuggestActionOffline,
		Enabled:       true,
		Scopes:        []string{KeywordScopeTitle},
	})
	if err != nil {
		t.Fatalf("create keyword fixture: %v", err)
	}

	result := sendArticleInspectJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/tasks", map[string]any{
		"orgid":         100,
		"keyword_ids":   []uint64{createdKeyword.ID},
		"include_body":  true,
		"article_state": ArticleStateOnline,
	})
	if result.status != http.StatusCreated {
		t.Fatalf("create task status = %d, want %d", result.status, http.StatusCreated)
	}
	if result.envelope.Code != 0 {
		t.Fatalf("create task envelope = %+v, want success code", result.envelope)
	}

	var taskCount int64
	if err := db.Model(&InspectionTask{}).Where("orgid = ?", 100).Count(&taskCount).Error; err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	if taskCount != 1 {
		t.Fatalf("task count after enqueue failure = %d, want %d", taskCount, 1)
	}

	var taskKeywordCount int64
	if err := db.Model(&InspectionTaskKeyword{}).Where("orgid = ?", 100).Count(&taskKeywordCount).Error; err != nil {
		t.Fatalf("count task keywords: %v", err)
	}
	if taskKeywordCount != 1 {
		t.Fatalf("task keyword count after enqueue failure = %d, want %d", taskKeywordCount, 1)
	}

	var outbox InspectionTaskOutboxMessage
	if err := db.Where("orgid = ?", 100).First(&outbox).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	if outbox.Status != TaskOutboxStatusPending {
		t.Fatalf("outbox status after enqueue failure = %q, want %q", outbox.Status, TaskOutboxStatusPending)
	}
	if outbox.AttemptCount != 1 {
		t.Fatalf("outbox attempt_count after enqueue failure = %d, want %d", outbox.AttemptCount, 1)
	}
	if !strings.Contains(outbox.LastError, "queue down") {
		t.Fatalf("outbox last_error = %q, want contains %q", outbox.LastError, "queue down")
	}
}

func TestTaskCreateWithoutDispatcherStillCreatesPendingOutbox(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedOrgCategoryFixtures(t, db)
	handler := newArticleInspectHandler(t, db, nil)

	keywordService := NewKeywordService(NewKeywordRepository(db))
	createdKeyword, err := keywordService.Create(context.Background(), CreateKeywordInput{
		OrgID:         100,
		Name:          "spam",
		CategoryID:    1001,
		MatchType:     MatchTypeContains,
		RiskLevel:     RiskLevelHigh,
		SuggestAction: SuggestActionOffline,
		Enabled:       true,
		Scopes:        []string{KeywordScopeTitle},
	})
	if err != nil {
		t.Fatalf("create keyword fixture: %v", err)
	}

	result := sendArticleInspectJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/tasks", map[string]any{
		"orgid":         100,
		"keyword_ids":   []uint64{createdKeyword.ID},
		"include_body":  true,
		"article_state": ArticleStateOnline,
	})
	if result.status != http.StatusCreated {
		t.Fatalf("create task status = %d, want %d", result.status, http.StatusCreated)
	}

	var taskCount int64
	if err := db.Model(&InspectionTask{}).Where("orgid = ?", 100).Count(&taskCount).Error; err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	if taskCount != 1 {
		t.Fatalf("task count without dispatcher = %d, want %d", taskCount, 1)
	}

	var outbox InspectionTaskOutboxMessage
	if err := db.Where("orgid = ?", 100).First(&outbox).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	if outbox.Status != TaskOutboxStatusPending {
		t.Fatalf("outbox status without dispatcher = %q, want %q", outbox.Status, TaskOutboxStatusPending)
	}
	if outbox.AttemptCount != 0 {
		t.Fatalf("outbox attempt_count without dispatcher = %d, want %d", outbox.AttemptCount, 0)
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

type articleInspectRequestOptions struct {
	Anonymous bool
	Session   *identity.AdminSession
}

type articleInspectTaskDispatcherStub struct {
	payloads []queuetasks.ArticleInspectTaskPayload
	err      error
}

type outboxPhase3Row struct {
	Status         string
	AttemptCount   int64
	ClaimedBy      sql.NullString
	ClaimedAt      sql.NullTime
	ClaimUntil     sql.NullTime
	NextAttemptAt  sql.NullTime
	LastErrorCode  sql.NullString
	DeadLetteredAt sql.NullTime
	RetainedUntil  sql.NullTime
	DispatchedAt   sql.NullTime
}

func (s *articleInspectTaskDispatcherStub) DispatchArticleInspectTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) error {
	_ = ctx
	s.payloads = append(s.payloads, payload)
	return s.err
}

func newArticleInspectHandler(t *testing.T, db *gorm.DB, dispatcher TaskDispatcher) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterRoutes(api, NewRoutes(db, dispatcher))
	return mux
}

func sendArticleInspectJSONRequest(t *testing.T, handler http.Handler, method, path string, body any) articleInspectHTTPResult {
	t.Helper()

	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	return sendArticleInspectRequestWithOptions(t, handler, method, path, bytes.NewReader(encoded), articleInspectRequestOptions{})
}

func sendArticleInspectJSONRequestWithOptions(t *testing.T, handler http.Handler, method, path string, body any, options articleInspectRequestOptions) articleInspectHTTPResult {
	t.Helper()

	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	return sendArticleInspectRequestWithOptions(t, handler, method, path, bytes.NewReader(encoded), options)
}

func sendArticleInspectRequest(t *testing.T, handler http.Handler, method, path string, body *bytes.Reader) articleInspectHTTPResult {
	t.Helper()
	return sendArticleInspectRequestWithOptions(t, handler, method, path, body, articleInspectRequestOptions{})
}

func sendArticleInspectRequestWithOptions(t *testing.T, handler http.Handler, method, path string, body *bytes.Reader, options articleInspectRequestOptions) articleInspectHTTPResult {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = io.ReadAll(body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
	}

	var req *http.Request
	if payload != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	if !options.Anonymous {
		session := options.Session
		if session == nil {
			derived := deriveArticleInspectSession(t, path, payload)
			session = &derived
		}
		actor := session.Actor()
		ctx := identity.ContextWithAdminSession(req.Context(), *session)
		ctx = identity.ContextWithActor(ctx, actor)
		ctx = identity.ContextWithPrincipal(ctx, identity.PrincipalFromActor(actor))
		req = req.WithContext(ctx)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return articleInspectHTTPResult{status: rec.Code, envelope: decodeArticleInspectEnvelope(t, rec)}
}

func deriveArticleInspectSession(t *testing.T, requestPath string, payload []byte) identity.AdminSession {
	t.Helper()

	orgID := uint64(29)
	if parsedURL, err := url.Parse(requestPath); err == nil {
		if raw := strings.TrimSpace(parsedURL.Query().Get("orgid")); raw != "" {
			if parsed, parseErr := strconv.ParseUint(raw, 10, 64); parseErr == nil && parsed > 0 {
				orgID = parsed
			}
		}
	}

	if len(payload) > 0 {
		var body map[string]any
		if err := json.Unmarshal(payload, &body); err == nil {
			if parsed := articleInspectBodyOrgID(body); parsed > 0 {
				orgID = parsed
			}
		}
	}

	return identity.AdminSession{
		UserID:   7,
		OrgID:    orgID,
		OrgName:  fmt.Sprintf("org-%d", orgID),
		Nickname: "alice",
		Priv:     "admin",
		Status:   "active",
	}
}

func articleInspectBodyOrgID(body map[string]any) uint64 {
	value, ok := body["orgid"]
	if !ok {
		return 0
	}

	switch typed := value.(type) {
	case float64:
		return uint64(typed)
	case string:
		parsed, err := strconv.ParseUint(strings.TrimSpace(typed), 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
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

func loadOutboxPhase3Row(t *testing.T, db *gorm.DB, id uint64) outboxPhase3Row {
	t.Helper()

	var row outboxPhase3Row
	if err := db.Raw(
		`SELECT status, attempt_count, claimed_by, claimed_at, claim_until, next_attempt_at,
		        last_error_code, dead_lettered_at, retained_until, dispatched_at
		   FROM xt_article_inspect_task_outbox
		  WHERE id = ?`,
		id,
	).Scan(&row).Error; err != nil {
		t.Fatalf("load outbox phase 3 row: %v", err)
	}
	return row
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

func assertArticleInspectOperationID(t *testing.T, paths map[string]map[string]any, path, method, want string) {
	t.Helper()
	got := articleInspectOperationFieldAsString(t, paths, path, method, "operationId")
	if got != want {
		t.Fatalf("%s %s operationId = %q, want %q", method, path, got, want)
	}
}

func articleInspectHasResponseStatus(paths map[string]map[string]any, path, method, status string) bool {
	operation, ok := paths[path][strings.ToLower(method)].(map[string]any)
	if !ok {
		return false
	}
	responses, ok := operation["responses"].(map[string]any)
	if !ok {
		return false
	}
	_, ok = responses[status]
	return ok
}

func articleInspectParameterSchemaType(t *testing.T, paths map[string]map[string]any, path, method, name string) string {
	t.Helper()
	operation := articleInspectOperationMap(t, paths, path, method)
	params, ok := operation["parameters"].([]any)
	if !ok {
		t.Fatalf("%s %s parameters type = %T, want []any", method, path, operation["parameters"])
	}
	for _, raw := range params {
		param, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("%s %s parameter item type = %T, want map[string]any", method, path, raw)
		}
		if paramName, _ := param["name"].(string); paramName == name {
			schema, ok := param["schema"].(map[string]any)
			if !ok {
				t.Fatalf("%s %s parameter %s schema type = %T, want map[string]any", method, path, name, param["schema"])
			}
			schemaType, _ := schema["type"].(string)
			return schemaType
		}
	}
	t.Fatalf("%s %s missing parameter %q", method, path, name)
	return ""
}

func articleInspectOperationFieldAsString(t *testing.T, paths map[string]map[string]any, path, method, field string) string {
	t.Helper()
	operation := articleInspectOperationMap(t, paths, path, method)
	value, ok := operation[field].(string)
	if !ok {
		t.Fatalf("%s %s field %q type = %T, want string", method, path, field, operation[field])
	}
	return value
}

func articleInspectOperationMap(t *testing.T, paths map[string]map[string]any, path, method string) map[string]any {
	t.Helper()
	item, ok := paths[path]
	if !ok {
		t.Fatalf("openapi missing path %s", path)
	}
	operation, ok := item[strings.ToLower(method)].(map[string]any)
	if !ok {
		t.Fatalf("openapi path %s missing method %s", path, method)
	}
	return operation
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

func seedTaskForDeletion(t *testing.T, db *gorm.DB, orgID, baseID uint64, status string) *InspectionTask {
	t.Helper()

	task := &InspectionTask{
		ID:                 baseID,
		OrgID:              orgID,
		TaskNo:             fmt.Sprintf("inspect-delete-%d", baseID),
		Status:             status,
		ArticleStateFilter: "9",
		RequestSnapshot:    "{}",
		RuleSnapshot:       "[]",
	}
	if err := db.Create(task).Error; err != nil {
		t.Fatalf("create deletion task error = %v", err)
	}

	taskKeywords := []InspectionTaskKeyword{
		{ID: baseID * 10, OrgID: orgID, TaskID: task.ID, KeywordID: baseID + 1},
	}
	results := []InspectionResult{
		{ID: baseID*10 + 1, OrgID: orgID, TaskID: task.ID, ArticleID: baseID + 100, ArticleTitle: "Delete me", DispositionStatus: ResultDispositionPending},
	}
	hits := []InspectionResultHit{
		{ID: baseID*10 + 2, OrgID: orgID, TaskID: task.ID, ResultID: results[0].ID, ArticleID: results[0].ArticleID, KeywordID: baseID + 1, KeywordText: "delete", FieldName: KeywordScopeTitle, MatchType: MatchTypeContains, RiskLevel: RiskLevelHigh, SuggestAction: SuggestActionOffline, Snippet: "delete snippet"},
	}
	actions := []InspectionAction{
		{ID: baseID*10 + 3, OrgID: orgID, ActionNo: fmt.Sprintf("act-%d", baseID), ActionType: ActionTypeBatchIgnore, TaskID: task.ID, Status: ActionStatusSuccess},
	}
	opLogs := []InspectionOperationLog{
		{ID: baseID*10 + 4, OrgID: orgID, ActionID: actions[0].ID, TaskID: task.ID, ResultID: results[0].ID, ArticleID: results[0].ArticleID, OperationType: ActionTypeBatchIgnore, Status: ActionStatusSuccess},
	}
	changeLogs := []InspectionFieldChangeLog{
		{ID: baseID*10 + 5, OrgID: orgID, ActionID: actions[0].ID, TaskID: task.ID, ResultID: results[0].ID, ArticleID: results[0].ArticleID, FieldName: KeywordScopeBody},
	}

	if err := db.Create(&taskKeywords).Error; err != nil {
		t.Fatalf("create task keywords error = %v", err)
	}
	if err := db.Create(&results).Error; err != nil {
		t.Fatalf("create results error = %v", err)
	}
	if err := db.Create(&hits).Error; err != nil {
		t.Fatalf("create hits error = %v", err)
	}
	if err := db.Create(&actions).Error; err != nil {
		t.Fatalf("create actions error = %v", err)
	}
	if err := db.Create(&opLogs).Error; err != nil {
		t.Fatalf("create operation logs error = %v", err)
	}
	if err := db.Create(&changeLogs).Error; err != nil {
		t.Fatalf("create field change logs error = %v", err)
	}

	return task
}

func assertTaskOwnedRowsDeleted(t *testing.T, db *gorm.DB, orgID, taskID uint64) {
	t.Helper()

	var taskCount int64
	if err := db.Model(&InspectionTask{}).Where("orgid = ? AND id = ?", orgID, taskID).Count(&taskCount).Error; err != nil {
		t.Fatalf("count task error = %v", err)
	}
	if taskCount != 0 {
		t.Fatalf("task count = %d, want %d", taskCount, 0)
	}

	checks := []struct {
		name  string
		model any
	}{
		{name: "task keywords", model: &InspectionTaskKeyword{}},
		{name: "results", model: &InspectionResult{}},
		{name: "result hits", model: &InspectionResultHit{}},
		{name: "actions", model: &InspectionAction{}},
		{name: "operation logs", model: &InspectionOperationLog{}},
		{name: "field change logs", model: &InspectionFieldChangeLog{}},
	}

	for _, check := range checks {
		var count int64
		if err := db.Model(check.model).Where("orgid = ? AND task_id = ?", orgID, taskID).Count(&count).Error; err != nil {
			t.Fatalf("count %s error = %v", check.name, err)
		}
		if count != 0 {
			t.Fatalf("%s count = %d, want %d", check.name, count, 0)
		}
	}
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
		`DROP TABLE IF EXISTS xt_article_inspect_categories`,
		`CREATE TABLE xt_article_inspect_categories (
			id INTEGER PRIMARY KEY,
			orgid INTEGER NOT NULL,
			name TEXT NOT NULL,
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
		`INSERT INTO xt_article_inspect_categories (id, orgid, name, enabled, sort, creator_id, creator_name, updater_id, updater_name, create_at, update_at) VALUES
			(501, 29, '政策红线', 1, 10, 7, 'alice', 7, 'alice', ?, ?),
			(502, 29, '高频违规', 1, 20, 7, 'alice', 7, 'alice', ?, ?),
			(601, 30, '外部分类', 1, 10, 8, 'bob', 8, 'bob', ?, ?),
			(1001, 100, '政策红线', 1, 10, 7, 'alice', 7, 'alice', ?, ?),
			(1002, 100, '高频违规', 1, 20, 7, 'alice', 7, 'alice', ?, ?),
			(2001, 200, '其他组织分类', 1, 10, 8, 'bob', 8, 'bob', ?, ?)`,
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
		{ID: 9003, OrgID: 29, Title: "待审稿件", State: ArticleStateAuditPending, PublishAtUnix: laterPublishAt.Unix(), UpdateAtUnix: laterPublishAt.Unix()},
		{ID: 9901, OrgID: 30, Title: "外部组织稿件", State: ArticleStateOnline, PublishAtUnix: publishAt.Unix(), UpdateAtUnix: publishAt.Unix()},
	}
	if err := db.Create(&articles).Error; err != nil {
		t.Fatalf("seed article center articles error = %v", err)
	}

	infos := []ArticleInfo{
		{ID: 9001, OrgID: 29, Body: "<p>real body one</p>"},
		{ID: 9002, OrgID: 29, Body: "<p>real body two</p>"},
		{ID: 9003, OrgID: 29, Body: "<p>pending body</p>"},
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

func seedBatchOfflineFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()

	results := []InspectionResult{
		{ID: 2001, OrgID: 100, TaskID: 501, ArticleID: 10, ArticleState: ArticleStateOnline, DispositionStatus: ResultDispositionPending},
		{ID: 2002, OrgID: 100, TaskID: 501, ArticleID: 11, ArticleState: ArticleStateOffline, DispositionStatus: ResultDispositionOfflined},
	}
	if err := db.Create(&results).Error; err != nil {
		t.Fatalf("seed batch offline results error = %v", err)
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

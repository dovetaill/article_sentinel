package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	path := filepath.Join("..", "..", "..", "..", "migrations", "20260420_01_article_inspection.sql")
	dropPath := filepath.Join("..", "..", "..", "..", "migrations", "20260428_01_drop_category_code.sql")

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

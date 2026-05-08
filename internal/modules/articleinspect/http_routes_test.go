package articleinspect

import (
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/identity"
)

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
	if _, ok := detailData["operation_logs"].([]any); !ok {
		t.Fatalf("detail operation_logs type = %T, want []any", detailData["operation_logs"])
	}
	if _, ok := detailData["field_change_logs"].([]any); !ok {
		t.Fatalf("detail field_change_logs type = %T, want []any", detailData["field_change_logs"])
	}
	if _, ok := detailData["field_changes"]; ok {
		t.Fatalf("detail keys = %#v, do not want field_changes alias", detailData)
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

func TestArticleInspectRoutesAcceptSessionScopedMutationBodiesWithoutOrgID(t *testing.T) {
	db := newArticleInspectTestDB(t)
	seedOrgCategoryFixtures(t, db)
	seedQueryFixtures(t, db)
	dispatcher := &articleInspectTaskDispatcherStub{}
	handler := newArticleInspectHandler(t, db, dispatcher)

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

	createdKeyword := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPost, "/api/v1/article-inspect/keywords", map[string]any{
		"name":           "session-keyword-without-orgid",
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

	createdTask := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPost, "/api/v1/article-inspect/tasks", map[string]any{
		"keyword_ids":   []uint64{keywordID},
		"include_body":  true,
		"article_state": ArticleStateOnline,
	}, session100)
	if createdTask.status != http.StatusCreated {
		t.Fatalf("create task status = %d, want %d", createdTask.status, http.StatusCreated)
	}
	if len(dispatcher.payloads) != 1 || dispatcher.payloads[0].OrgID != 100 {
		t.Fatalf("dispatcher payloads = %+v, want session org 100", dispatcher.payloads)
	}

	patchedCategory := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPatch, "/api/v1/article-inspect/categories/1001/status", map[string]any{
		"enabled": false,
	}, session100)
	if patchedCategory.status != http.StatusOK {
		t.Fatalf("patch category status = %d, want %d", patchedCategory.status, http.StatusOK)
	}

	patchedKeyword := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPatch, "/api/v1/article-inspect/keywords/"+strconv.FormatUint(keywordID, 10)+"/status", map[string]any{
		"enabled": false,
	}, session100)
	if patchedKeyword.status != http.StatusOK {
		t.Fatalf("patch keyword status = %d, want %d", patchedKeyword.status, http.StatusOK)
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

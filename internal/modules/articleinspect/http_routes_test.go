package articleinspect

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/identity"
	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
)

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
		{name: "category list", method: http.MethodGet, path: "/api/v1/article-inspect/categories?page=1&page_size=20"},
		{name: "article rectify", method: http.MethodPut, path: "/api/v1/article-inspect/articles/12/rectify", body: map[string]any{"title": "Updated title", "body": "Updated body content"}},
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
	if err := db.Create(&domainpkg.InspectionResult{ID: 3001, OrgID: 100, TaskID: 501, ArticleID: 10, ArticleState: domainpkg.ArticleStateOnline, DispositionStatus: domainpkg.ResultDispositionPending}).Error; err != nil {
		t.Fatalf("seed session-scoped batch result error = %v", err)
	}

	session100 := articleInspectRequestOptions{Session: &identity.AdminSession{UserID: 7, OrgID: 100, OrgName: "测试组织A", Nickname: "alice", Priv: "admin", Status: "active"}}
	session29 := articleInspectRequestOptions{Session: &identity.AdminSession{UserID: 7, OrgID: 29, OrgName: "一县一端", Nickname: "alice", Priv: "admin", Status: "active"}}

	createdCategory := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPost, "/api/v1/article-inspect/categories", map[string]any{"orgid": 100, "name": "会话机构分类", "enabled": true, "sort": 15}, session29)
	if createdCategory.status != http.StatusCreated {
		t.Fatalf("create category status = %d, want %d", createdCategory.status, http.StatusCreated)
	}
	categoryData := articleInspectDataMap(t, createdCategory.envelope.Data)
	if articleInspectUint64Field(t, categoryData, "orgid") != 29 {
		t.Fatalf("created category orgid = %d, want session org %d", articleInspectUint64Field(t, categoryData, "orgid"), 29)
	}

	createdKeyword := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPost, "/api/v1/article-inspect/keywords", map[string]any{"orgid": 200, "name": "session-keyword", "category_id": 1001, "match_type": domainpkg.MatchTypeContains, "risk_level": domainpkg.RiskLevelHigh, "suggest_action": domainpkg.SuggestActionOffline, "enabled": true, "scopes": []string{domainpkg.KeywordScopeTitle}}, session100)
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

	createdTask := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPost, "/api/v1/article-inspect/tasks", map[string]any{"orgid": 200, "keyword_ids": []uint64{keywordID}, "include_body": true, "article_state": domainpkg.ArticleStateOnline}, session100)
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

	batchOffline := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPost, "/api/v1/article-inspect/actions/batch-offline", map[string]any{"orgid": 200, "task_id": 501, "result_ids": []uint64{3001}, "reason": "session org should win"}, session100)
	if batchOffline.status != http.StatusOK {
		t.Fatalf("batch offline status = %d, want %d", batchOffline.status, http.StatusOK)
	}

	rectify := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPut, "/api/v1/article-inspect/articles/12/rectify", map[string]any{"orgid": 200, "title": "Session updated title", "body": "Session updated body"}, session100)
	if rectify.status != http.StatusOK {
		t.Fatalf("rectify status = %d, want %d", rectify.status, http.StatusOK)
	}

	republish := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPost, "/api/v1/article-inspect/articles/11/republish", map[string]any{"orgid": 200, "reason": "session org should win"}, session100)
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

	session100 := articleInspectRequestOptions{Session: &identity.AdminSession{UserID: 7, OrgID: 100, OrgName: "测试组织A", Nickname: "alice", Priv: "admin", Status: "active"}}

	createdKeyword := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPost, "/api/v1/article-inspect/keywords", map[string]any{"name": "session-keyword-without-orgid", "category_id": 1001, "match_type": domainpkg.MatchTypeContains, "risk_level": domainpkg.RiskLevelHigh, "suggest_action": domainpkg.SuggestActionOffline, "enabled": true, "scopes": []string{domainpkg.KeywordScopeTitle}}, session100)
	if createdKeyword.status != http.StatusCreated {
		t.Fatalf("create keyword status = %d, want %d", createdKeyword.status, http.StatusCreated)
	}
	createdKeywordData := articleInspectDataMap(t, createdKeyword.envelope.Data)
	keywordID := articleInspectUint64Field(t, createdKeywordData, "id")

	createdTask := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPost, "/api/v1/article-inspect/tasks", map[string]any{"keyword_ids": []uint64{keywordID}, "include_body": true, "article_state": domainpkg.ArticleStateOnline}, session100)
	if createdTask.status != http.StatusCreated {
		t.Fatalf("create task status = %d, want %d", createdTask.status, http.StatusCreated)
	}
	if len(dispatcher.payloads) != 1 || dispatcher.payloads[0].OrgID != 100 {
		t.Fatalf("dispatcher payloads = %+v, want session org 100", dispatcher.payloads)
	}

	patchedCategory := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPatch, "/api/v1/article-inspect/categories/1001/status", map[string]any{"enabled": false}, session100)
	if patchedCategory.status != http.StatusOK {
		t.Fatalf("patch category status = %d, want %d", patchedCategory.status, http.StatusOK)
	}

	patchedKeyword := sendArticleInspectJSONRequestWithOptions(t, handler, http.MethodPatch, "/api/v1/article-inspect/keywords/"+strconv.FormatUint(keywordID, 10)+"/status", map[string]any{"enabled": false}, session100)
	if patchedKeyword.status != http.StatusOK {
		t.Fatalf("patch keyword status = %d, want %d", patchedKeyword.status, http.StatusOK)
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

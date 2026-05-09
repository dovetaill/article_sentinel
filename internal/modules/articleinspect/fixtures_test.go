package articleinspect

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/article-sentinel/internal/api/response"
	"github.com/dovetaill/article-sentinel/internal/identity"
	"github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
	"gorm.io/gorm"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func newArticleInspectTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testutil.NewArticleInspectTestDB(t)
}

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	return testutil.MustTime(t, value)
}

func timePointer(value time.Time) *time.Time {
	return &value
}

func sortedStrings(values []string) []string {
	cloned := append([]string(nil), values...)
	sort.Strings(cloned)
	return cloned
}

func seedCandidateArticles(t *testing.T, db *gorm.DB) {
	t.Helper()
	testutil.SeedCandidateArticles(t, db)
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
	return testutil.SeedInspectionTaskForWorker(t, db, rules)
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

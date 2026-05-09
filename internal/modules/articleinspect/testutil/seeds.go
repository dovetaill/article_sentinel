package testutil

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	scanpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/scan"
	"gorm.io/gorm"
)

func SeedCandidateArticles(t *testing.T, db *gorm.DB) {
	t.Helper()

	articles := []domainpkg.Article{
		{ID: 1, OrgID: 100, Title: "Alpha news", State: domainpkg.ArticleStateOnline, PublishAtUnix: MustTime(t, "2026-04-20T10:00:00Z").Unix()},
		{ID: 2, OrgID: 100, Title: "Beta update", State: domainpkg.ArticleStateOnline, PublishAtUnix: MustTime(t, "2026-04-20T11:00:00Z").Unix()},
		{ID: 3, OrgID: 100, Title: "Gamma draft", State: domainpkg.ArticleStateDraft, PublishAtUnix: MustTime(t, "2026-04-20T12:00:00Z").Unix()},
		{ID: 4, OrgID: 200, Title: "Other org", State: domainpkg.ArticleStateOnline, PublishAtUnix: MustTime(t, "2026-04-20T10:30:00Z").Unix()},
	}
	if err := db.Create(&articles).Error; err != nil {
		t.Fatalf("seed articles error = %v", err)
	}

	infos := []domainpkg.ArticleInfo{
		{ID: 1, OrgID: 100, Body: "body one"},
		{ID: 2, OrgID: 100, Body: "body two"},
		{ID: 3, OrgID: 100, Body: "body three"},
		{ID: 4, OrgID: 200, Body: "body four"},
	}
	if err := db.Create(&infos).Error; err != nil {
		t.Fatalf("seed article infos error = %v", err)
	}
}

func SeedInspectionTaskForWorker(t *testing.T, db *gorm.DB, rules []scanpkg.KeywordRule) *domainpkg.InspectionTask {
	t.Helper()

	start := MustTime(t, "2026-04-20T09:00:00Z")
	end := MustTime(t, "2026-04-20T13:00:00Z")
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

	task := &domainpkg.InspectionTask{
		OrgID:              100,
		TaskNo:             "inspect-test",
		Status:             domainpkg.TaskStatusPending,
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

func SeedTaskForDeletion(t *testing.T, db *gorm.DB, orgID, baseID uint64, status string) *domainpkg.InspectionTask {
	t.Helper()

	task := &domainpkg.InspectionTask{
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

	taskKeywords := []domainpkg.InspectionTaskKeyword{{ID: baseID * 10, OrgID: orgID, TaskID: task.ID, KeywordID: baseID + 1}}
	results := []domainpkg.InspectionResult{{ID: baseID*10 + 1, OrgID: orgID, TaskID: task.ID, ArticleID: baseID + 100, ArticleTitle: "Delete me", DispositionStatus: domainpkg.ResultDispositionPending}}
	hits := []domainpkg.InspectionResultHit{{ID: baseID*10 + 2, OrgID: orgID, TaskID: task.ID, ResultID: results[0].ID, ArticleID: results[0].ArticleID, KeywordID: baseID + 1, KeywordText: "delete", FieldName: domainpkg.KeywordScopeTitle, MatchType: domainpkg.MatchTypeContains, RiskLevel: domainpkg.RiskLevelHigh, SuggestAction: domainpkg.SuggestActionOffline, Snippet: "delete snippet"}}
	actions := []domainpkg.InspectionAction{{ID: baseID*10 + 3, OrgID: orgID, ActionNo: fmt.Sprintf("act-%d", baseID), ActionType: domainpkg.ActionTypeBatchIgnore, TaskID: task.ID, Status: domainpkg.ActionStatusSuccess}}
	opLogs := []domainpkg.InspectionOperationLog{{ID: baseID*10 + 4, OrgID: orgID, ActionID: actions[0].ID, TaskID: task.ID, ResultID: results[0].ID, ArticleID: results[0].ArticleID, OperationType: domainpkg.ActionTypeBatchIgnore, Status: domainpkg.ActionStatusSuccess}}
	changeLogs := []domainpkg.InspectionFieldChangeLog{{ID: baseID*10 + 5, OrgID: orgID, ActionID: actions[0].ID, TaskID: task.ID, ResultID: results[0].ID, ArticleID: results[0].ArticleID, FieldName: domainpkg.KeywordScopeBody}}

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

func AssertTaskOwnedRowsDeleted(t *testing.T, db *gorm.DB, orgID, taskID uint64) {
	t.Helper()

	var taskCount int64
	if err := db.Model(&domainpkg.InspectionTask{}).Where("orgid = ? AND id = ?", orgID, taskID).Count(&taskCount).Error; err != nil {
		t.Fatalf("count task error = %v", err)
	}
	if taskCount != 0 {
		t.Fatalf("task count = %d, want %d", taskCount, 0)
	}

	checks := []struct {
		name  string
		model any
	}{
		{name: "task keywords", model: &domainpkg.InspectionTaskKeyword{}},
		{name: "results", model: &domainpkg.InspectionResult{}},
		{name: "result hits", model: &domainpkg.InspectionResultHit{}},
		{name: "actions", model: &domainpkg.InspectionAction{}},
		{name: "operation logs", model: &domainpkg.InspectionOperationLog{}},
		{name: "field change logs", model: &domainpkg.InspectionFieldChangeLog{}},
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

func SeedLifecycleArticles(t *testing.T, db *gorm.DB) {
	t.Helper()

	articles := []domainpkg.Article{
		{ID: 10, OrgID: 100, Title: "Online article", State: domainpkg.ArticleStateOnline},
		{ID: 11, OrgID: 100, Title: "Offline article", State: domainpkg.ArticleStateOffline},
		{ID: 12, OrgID: 100, Title: "Needs rectify", State: domainpkg.ArticleStateOffline},
	}
	if err := db.Create(&articles).Error; err != nil {
		t.Fatalf("seed lifecycle articles error = %v", err)
	}
	infos := []domainpkg.ArticleInfo{{ID: 10, OrgID: 100, Body: "body a"}, {ID: 11, OrgID: 100, Body: "body b"}, {ID: 12, OrgID: 100, Body: "body c"}}
	if err := db.Create(&infos).Error; err != nil {
		t.Fatalf("seed lifecycle article info error = %v", err)
	}
}

func SeedOrgCategoryFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()

	timestamp := MustTime(t, "2026-04-20T08:00:00Z")
	orgs := []domainpkg.ChuangqiOrg{
		{ID: 29, Name: "一县一端", CateID: 0, Enabled: true, Sort: 10, CreateAt: timestamp, UpdateAt: timestamp},
		{ID: 30, Name: "其他组织", CateID: 0, Enabled: true, Sort: 20, CreateAt: timestamp, UpdateAt: timestamp},
		{ID: 100, Name: "测试组织A", CateID: 0, Enabled: true, Sort: 30, CreateAt: timestamp, UpdateAt: timestamp},
		{ID: 200, Name: "测试组织B", CateID: 0, Enabled: true, Sort: 40, CreateAt: timestamp, UpdateAt: timestamp},
	}
	if err := db.Create(&orgs).Error; err != nil {
		t.Fatalf("seed orgs error = %v", err)
	}

	categories := []domainpkg.InspectionCategory{
		{ID: 501, OrgID: 29, Name: "政策红线", Enabled: true, Sort: 10, CreatorID: 7, CreatorName: "alice", UpdaterID: 7, UpdaterName: "alice", InspectionTimestamps: domainpkg.InspectionTimestamps{CreateAt: timestamp, UpdateAt: timestamp}},
		{ID: 502, OrgID: 29, Name: "高频违规", Enabled: true, Sort: 20, CreatorID: 7, CreatorName: "alice", UpdaterID: 7, UpdaterName: "alice", InspectionTimestamps: domainpkg.InspectionTimestamps{CreateAt: timestamp, UpdateAt: timestamp}},
		{ID: 601, OrgID: 30, Name: "外部分类", Enabled: true, Sort: 10, CreatorID: 8, CreatorName: "bob", UpdaterID: 8, UpdaterName: "bob", InspectionTimestamps: domainpkg.InspectionTimestamps{CreateAt: timestamp, UpdateAt: timestamp}},
		{ID: 1001, OrgID: 100, Name: "政策红线", Enabled: true, Sort: 10, CreatorID: 7, CreatorName: "alice", UpdaterID: 7, UpdaterName: "alice", InspectionTimestamps: domainpkg.InspectionTimestamps{CreateAt: timestamp, UpdateAt: timestamp}},
		{ID: 1002, OrgID: 100, Name: "高频违规", Enabled: true, Sort: 20, CreatorID: 7, CreatorName: "alice", UpdaterID: 7, UpdaterName: "alice", InspectionTimestamps: domainpkg.InspectionTimestamps{CreateAt: timestamp, UpdateAt: timestamp}},
		{ID: 2001, OrgID: 200, Name: "其他组织分类", Enabled: true, Sort: 10, CreatorID: 8, CreatorName: "bob", UpdaterID: 8, UpdaterName: "bob", InspectionTimestamps: domainpkg.InspectionTimestamps{CreateAt: timestamp, UpdateAt: timestamp}},
	}
	if err := db.Create(&categories).Error; err != nil {
		t.Fatalf("seed categories error = %v", err)
	}
}

func SeedArticleCenterFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()

	publishAt := MustTime(t, "2026-04-21T09:00:00Z")
	laterPublishAt := MustTime(t, "2026-04-21T10:30:00Z")
	latestActionAt := MustTime(t, "2026-04-21T12:30:00Z")
	olderActionAt := MustTime(t, "2026-04-21T11:00:00Z")

	articles := []domainpkg.Article{
		{ID: 9001, OrgID: 29, Title: "县域要闻一", Thumbnail: "https://example.com/article-9001.png", State: domainpkg.ArticleStateOnline, PublishAtUnix: publishAt.Unix(), UpdateAtUnix: latestActionAt.Unix()},
		{ID: 9002, OrgID: 29, Title: "县域要闻二", State: domainpkg.ArticleStateOffline, PublishAtUnix: laterPublishAt.Unix(), UpdateAtUnix: laterPublishAt.Unix()},
		{ID: 9003, OrgID: 29, Title: "待审稿件", State: domainpkg.ArticleStateAuditPending, PublishAtUnix: laterPublishAt.Unix(), UpdateAtUnix: laterPublishAt.Unix()},
		{ID: 9901, OrgID: 30, Title: "外部组织稿件", State: domainpkg.ArticleStateOnline, PublishAtUnix: publishAt.Unix(), UpdateAtUnix: publishAt.Unix()},
	}
	if err := db.Create(&articles).Error; err != nil {
		t.Fatalf("seed article center articles error = %v", err)
	}

	infos := []domainpkg.ArticleInfo{{ID: 9001, OrgID: 29, Body: "<p>real body one</p>"}, {ID: 9002, OrgID: 29, Body: "<p>real body two</p>"}, {ID: 9003, OrgID: 29, Body: "<p>pending body</p>"}, {ID: 9901, OrgID: 30, Body: "<p>other org body</p>"}}
	if err := db.Create(&infos).Error; err != nil {
		t.Fatalf("seed article center infos error = %v", err)
	}

	results := []domainpkg.InspectionResult{
		{ID: 7101, OrgID: 29, TaskID: 701, ArticleID: 9001, ArticleTitle: "县域要闻一", ArticleState: domainpkg.ArticleStateOnline, PublishAtTime: &publishAt, RiskLevel: domainpkg.RiskLevelLow, SuggestAction: domainpkg.SuggestActionIgnore, HitCount: 1, DispositionStatus: domainpkg.ResultDispositionPending, LatestActionAt: &olderActionAt, LatestOperatorID: 7, LatestOperatorName: "alice"},
		{ID: 7102, OrgID: 29, TaskID: 702, ArticleID: 9001, ArticleTitle: "县域要闻一", ArticleState: domainpkg.ArticleStateOnline, PublishAtTime: &publishAt, RiskLevel: domainpkg.RiskLevelHigh, SuggestAction: domainpkg.SuggestActionOffline, HitCount: 2, DispositionStatus: domainpkg.ResultDispositionProcessed, LatestActionAt: &latestActionAt, LatestOperatorID: 8, LatestOperatorName: "bob"},
	}
	if err := db.Create(&results).Error; err != nil {
		t.Fatalf("seed article center results error = %v", err)
	}
}

func SeedActionFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()

	results := []domainpkg.InspectionResult{
		{ID: 1001, OrgID: 100, TaskID: 501, ArticleID: 1, DispositionStatus: domainpkg.ResultDispositionPending},
		{ID: 1002, OrgID: 100, TaskID: 501, ArticleID: 2, DispositionStatus: domainpkg.ResultDispositionIgnored},
		{ID: 1003, OrgID: 100, TaskID: 501, ArticleID: 3, DispositionStatus: domainpkg.ResultDispositionPending},
		{ID: 1004, OrgID: 100, TaskID: 501, ArticleID: 4, DispositionStatus: domainpkg.ResultDispositionProcessed},
	}
	if err := db.Create(&results).Error; err != nil {
		t.Fatalf("seed action results error = %v", err)
	}
}

func SeedBatchOfflineFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()

	results := []domainpkg.InspectionResult{
		{ID: 2001, OrgID: 100, TaskID: 501, ArticleID: 10, ArticleState: domainpkg.ArticleStateOnline, DispositionStatus: domainpkg.ResultDispositionPending},
		{ID: 2002, OrgID: 100, TaskID: 501, ArticleID: 11, ArticleState: domainpkg.ArticleStateOffline, DispositionStatus: domainpkg.ResultDispositionOfflined},
	}
	if err := db.Create(&results).Error; err != nil {
		t.Fatalf("seed batch offline results error = %v", err)
	}
}

func SeedQueryFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()

	publishAt := MustTime(t, "2026-04-20T10:00:00Z")
	later := MustTime(t, "2026-04-20T11:00:00Z")
	createAt := MustTime(t, "2026-04-20T10:30:00Z")
	updateAt := MustTime(t, "2026-04-20T12:00:00Z")

	results := []domainpkg.InspectionResult{
		{ID: 1001, OrgID: 100, TaskID: 501, ArticleID: 1, ArticleTitle: "Alpha news", ArticleState: domainpkg.ArticleStateOnline, PublishAtTime: &publishAt, RiskLevel: domainpkg.RiskLevelHigh, SuggestAction: domainpkg.SuggestActionOffline, HitFieldsCount: 2, HitKeywordsCount: 2, HitCount: 2, DispositionStatus: domainpkg.ResultDispositionPending},
		{ID: 1002, OrgID: 100, TaskID: 501, ArticleID: 2, ArticleTitle: "Beta update", ArticleState: domainpkg.ArticleStateOnline, PublishAtTime: &later, RiskLevel: domainpkg.RiskLevelLow, SuggestAction: domainpkg.SuggestActionProcess, HitFieldsCount: 1, HitKeywordsCount: 1, HitCount: 1, DispositionStatus: domainpkg.ResultDispositionProcessed},
		{ID: 2001, OrgID: 200, TaskID: 601, ArticleID: 9, ArticleTitle: "Other org", ArticleState: domainpkg.ArticleStateOnline, PublishAtTime: &later, RiskLevel: domainpkg.RiskLevelHigh, SuggestAction: domainpkg.SuggestActionOffline, HitFieldsCount: 1, HitKeywordsCount: 1, HitCount: 1, DispositionStatus: domainpkg.ResultDispositionPending},
	}
	if err := db.Create(&results).Error; err != nil {
		t.Fatalf("seed query results error = %v", err)
	}

	hits := []domainpkg.InspectionResultHit{
		{ID: 1, OrgID: 100, TaskID: 501, ResultID: 1001, ArticleID: 1, KeywordID: 1, KeywordText: "alpha", FieldName: domainpkg.KeywordScopeTitle, MatchType: domainpkg.MatchTypeContains, RiskLevel: domainpkg.RiskLevelHigh, SuggestAction: domainpkg.SuggestActionOffline, MatchedText: "Alpha", Snippet: "Alpha news"},
		{ID: 2, OrgID: 100, TaskID: 501, ResultID: 1001, ArticleID: 1, KeywordID: 2, KeywordText: "body", FieldName: domainpkg.KeywordScopeBody, MatchType: domainpkg.MatchTypeContains, RiskLevel: domainpkg.RiskLevelHigh, SuggestAction: domainpkg.SuggestActionOffline, MatchedText: "body", Snippet: "body snippet"},
		{ID: 3, OrgID: 100, TaskID: 501, ResultID: 1002, ArticleID: 2, KeywordID: 3, KeywordText: "beta", FieldName: domainpkg.KeywordScopeTitle, MatchType: domainpkg.MatchTypeContains, RiskLevel: domainpkg.RiskLevelLow, SuggestAction: domainpkg.SuggestActionProcess, MatchedText: "Beta", Snippet: "Beta update"},
	}
	if err := db.Create(&hits).Error; err != nil {
		t.Fatalf("seed query hits error = %v", err)
	}

	opLogs := []domainpkg.InspectionOperationLog{
		{ID: 1, OrgID: 100, TaskID: 501, ResultID: 1001, ArticleID: 1, OperationType: domainpkg.ActionTypeOffline, BeforeState: "9", AfterState: "8", Status: domainpkg.ActionStatusSuccess, OperatorID: 7, OperatorName: "alice", InspectionTimestamps: domainpkg.InspectionTimestamps{CreateAt: createAt, UpdateAt: createAt}},
		{ID: 2, OrgID: 100, TaskID: 501, ResultID: 1001, ArticleID: 1, OperationType: domainpkg.ActionTypeRectify, BeforeState: "8", AfterState: "8", Status: domainpkg.ActionStatusSuccess, OperatorID: 7, OperatorName: "alice", InspectionTimestamps: domainpkg.InspectionTimestamps{CreateAt: updateAt, UpdateAt: updateAt}},
		{ID: 3, OrgID: 100, TaskID: 501, ResultID: 1002, ArticleID: 2, OperationType: domainpkg.ActionTypeBatchProcess, BeforeState: domainpkg.ResultDispositionPending, AfterState: domainpkg.ResultDispositionProcessed, Status: domainpkg.ActionStatusSuccess, OperatorID: 8, OperatorName: "bob", InspectionTimestamps: domainpkg.InspectionTimestamps{CreateAt: updateAt, UpdateAt: updateAt}},
	}
	if err := db.Create(&opLogs).Error; err != nil {
		t.Fatalf("seed operation logs error = %v", err)
	}

	changeLogs := []domainpkg.InspectionFieldChangeLog{
		{ID: 1, OrgID: 100, TaskID: 501, ResultID: 1001, ArticleID: 1, FieldName: domainpkg.KeywordScopeBody, BeforeValue: "old", AfterValue: "new", DiffSummary: "body: old -> new", OperatorID: 7, OperatorName: "alice", InspectionTimestamps: domainpkg.InspectionTimestamps{CreateAt: updateAt, UpdateAt: updateAt}},
		{ID: 2, OrgID: 100, TaskID: 501, ResultID: 1001, ArticleID: 1, FieldName: domainpkg.KeywordScopeTitle, BeforeValue: "old title", AfterValue: "new title", DiffSummary: "title: old title -> new title", OperatorID: 7, OperatorName: "alice", InspectionTimestamps: domainpkg.InspectionTimestamps{CreateAt: createAt, UpdateAt: createAt}},
	}
	if err := db.Create(&changeLogs).Error; err != nil {
		t.Fatalf("seed field change logs error = %v", err)
	}
}

func marshalJSON(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func timePointer(value time.Time) *time.Time {
	return &value
}

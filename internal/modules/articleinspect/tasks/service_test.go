package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/article-sentinel/internal/identity"
	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	outboxpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/outbox"
	rulespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/rules"
	scanpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/scan"
	testutil "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
	"gorm.io/gorm"
)

type articleRepositoryStub struct{}

func (articleRepositoryStub) ListCandidateArticles(ctx context.Context, filter CandidateArticleFilter) ([]scanpkg.CandidateArticle, uint64, error) {
	_ = ctx
	_ = filter
	return nil, 0, nil
}

type taskOutboxImmediateRelayStub struct {
	outboxIDs []uint64
}

func (s *taskOutboxImmediateRelayStub) TryDispatchMessage(ctx context.Context, outboxID uint64) bool {
	_ = ctx
	s.outboxIDs = append(s.outboxIDs, outboxID)
	return true
}

type dispatcherStub struct {
	payloads []queuetasks.ArticleInspectTaskPayload
	err      error
}

func (s *dispatcherStub) DispatchArticleInspectTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) error {
	_ = ctx
	s.payloads = append(s.payloads, payload)
	return s.err
}

func mustCreateKeyword(t *testing.T, service *rulespkg.KeywordService) *rulespkg.KeywordDTO {
	t.Helper()

	ctx := identity.ContextWithActor(context.Background(), identity.NewActor(9, "operator", "ops", "active"))
	keyword, err := service.Create(ctx, rulespkg.CreateKeywordInput{
		OrgID:         100,
		Name:          "spam",
		CategoryID:    1001,
		MatchType:     domainpkg.MatchTypeContains,
		RiskLevel:     domainpkg.RiskLevelHigh,
		SuggestAction: domainpkg.SuggestActionOffline,
		Enabled:       true,
		Scopes:        []string{domainpkg.KeywordScopeTitle, domainpkg.KeywordScopeBody},
	})
	if err != nil {
		t.Fatalf("Create keyword error = %v", err)
	}
	return keyword
}

func newTaskHTTPHandler(t *testing.T, db *gorm.DB, dispatcher outboxpkg.TaskDispatcher) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	group := huma.NewGroup(api, "/api/v1/article-inspect")
	group.UseSimpleModifier(func(op *huma.Operation) {
		op.SkipValidateParams = true
	})
	RegisterTaskRoutes(group, NewTaskService(db, rulespkg.NewKeywordRepository(db), articleRepositoryStub{}), dispatcher, nil, outboxpkg.TaskOutboxSettings{})
	return mux
}

func TestRegisterTaskRoutesUsesOutboxDispatcher(t *testing.T) {
	var register func(huma.API, *TaskService, outboxpkg.TaskDispatcher, *slog.Logger, outboxpkg.TaskOutboxSettings)
	register = RegisterTaskRoutes
	if register == nil {
		t.Fatal("RegisterTaskRoutes signature adapter = nil")
	}
}

func TestTaskCreation(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	service := NewTaskService(db, rulespkg.NewKeywordRepository(db), articleRepositoryStub{})
	keywordService := rulespkg.NewKeywordService(rulespkg.NewKeywordRepository(db))
	ctx := identity.ContextWithActor(context.Background(), identity.NewActor(9, "operator", "ops", "active"))

	keyword := mustCreateKeyword(t, keywordService)
	start := testutil.MustTime(t, "2026-04-20T09:00:00Z")
	end := testutil.MustTime(t, "2026-04-20T13:00:00Z")
	created, err := service.Create(ctx, CreateInspectionTaskInput{
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
	if created.Status != domainpkg.TaskStatusPending {
		t.Fatalf("Create().Status = %q, want %q", created.Status, domainpkg.TaskStatusPending)
	}
	if created.RuleSnapshot == "" || !strings.Contains(created.RuleSnapshot, "spam") {
		t.Fatalf("Create().RuleSnapshot = %q, want contains %q", created.RuleSnapshot, "spam")
	}
	if created.RequestSnapshot == "" || !strings.Contains(created.RequestSnapshot, "\"orgid\":100") {
		t.Fatalf("Create().RequestSnapshot = %q, want contains %q", created.RequestSnapshot, "\"orgid\":100")
	}

	var taskKeywords []domainpkg.InspectionTaskKeyword
	if err := db.Where("orgid = ? AND task_id = ?", 100, created.ID).Find(&taskKeywords).Error; err != nil {
		t.Fatalf("Find task keywords error = %v", err)
	}
	if len(taskKeywords) != 1 || taskKeywords[0].KeywordID != keyword.ID {
		t.Fatalf("task keywords = %#v, want keyword %d linked once", taskKeywords, keyword.ID)
	}

	_, err = service.Create(ctx, CreateInspectionTaskInput{KeywordIDs: []uint64{keyword.ID}})
	if !errors.Is(err, ErrInvalidTaskInput) {
		t.Fatalf("Create(missing orgid) error = %v, want %v", err, ErrInvalidTaskInput)
	}
}

func TestTaskCreationRuleSnapshotUsesStableRuleShape(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	service := NewTaskService(db, rulespkg.NewKeywordRepository(db), articleRepositoryStub{})
	keywordService := rulespkg.NewKeywordService(rulespkg.NewKeywordRepository(db))
	ctx := identity.ContextWithActor(context.Background(), identity.NewActor(9, "operator", "ops", "active"))

	keyword, err := keywordService.Create(ctx, rulespkg.CreateKeywordInput{
		OrgID:         100,
		Name:          "spam",
		CategoryID:    1001,
		MatchType:     domainpkg.MatchTypeContains,
		RiskLevel:     domainpkg.RiskLevelHigh,
		SuggestAction: domainpkg.SuggestActionOffline,
		Enabled:       true,
		Remark:        "internal note",
		Scopes:        []string{domainpkg.KeywordScopeTitle, domainpkg.KeywordScopeBody},
	})
	if err != nil {
		t.Fatalf("Create keyword error = %v", err)
	}

	created, err := service.Create(ctx, CreateInspectionTaskInput{OrgID: 100, KeywordIDs: []uint64{keyword.ID}})
	if err != nil {
		t.Fatalf("Create task error = %v", err)
	}

	var snapshot []map[string]any
	if err := json.Unmarshal([]byte(created.RuleSnapshot), &snapshot); err != nil {
		t.Fatalf("unmarshal RuleSnapshot error = %v", err)
	}
	if len(snapshot) != 1 {
		t.Fatalf("RuleSnapshot items len = %d, want %d", len(snapshot), 1)
	}
	item := snapshot[0]
	if _, ok := item["category_name"]; !ok {
		t.Fatalf("RuleSnapshot item = %#v, want category_name", item)
	}
	for _, forbidden := range []string{"enabled", "remark", "creator_id", "creator_name", "updater_id", "updater_name", "category_id"} {
		if _, ok := item[forbidden]; ok {
			t.Fatalf("RuleSnapshot item = %#v, do not want %q in stable rule snapshot", item, forbidden)
		}
	}
}

func TestTaskCreationWithOutbox(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	service := NewTaskService(db, rulespkg.NewKeywordRepository(db), articleRepositoryStub{})
	keywordService := rulespkg.NewKeywordService(rulespkg.NewKeywordRepository(db))
	ctx := identity.ContextWithActor(context.Background(), identity.NewActor(9, "operator", "ops", "active"))

	keyword := mustCreateKeyword(t, keywordService)
	created, outbox, err := service.CreateWithOutbox(ctx, CreateInspectionTaskInput{
		OrgID:          100,
		KeywordIDs:     []uint64{keyword.ID},
		IncludeBody:    true,
		ArticleState:   domainpkg.ArticleStateOnline,
		PublishTimeEnd: timePointer(testutil.MustTime(t, "2026-04-20T13:00:00Z")),
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
	if outbox.Status != domainpkg.TaskOutboxStatusPending {
		t.Fatalf("CreateWithOutbox().Outbox.Status = %q, want %q", outbox.Status, domainpkg.TaskOutboxStatusPending)
	}
	if outbox.MessageType != domainpkg.TaskOutboxMessageTypeRunTask {
		t.Fatalf("CreateWithOutbox().Outbox.MessageType = %q, want %q", outbox.MessageType, domainpkg.TaskOutboxMessageTypeRunTask)
	}
	if !strings.Contains(outbox.Payload, strconv.FormatUint(created.ID, 10)) {
		t.Fatalf("CreateWithOutbox().Outbox.Payload = %q, want contains task id %d", outbox.Payload, created.ID)
	}

	var stored domainpkg.InspectionTaskOutboxMessage
	if err := db.Where("orgid = ? AND task_id = ?", 100, created.ID).First(&stored).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	if stored.Status != domainpkg.TaskOutboxStatusPending {
		t.Fatalf("stored outbox status = %q, want %q", stored.Status, domainpkg.TaskOutboxStatusPending)
	}
}

func TestTaskCreateAndEnqueueUsesImmediateRelaySeam(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	service := NewTaskService(db, rulespkg.NewKeywordRepository(db), articleRepositoryStub{})
	keywordService := rulespkg.NewKeywordService(rulespkg.NewKeywordRepository(db))
	ctx := identity.ContextWithActor(context.Background(), identity.NewActor(9, "operator", "ops", "active"))

	keyword := mustCreateKeyword(t, keywordService)
	relay := &taskOutboxImmediateRelayStub{}
	created, err := service.CreateAndEnqueue(ctx, CreateInspectionTaskInput{
		OrgID:        100,
		KeywordIDs:   []uint64{keyword.ID},
		IncludeBody:  true,
		ArticleState: domainpkg.ArticleStateOnline,
	}, relay)
	if err != nil {
		t.Fatalf("CreateAndEnqueue() error = %v", err)
	}
	if created.ID == 0 {
		t.Fatal("CreateAndEnqueue().Task.ID = 0, want persisted task id")
	}
	if len(relay.outboxIDs) != 1 {
		t.Fatalf("relay outboxIDs len = %d, want %d", len(relay.outboxIDs), 1)
	}

	var outbox domainpkg.InspectionTaskOutboxMessage
	if err := db.Where("orgid = ? AND task_id = ?", 100, created.ID).First(&outbox).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	if relay.outboxIDs[0] != outbox.ID {
		t.Fatalf("relay outboxID = %d, want %d", relay.outboxIDs[0], outbox.ID)
	}
	if outbox.Status != domainpkg.TaskOutboxStatusPending {
		t.Fatalf("outbox status = %q, want %q", outbox.Status, domainpkg.TaskOutboxStatusPending)
	}
}

func TestTaskCreateAndEnqueueWithTaskOutboxRelayDispatchesCommittedMessage(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	service := NewTaskService(db, rulespkg.NewKeywordRepository(db), articleRepositoryStub{})
	keywordService := rulespkg.NewKeywordService(rulespkg.NewKeywordRepository(db))
	ctx := identity.ContextWithActor(context.Background(), identity.NewActor(9, "operator", "ops", "active"))

	keyword := mustCreateKeyword(t, keywordService)
	dispatcher := &dispatcherStub{}
	relay := outboxpkg.NewTaskOutboxRelay(db, dispatcher, nil)
	created, err := service.CreateAndEnqueue(ctx, CreateInspectionTaskInput{
		OrgID:        100,
		KeywordIDs:   []uint64{keyword.ID},
		IncludeBody:  true,
		ArticleState: domainpkg.ArticleStateOnline,
	}, relay)
	if err != nil {
		t.Fatalf("CreateAndEnqueue() error = %v", err)
	}
	if created.ID == 0 {
		t.Fatal("CreateAndEnqueue().Task.ID = 0, want persisted task id")
	}
	if len(dispatcher.payloads) != 1 {
		t.Fatalf("dispatcher payloads len = %d, want %d", len(dispatcher.payloads), 1)
	}
	if dispatcher.payloads[0].TaskID != created.ID || dispatcher.payloads[0].OrgID != created.OrgID {
		t.Fatalf("dispatcher payload = %+v, want task=%d org=%d", dispatcher.payloads[0], created.ID, created.OrgID)
	}

	var outbox domainpkg.InspectionTaskOutboxMessage
	if err := db.Where("orgid = ? AND task_id = ?", 100, created.ID).First(&outbox).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	if outbox.Status != domainpkg.TaskOutboxStatusDispatched {
		t.Fatalf("outbox status = %q, want %q", outbox.Status, domainpkg.TaskOutboxStatusDispatched)
	}
	if outbox.AttemptCount != 1 {
		t.Fatalf("outbox attempt_count = %d, want %d", outbox.AttemptCount, 1)
	}
	if outbox.DispatchedAt == nil {
		t.Fatal("outbox dispatched_at = nil, want timestamp")
	}
}

func TestTaskDelete(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	service := NewTaskService(db, rulespkg.NewKeywordRepository(db), articleRepositoryStub{})

	t.Run("deletes pending task and dependent rows", func(t *testing.T) {
		task := testutil.SeedTaskForDeletion(t, db, 100, 901, domainpkg.TaskStatusPending)

		if err := service.Delete(context.Background(), 100, task.ID); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		testutil.AssertTaskOwnedRowsDeleted(t, db, 100, task.ID)
	})

	t.Run("deletes failed task and dependent rows", func(t *testing.T) {
		task := testutil.SeedTaskForDeletion(t, db, 100, 902, domainpkg.TaskStatusFailed)

		if err := service.Delete(context.Background(), 100, task.ID); err != nil {
			t.Fatalf("Delete(failed) error = %v", err)
		}

		testutil.AssertTaskOwnedRowsDeleted(t, db, 100, task.ID)
	})

	for index, status := range []string{domainpkg.TaskStatusRunning, domainpkg.TaskStatusSuccess, domainpkg.TaskStatusPartialSuccess} {
		t.Run("rejects "+status+" task deletion", func(t *testing.T) {
			task := testutil.SeedTaskForDeletion(t, db, 100, uint64(1001+index), status)

			err := service.Delete(context.Background(), 100, task.ID)
			if !errors.Is(err, ErrTaskDeleteNotAllowed) {
				t.Fatalf("Delete(%s) error = %v, want %v", status, err, ErrTaskDeleteNotAllowed)
			}

			var count int64
			if err := db.Model(&domainpkg.InspectionTask{}).Where("orgid = ? AND id = ?", 100, task.ID).Count(&count).Error; err != nil {
				t.Fatalf("count task error = %v", err)
			}
			if count != 1 {
				t.Fatalf("task count after rejected delete = %d, want %d", count, 1)
			}
		})
	}
}

func TestTaskRoutesCRUD(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	keywordService := rulespkg.NewKeywordService(rulespkg.NewKeywordRepository(db))
	keyword := mustCreateKeyword(t, keywordService)
	dispatcher := &dispatcherStub{}
	handler := newTaskHTTPHandler(t, db, dispatcher)

	created := testutil.SendJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/tasks", map[string]any{
		"orgid":         100,
		"keyword_ids":   []uint64{keyword.ID},
		"include_body":  true,
		"article_state": domainpkg.ArticleStateOnline,
	})
	if created.Status != http.StatusCreated {
		t.Fatalf("create task status = %d, want %d", created.Status, http.StatusCreated)
	}
	createdData := testutil.DataMap(t, created.Envelope.Data)
	if len(dispatcher.payloads) != 1 {
		t.Fatalf("dispatcher payloads len = %d, want %d", len(dispatcher.payloads), 1)
	}
	if dispatcher.payloads[0].OrgID != 100 || dispatcher.payloads[0].TaskID == 0 {
		t.Fatalf("dispatcher payload = %+v, want orgid and task id", dispatcher.payloads[0])
	}

	listed := testutil.SendRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/tasks?orgid=100&page=1&page_size=20", nil)
	if listed.Status != http.StatusOK {
		t.Fatalf("list tasks status = %d, want %d", listed.Status, http.StatusOK)
	}
	listData := testutil.DataMap(t, listed.Envelope.Data)
	if total := testutil.NumberField(t, listData, "total"); total != 1 {
		t.Fatalf("list tasks total = %v, want %d", total, 1)
	}
	items := testutil.ListField(t, listData, "items")
	if len(items) != 1 {
		t.Fatalf("list tasks items len = %d, want %d", len(items), 1)
	}
	listedTask := testutil.DataMap(t, items[0])
	if testutil.StringField(t, listedTask, "status") != domainpkg.TaskStatusPending {
		t.Fatalf("listed task status = %q, want %q", testutil.StringField(t, listedTask, "status"), domainpkg.TaskStatusPending)
	}
	if _, ok := listedTask["created_at"]; !ok {
		t.Fatalf("listed task keys = %#v, want created_at", listedTask)
	}
	if _, ok := listedTask["create_at"]; ok {
		t.Fatalf("listed task keys = %#v, do not want create_at", listedTask)
	}

	taskID := testutil.Uint64Field(t, createdData, "id")
	detail := testutil.SendRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/tasks/"+strconv.FormatUint(taskID, 10)+"?orgid=100", nil)
	if detail.Status != http.StatusOK {
		t.Fatalf("get task status = %d, want %d", detail.Status, http.StatusOK)
	}
	detailData := testutil.DataMap(t, detail.Envelope.Data)
	if testutil.StringField(t, detailData, "task_no") != testutil.StringField(t, createdData, "task_no") {
		t.Fatalf("task detail task_no = %q, want %q", testutil.StringField(t, detailData, "task_no"), testutil.StringField(t, createdData, "task_no"))
	}

	deleted := testutil.SendRequest(t, handler, http.MethodDelete, "/api/v1/article-inspect/tasks/"+strconv.FormatUint(taskID, 10)+"?orgid=100", nil)
	if deleted.Status != http.StatusOK {
		t.Fatalf("delete task status = %d, want %d", deleted.Status, http.StatusOK)
	}
}

func TestTaskCreateEnqueueFailureLeavesPendingOutbox(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	handler := newTaskHTTPHandler(t, db, &dispatcherStub{err: errors.New("queue down")})
	keywordService := rulespkg.NewKeywordService(rulespkg.NewKeywordRepository(db))
	createdKeyword := mustCreateKeyword(t, keywordService)

	result := testutil.SendJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/tasks", map[string]any{
		"orgid":         100,
		"keyword_ids":   []uint64{createdKeyword.ID},
		"include_body":  true,
		"article_state": domainpkg.ArticleStateOnline,
	})
	if result.Status != http.StatusCreated {
		t.Fatalf("create task status = %d, want %d", result.Status, http.StatusCreated)
	}
	if result.Envelope.Code != 0 {
		t.Fatalf("create task envelope = %+v, want success code", result.Envelope)
	}

	var taskCount int64
	if err := db.Model(&domainpkg.InspectionTask{}).Where("orgid = ?", 100).Count(&taskCount).Error; err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	if taskCount != 1 {
		t.Fatalf("task count after enqueue failure = %d, want %d", taskCount, 1)
	}

	var taskKeywordCount int64
	if err := db.Model(&domainpkg.InspectionTaskKeyword{}).Where("orgid = ?", 100).Count(&taskKeywordCount).Error; err != nil {
		t.Fatalf("count task keywords: %v", err)
	}
	if taskKeywordCount != 1 {
		t.Fatalf("task keyword count after enqueue failure = %d, want %d", taskKeywordCount, 1)
	}

	var outbox domainpkg.InspectionTaskOutboxMessage
	if err := db.Where("orgid = ?", 100).First(&outbox).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	if outbox.Status != domainpkg.TaskOutboxStatusPending {
		t.Fatalf("outbox status after enqueue failure = %q, want %q", outbox.Status, domainpkg.TaskOutboxStatusPending)
	}
	if outbox.AttemptCount != 1 {
		t.Fatalf("outbox attempt_count after enqueue failure = %d, want %d", outbox.AttemptCount, 1)
	}
	if !strings.Contains(outbox.LastError, "queue down") {
		t.Fatalf("outbox last_error = %q, want contains %q", outbox.LastError, "queue down")
	}
}

func TestTaskCreateWithoutDispatcherStillCreatesPendingOutbox(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	handler := newTaskHTTPHandler(t, db, nil)
	keywordService := rulespkg.NewKeywordService(rulespkg.NewKeywordRepository(db))
	createdKeyword := mustCreateKeyword(t, keywordService)

	result := testutil.SendJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/tasks", map[string]any{
		"orgid":         100,
		"keyword_ids":   []uint64{createdKeyword.ID},
		"include_body":  true,
		"article_state": domainpkg.ArticleStateOnline,
	})
	if result.Status != http.StatusCreated {
		t.Fatalf("create task status = %d, want %d", result.Status, http.StatusCreated)
	}

	var taskCount int64
	if err := db.Model(&domainpkg.InspectionTask{}).Where("orgid = ?", 100).Count(&taskCount).Error; err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	if taskCount != 1 {
		t.Fatalf("task count without dispatcher = %d, want %d", taskCount, 1)
	}

	var outbox domainpkg.InspectionTaskOutboxMessage
	if err := db.Where("orgid = ?", 100).First(&outbox).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	if outbox.Status != domainpkg.TaskOutboxStatusPending {
		t.Fatalf("outbox status without dispatcher = %q, want %q", outbox.Status, domainpkg.TaskOutboxStatusPending)
	}
	if outbox.AttemptCount != 0 {
		t.Fatalf("outbox attempt_count without dispatcher = %d, want %d", outbox.AttemptCount, 0)
	}
}

func timePointer(value time.Time) *time.Time {
	return &value
}

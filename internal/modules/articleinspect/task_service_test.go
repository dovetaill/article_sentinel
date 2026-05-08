package articleinspect

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/identity"
)

type taskOutboxImmediateRelayStub struct {
	outboxIDs []uint64
}

func (s *taskOutboxImmediateRelayStub) TryDispatchMessage(ctx context.Context, outboxID uint64) bool {
	_ = ctx
	s.outboxIDs = append(s.outboxIDs, outboxID)
	return true
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

func TestTaskCreationRuleSnapshotUsesStableRuleShape(t *testing.T) {
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
		Remark:        "internal note",
		Scopes:        []string{KeywordScopeTitle, KeywordScopeBody},
	})
	if err != nil {
		t.Fatalf("Create keyword error = %v", err)
	}

	created, err := taskService.Create(ctx, CreateInspectionTaskInput{
		OrgID:      100,
		KeywordIDs: []uint64{keyword.ID},
	})
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

func TestTaskCreateAndEnqueueUsesImmediateRelaySeam(t *testing.T) {
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

	relay := &taskOutboxImmediateRelayStub{}
	created, err := taskService.CreateAndEnqueue(ctx, CreateInspectionTaskInput{
		OrgID:        100,
		KeywordIDs:   []uint64{keyword.ID},
		IncludeBody:  true,
		ArticleState: ArticleStateOnline,
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

	var outbox InspectionTaskOutboxMessage
	if err := db.Where("orgid = ? AND task_id = ?", 100, created.ID).First(&outbox).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	if relay.outboxIDs[0] != outbox.ID {
		t.Fatalf("relay outboxID = %d, want %d", relay.outboxIDs[0], outbox.ID)
	}
	if outbox.Status != TaskOutboxStatusPending {
		t.Fatalf("outbox status = %q, want %q", outbox.Status, TaskOutboxStatusPending)
	}
}

func TestTaskCreateAndEnqueueWithTaskOutboxRelayDispatchesCommittedMessage(t *testing.T) {
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

	dispatcher := &articleInspectTaskDispatcherStub{}
	relay := NewTaskOutboxRelay(db, dispatcher, nil)
	created, err := taskService.CreateAndEnqueue(ctx, CreateInspectionTaskInput{
		OrgID:        100,
		KeywordIDs:   []uint64{keyword.ID},
		IncludeBody:  true,
		ArticleState: ArticleStateOnline,
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

	var outbox InspectionTaskOutboxMessage
	if err := db.Where("orgid = ? AND task_id = ?", 100, created.ID).First(&outbox).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	if outbox.Status != TaskOutboxStatusDispatched {
		t.Fatalf("outbox status = %q, want %q", outbox.Status, TaskOutboxStatusDispatched)
	}
	if outbox.AttemptCount != 1 {
		t.Fatalf("outbox attempt_count = %d, want %d", outbox.AttemptCount, 1)
	}
	if outbox.DispatchedAt == nil {
		t.Fatal("outbox dispatched_at = nil, want timestamp")
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

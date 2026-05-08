package articleinspect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dovetaill/article-sentinel/internal/identity"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
)

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

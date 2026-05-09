package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	testutil "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
	"gorm.io/gorm"
)

type dispatcherStub struct {
	payloads []queuetasks.ArticleInspectTaskPayload
	err      error
}

func (s *dispatcherStub) DispatchArticleInspectTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) error {
	_ = ctx
	s.payloads = append(s.payloads, payload)
	return s.err
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

func TestTaskOutboxRelayDispatchesPendingMessage(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	payload, err := json.Marshal(queuetasks.ArticleInspectTaskPayload{TaskID: 88, OrgID: 100, TriggerSource: "api"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	message := domainpkg.InspectionTaskOutboxMessage{
		OrgID:       100,
		TaskID:      88,
		MessageType: domainpkg.TaskOutboxMessageTypeRunTask,
		Status:      domainpkg.TaskOutboxStatusPending,
		Payload:     string(payload),
	}
	if err := db.Create(&message).Error; err != nil {
		t.Fatalf("create outbox row: %v", err)
	}

	dispatcher := &dispatcherStub{}
	relay := NewTaskOutboxRelay(db, dispatcher, nil)
	if _, err := relay.DispatchPending(context.Background(), 10); err != nil {
		t.Fatalf("DispatchPending() error = %v", err)
	}
	if len(dispatcher.payloads) != 1 {
		t.Fatalf("dispatcher payloads len = %d, want %d", len(dispatcher.payloads), 1)
	}
	if dispatcher.payloads[0].TaskID != 88 || dispatcher.payloads[0].OrgID != 100 {
		t.Fatalf("dispatcher payload = %+v, want task=88 org=100", dispatcher.payloads[0])
	}

	var stored domainpkg.InspectionTaskOutboxMessage
	if err := db.First(&stored, message.ID).Error; err != nil {
		t.Fatalf("load outbox row: %v", err)
	}
	if stored.Status != domainpkg.TaskOutboxStatusDispatched {
		t.Fatalf("outbox status = %q, want %q", stored.Status, domainpkg.TaskOutboxStatusDispatched)
	}
	if stored.AttemptCount != 1 {
		t.Fatalf("outbox attempt_count = %d, want %d", stored.AttemptCount, 1)
	}
	if stored.DispatchedAt == nil {
		t.Fatal("outbox dispatched_at = nil, want timestamp")
	}
}

func TestTaskOutboxRelayReclaimsExpiredLease(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	payload, err := json.Marshal(queuetasks.ArticleInspectTaskPayload{TaskID: 88, OrgID: 100, TriggerSource: "api"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	expiredAt := testutil.MustTime(t, "2026-04-20T13:00:00Z")
	claimedAt := expiredAt.Add(-time.Minute)
	if err := db.Exec(
		`INSERT INTO xt_article_inspect_task_outbox
		(orgid, task_id, message_type, status, payload, claimed_by, claimed_at, claim_until, create_at, update_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		100, 88, domainpkg.TaskOutboxMessageTypeRunTask, domainpkg.TaskOutboxStatusClaimed, string(payload), "scheduler@test", claimedAt, expiredAt, claimedAt, claimedAt,
	).Error; err != nil {
		t.Fatalf("insert claimed outbox row: %v", err)
	}

	dispatcher := &dispatcherStub{}
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
	if row.Status != domainpkg.TaskOutboxStatusDispatched {
		t.Fatalf("reclaimed outbox status = %q, want %q", row.Status, domainpkg.TaskOutboxStatusDispatched)
	}
	if !row.DispatchedAt.Valid {
		t.Fatal("reclaimed outbox dispatched_at = NULL, want timestamp")
	}
}

func TestTaskOutboxRelayRetryableFailureSchedulesNextAttempt(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	payload, err := json.Marshal(queuetasks.ArticleInspectTaskPayload{TaskID: 99, OrgID: 100, TriggerSource: "api"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	message := domainpkg.InspectionTaskOutboxMessage{OrgID: 100, TaskID: 99, MessageType: domainpkg.TaskOutboxMessageTypeRunTask, Status: domainpkg.TaskOutboxStatusPending, Payload: string(payload)}
	if err := db.Create(&message).Error; err != nil {
		t.Fatalf("create outbox row: %v", err)
	}

	dispatcher := &dispatcherStub{err: errors.New("queue down")}
	relay := NewTaskOutboxRelay(db, dispatcher, nil)

	report, err := relay.DispatchPending(context.Background(), 1)
	if err != nil {
		t.Fatalf("DispatchPending() error = %v", err)
	}
	if report.Failed != 1 {
		t.Fatalf("DispatchPending().Failed = %d, want %d", report.Failed, 1)
	}

	row := loadOutboxPhase3Row(t, db, message.ID)
	if row.Status != domainpkg.TaskOutboxStatusPending {
		t.Fatalf("retryable failure status = %q, want %q", row.Status, domainpkg.TaskOutboxStatusPending)
	}
	if !row.NextAttemptAt.Valid {
		t.Fatal("retryable failure next_attempt_at = NULL, want timestamp")
	}
	if row.LastErrorCode.String != domainpkg.TaskOutboxErrorDispatch {
		t.Fatalf("retryable failure last_error_code = %q, want %q", row.LastErrorCode.String, domainpkg.TaskOutboxErrorDispatch)
	}
}

func TestTaskOutboxRelayMovesPoisonMessageToDeadLetter(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	message := domainpkg.InspectionTaskOutboxMessage{OrgID: 100, TaskID: 77, MessageType: domainpkg.TaskOutboxMessageTypeRunTask, Status: domainpkg.TaskOutboxStatusPending, Payload: "{bad-json"}
	if err := db.Create(&message).Error; err != nil {
		t.Fatalf("create poison outbox row: %v", err)
	}

	dispatcher := &dispatcherStub{}
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
	if row.Status != domainpkg.TaskOutboxStatusDeadLetter {
		t.Fatalf("poison message status = %q, want %q", row.Status, domainpkg.TaskOutboxStatusDeadLetter)
	}
	if !row.DeadLetteredAt.Valid {
		t.Fatal("poison message dead_lettered_at = NULL, want timestamp")
	}
	if row.LastErrorCode.String != domainpkg.TaskOutboxErrorPayloadDecode {
		t.Fatalf("poison message last_error_code = %q, want %q", row.LastErrorCode.String, domainpkg.TaskOutboxErrorPayloadDecode)
	}
}

func TestTaskOutboxRelayImplementsCleanerContract(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
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
	db := testutil.NewArticleInspectTestDB(t)
	dispatcher := &dispatcherStub{}
	relay := NewTaskOutboxRelay(db, dispatcher, nil)

	bad := domainpkg.InspectionTaskOutboxMessage{OrgID: 100, TaskID: 1, MessageType: domainpkg.TaskOutboxMessageTypeRunTask, Status: domainpkg.TaskOutboxStatusPending, Payload: "{not-json"}
	if err := db.Create(&bad).Error; err != nil {
		t.Fatalf("create bad outbox row: %v", err)
	}

	goodPayload, err := json.Marshal(queuetasks.ArticleInspectTaskPayload{TaskID: 2, OrgID: 100, TriggerSource: "api"})
	if err != nil {
		t.Fatalf("marshal good payload: %v", err)
	}
	good := domainpkg.InspectionTaskOutboxMessage{OrgID: 100, TaskID: 2, MessageType: domainpkg.TaskOutboxMessageTypeRunTask, Status: domainpkg.TaskOutboxStatusPending, Payload: string(goodPayload)}
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

	var badStored domainpkg.InspectionTaskOutboxMessage
	if err := db.First(&badStored, bad.ID).Error; err != nil {
		t.Fatalf("load bad outbox row: %v", err)
	}
	badPhase3 := loadOutboxPhase3Row(t, db, bad.ID)
	if badStored.Status != domainpkg.TaskOutboxStatusDeadLetter {
		t.Fatalf("bad outbox status = %q, want %q", badStored.Status, domainpkg.TaskOutboxStatusDeadLetter)
	}
	if badPhase3.AttemptCount != 1 {
		t.Fatalf("bad outbox attempt_count = %d, want %d", badPhase3.AttemptCount, 1)
	}
	if !badPhase3.DeadLetteredAt.Valid {
		t.Fatal("bad outbox dead_lettered_at = NULL, want timestamp")
	}
	if badPhase3.LastErrorCode.String != domainpkg.TaskOutboxErrorPayloadDecode {
		t.Fatalf("bad outbox last_error_code = %q, want %q", badPhase3.LastErrorCode.String, domainpkg.TaskOutboxErrorPayloadDecode)
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

	var goodStored domainpkg.InspectionTaskOutboxMessage
	if err := db.First(&goodStored, good.ID).Error; err != nil {
		t.Fatalf("load good outbox row: %v", err)
	}
	if goodStored.Status != domainpkg.TaskOutboxStatusDispatched {
		t.Fatalf("good outbox status = %q, want %q", goodStored.Status, domainpkg.TaskOutboxStatusDispatched)
	}
}

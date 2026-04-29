package articleinspect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
	"gorm.io/gorm"
)

var ErrTaskOutboxDispatcherUnavailable = errors.New("task outbox dispatcher unavailable")

type TaskOutboxDispatchReport struct {
	Scanned    int
	Dispatched int
	Failed     int
}

type TaskOutboxRelay struct {
	db         *gorm.DB
	dispatcher TaskDispatcher
	logger     *slog.Logger
}

func NewTaskOutboxRelay(db *gorm.DB, dispatcher TaskDispatcher, logger *slog.Logger) *TaskOutboxRelay {
	if db == nil {
		return nil
	}
	return &TaskOutboxRelay{db: db, dispatcher: dispatcher, logger: logger}
}

func (r *TaskOutboxRelay) CanDispatch() bool {
	return r != nil && r.db != nil && r.dispatcher != nil
}

func (r *TaskOutboxRelay) TryDispatchMessage(ctx context.Context, outboxID uint64) bool {
	if r == nil || r.db == nil {
		return false
	}
	if r.dispatcher == nil {
		r.logError("article inspect outbox dispatcher unavailable", "outbox_id", outboxID)
		return false
	}
	if err := r.DispatchMessage(ctx, outboxID); err != nil {
		r.logError("dispatch article inspect outbox message", "outbox_id", outboxID, "error", err)
		return false
	}
	return true
}

func (r *TaskOutboxRelay) DispatchMessage(ctx context.Context, outboxID uint64) error {
	if r == nil || r.db == nil || outboxID == 0 {
		return ErrInvalidTaskInput
	}
	if r.dispatcher == nil {
		return ErrTaskOutboxDispatcherUnavailable
	}

	var message InspectionTaskOutboxMessage
	if err := r.db.WithContext(ctx).
		Where("id = ?", outboxID).
		First(&message).Error; err != nil {
		return err
	}
	if strings.TrimSpace(message.Status) == TaskOutboxStatusDispatched {
		return nil
	}

	payload, err := decodeTaskOutboxPayload(message.Payload)
	if err != nil {
		if updateErr := r.recordDispatchFailure(ctx, message.ID, time.Now().UTC(), err); updateErr != nil {
			return updateErr
		}
		return err
	}
	attemptedAt := time.Now().UTC()
	if dispatchErr := r.dispatcher.DispatchArticleInspectTask(ctx, payload); dispatchErr != nil {
		if err := r.recordDispatchFailure(ctx, message.ID, attemptedAt, dispatchErr); err != nil {
			return err
		}
		return fmt.Errorf("dispatch article inspect task: %w", dispatchErr)
	}

	return r.db.WithContext(ctx).Model(&InspectionTaskOutboxMessage{}).
		Where("id = ?", message.ID).
		Updates(map[string]any{
			"status":          TaskOutboxStatusDispatched,
			"attempt_count":   gorm.Expr("attempt_count + 1"),
			"last_error":      "",
			"last_attempt_at": attemptedAt,
			"dispatched_at":   attemptedAt,
		}).Error
}

func (r *TaskOutboxRelay) DispatchPending(ctx context.Context, limit int) (TaskOutboxDispatchReport, error) {
	if r == nil || r.db == nil {
		return TaskOutboxDispatchReport{}, ErrInvalidTaskInput
	}
	if r.dispatcher == nil {
		return TaskOutboxDispatchReport{}, ErrTaskOutboxDispatcherUnavailable
	}
	if limit <= 0 {
		limit = 20
	}

	items := make([]InspectionTaskOutboxMessage, 0, limit)
	if err := r.db.WithContext(ctx).
		Where("status = ?", TaskOutboxStatusPending).
		Order("attempt_count ASC").
		Order("last_attempt_at ASC").
		Order("id ASC").
		Limit(limit).
		Find(&items).Error; err != nil {
		return TaskOutboxDispatchReport{}, err
	}

	report := TaskOutboxDispatchReport{Scanned: len(items)}
	for _, item := range items {
		if err := r.DispatchMessage(ctx, item.ID); err != nil {
			report.Failed++
			r.logError("dispatch article inspect outbox message", "outbox_id", item.ID, "error", err)
			continue
		}
		report.Dispatched++
	}
	return report, nil
}

func (r *TaskOutboxRelay) RelayPendingArticleInspectTaskOutbox(ctx context.Context, limit int) (int, error) {
	report, err := r.DispatchPending(ctx, limit)
	if err != nil {
		return 0, err
	}
	if report.Failed > 0 {
		r.logError("article inspect outbox relay completed with failures", "scanned", report.Scanned, "dispatched", report.Dispatched, "failed", report.Failed)
	}
	return report.Dispatched, nil
}

func decodeTaskOutboxPayload(raw string) (queuetasks.ArticleInspectTaskPayload, error) {
	var payload queuetasks.ArticleInspectTaskPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return queuetasks.ArticleInspectTaskPayload{}, err
	}
	return payload, nil
}

func (r *TaskOutboxRelay) logError(message string, args ...any) {
	if r == nil || r.logger == nil {
		return
	}
	r.logger.Error(message, args...)
}

func (r *TaskOutboxRelay) recordDispatchFailure(ctx context.Context, messageID uint64, attemptedAt time.Time, dispatchErr error) error {
	if r == nil || r.db == nil || messageID == 0 || dispatchErr == nil {
		return ErrInvalidTaskInput
	}
	return r.db.WithContext(ctx).Model(&InspectionTaskOutboxMessage{}).
		Where("id = ?", messageID).
		Updates(map[string]any{
			"attempt_count":   gorm.Expr("attempt_count + 1"),
			"last_error":      dispatchErr.Error(),
			"last_attempt_at": attemptedAt,
		}).Error
}

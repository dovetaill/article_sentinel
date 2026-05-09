package outbox

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	"gorm.io/gorm"
)

type TaskOutboxRelay struct {
	db         *gorm.DB
	dispatcher TaskDispatcher
	logger     *slog.Logger
	settings   TaskOutboxSettings
	claimOwner string
}

func NewTaskOutboxRelay(db *gorm.DB, dispatcher TaskDispatcher, logger *slog.Logger) *TaskOutboxRelay {
	if db == nil {
		return nil
	}
	return &TaskOutboxRelay{
		db:         db,
		dispatcher: dispatcher,
		logger:     logger,
		settings:   defaultTaskOutboxSettings(),
		claimOwner: defaultTaskOutboxClaimOwner(),
	}
}

func (r *TaskOutboxRelay) WithSettings(settings TaskOutboxSettings) *TaskOutboxRelay {
	if r == nil {
		return nil
	}
	if settings.LeaseDuration > 0 {
		r.settings.LeaseDuration = settings.LeaseDuration
	}
	if settings.MaxAttempts > 0 {
		r.settings.MaxAttempts = settings.MaxAttempts
	}
	if settings.DispatchedRetention > 0 {
		r.settings.DispatchedRetention = settings.DispatchedRetention
	}
	if settings.DeadLetterRetention > 0 {
		r.settings.DeadLetterRetention = settings.DeadLetterRetention
	}
	return r
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

	var message domainpkg.InspectionTaskOutboxMessage
	if err := r.db.WithContext(ctx).Where("id = ?", outboxID).First(&message).Error; err != nil {
		return err
	}

	now := time.Now().UTC()
	if !isDispatchableMessage(message, now) {
		return nil
	}

	claimed, claimedMessage, err := r.claimMessage(ctx, message, now)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}

	payload, err := decodeTaskOutboxPayload(claimedMessage.Payload)
	if err != nil {
		_, updateErr := r.moveToDeadLetter(ctx, claimedMessage, now, domainpkg.TaskOutboxErrorPayloadDecode, err)
		if updateErr != nil {
			return updateErr
		}
		return err
	}
	if strings.TrimSpace(claimedMessage.MessageType) != domainpkg.TaskOutboxMessageTypeRunTask {
		unsupportedErr := errors.New("unsupported outbox message type")
		_, updateErr := r.moveToDeadLetter(ctx, claimedMessage, now, domainpkg.TaskOutboxErrorUnsupportedMessage, unsupportedErr)
		if updateErr != nil {
			return updateErr
		}
		return unsupportedErr
	}

	if dispatchErr := r.dispatcher.DispatchArticleInspectTask(ctx, payload); dispatchErr != nil {
		errorCode := classifyDispatchError(dispatchErr)
		if r.shouldDeadLetter(claimedMessage, errorCode) {
			_, updateErr := r.moveToDeadLetter(ctx, claimedMessage, now, errorCode, dispatchErr)
			if updateErr != nil {
				return updateErr
			}
			return fmt.Errorf("dispatch article inspect task: %w", dispatchErr)
		}
		if _, updateErr := r.scheduleRetry(ctx, claimedMessage, now, errorCode, dispatchErr); updateErr != nil {
			return updateErr
		}
		return fmt.Errorf("dispatch article inspect task: %w", dispatchErr)
	}

	return r.markDispatched(ctx, claimedMessage, now)
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

	now := time.Now().UTC()
	items := make([]domainpkg.InspectionTaskOutboxMessage, 0, limit)
	if err := r.db.WithContext(ctx).
		Where(
			`(status = ? AND (next_attempt_at IS NULL OR next_attempt_at <= ?))
			  OR (status = ? AND (claim_until IS NULL OR claim_until < ?))`,
			domainpkg.TaskOutboxStatusPending,
			now,
			domainpkg.TaskOutboxStatusClaimed,
			now,
		).
		Order("attempt_count ASC").
		Order("id ASC").
		Limit(limit).
		Find(&items).Error; err != nil {
		return TaskOutboxDispatchReport{}, err
	}

	report := TaskOutboxDispatchReport{Scanned: len(items)}
	for _, item := range items {
		beforeAttempts := item.AttemptCount
		statusBefore := strings.TrimSpace(item.Status)
		if err := r.DispatchMessage(ctx, item.ID); err != nil {
			report.Failed++
			r.logError("dispatch article inspect outbox message", "outbox_id", item.ID, "error", err)
		}

		var updated domainpkg.InspectionTaskOutboxMessage
		if err := r.db.WithContext(ctx).Where("id = ?", item.ID).First(&updated).Error; err != nil {
			return report, err
		}
		if updated.AttemptCount > beforeAttempts {
			report.Claimed++
		}
		switch strings.TrimSpace(updated.Status) {
		case domainpkg.TaskOutboxStatusDispatched:
			report.Dispatched++
		case domainpkg.TaskOutboxStatusPending:
			if updated.AttemptCount > beforeAttempts || statusBefore == domainpkg.TaskOutboxStatusClaimed {
				report.Retried++
			}
		case domainpkg.TaskOutboxStatusDeadLetter:
			report.DeadLettered++
		}
	}
	return report, nil
}

func (r *TaskOutboxRelay) RelayPendingArticleInspectTaskOutbox(ctx context.Context, limit int) (int, error) {
	report, err := r.DispatchPending(ctx, limit)
	if err != nil {
		return 0, err
	}
	r.logInfo(
		"article inspect outbox relay finished",
		"scanned", report.Scanned,
		"claimed", report.Claimed,
		"dispatched", report.Dispatched,
		"retried", report.Retried,
		"dead_lettered", report.DeadLettered,
		"failed", report.Failed,
	)
	return report.Dispatched, nil
}

func (r *TaskOutboxRelay) claimMessage(ctx context.Context, message domainpkg.InspectionTaskOutboxMessage, now time.Time) (bool, domainpkg.InspectionTaskOutboxMessage, error) {
	if r == nil || r.db == nil {
		return false, domainpkg.InspectionTaskOutboxMessage{}, ErrInvalidTaskInput
	}
	if !isDispatchableMessage(message, now) {
		return false, message, nil
	}

	leaseUntil := now.Add(r.settings.LeaseDuration)
	query := r.db.WithContext(ctx).Model(&domainpkg.InspectionTaskOutboxMessage{}).Where("id = ?", message.ID)
	switch strings.TrimSpace(message.Status) {
	case domainpkg.TaskOutboxStatusPending:
		query = query.Where("status = ?", domainpkg.TaskOutboxStatusPending).
			Where("next_attempt_at IS NULL OR next_attempt_at <= ?", now)
	case domainpkg.TaskOutboxStatusClaimed:
		query = query.Where("status = ?", domainpkg.TaskOutboxStatusClaimed).
			Where("claim_until IS NULL OR claim_until < ?", now)
	default:
		return false, message, nil
	}

	result := query.Updates(map[string]any{
		"status":      domainpkg.TaskOutboxStatusClaimed,
		"claimed_by":  r.claimOwner,
		"claimed_at":  now,
		"claim_until": leaseUntil,
	})
	if result.Error != nil {
		return false, domainpkg.InspectionTaskOutboxMessage{}, result.Error
	}
	if result.RowsAffected == 0 {
		return false, message, nil
	}

	message.Status = domainpkg.TaskOutboxStatusClaimed
	message.ClaimedBy = r.claimOwner
	message.ClaimedAt = &now
	message.ClaimUntil = &leaseUntil
	return true, message, nil
}

func (r *TaskOutboxRelay) markDispatched(ctx context.Context, message domainpkg.InspectionTaskOutboxMessage, now time.Time) error {
	attemptCount := message.AttemptCount + 1
	retainedUntil := now.Add(r.settings.DispatchedRetention)
	return r.db.WithContext(ctx).Model(&domainpkg.InspectionTaskOutboxMessage{}).
		Where("id = ?", message.ID).
		Updates(map[string]any{
			"status":           domainpkg.TaskOutboxStatusDispatched,
			"attempt_count":    attemptCount,
			"claimed_by":       "",
			"claimed_at":       nil,
			"claim_until":      nil,
			"next_attempt_at":  nil,
			"last_error":       "",
			"last_error_code":  "",
			"last_attempt_at":  now,
			"dead_lettered_at": nil,
			"dispatched_at":    now,
			"retained_until":   retainedUntil,
		}).Error
}

func (r *TaskOutboxRelay) scheduleRetry(ctx context.Context, message domainpkg.InspectionTaskOutboxMessage, now time.Time, errorCode string, dispatchErr error) (int64, error) {
	attemptCount := message.AttemptCount + 1
	nextAttemptAt := now.Add(backoffForAttempt(attemptCount))
	errText := strings.TrimSpace(dispatchErr.Error())
	if errText == "" {
		errText = errorCode
	}
	return attemptCount, r.db.WithContext(ctx).Model(&domainpkg.InspectionTaskOutboxMessage{}).
		Where("id = ?", message.ID).
		Updates(map[string]any{
			"status":           domainpkg.TaskOutboxStatusPending,
			"attempt_count":    attemptCount,
			"claimed_by":       "",
			"claimed_at":       nil,
			"claim_until":      nil,
			"next_attempt_at":  nextAttemptAt,
			"last_error":       errText,
			"last_error_code":  errorCode,
			"last_attempt_at":  now,
			"dead_lettered_at": nil,
		}).Error
}

func (r *TaskOutboxRelay) moveToDeadLetter(ctx context.Context, message domainpkg.InspectionTaskOutboxMessage, now time.Time, errorCode string, dispatchErr error) (int64, error) {
	attemptCount := message.AttemptCount + 1
	retainedUntil := now.Add(r.settings.DeadLetterRetention)
	errText := strings.TrimSpace(dispatchErr.Error())
	if errText == "" {
		errText = errorCode
	}
	return attemptCount, r.db.WithContext(ctx).Model(&domainpkg.InspectionTaskOutboxMessage{}).
		Where("id = ?", message.ID).
		Updates(map[string]any{
			"status":           domainpkg.TaskOutboxStatusDeadLetter,
			"attempt_count":    attemptCount,
			"claimed_by":       "",
			"claimed_at":       nil,
			"claim_until":      nil,
			"next_attempt_at":  nil,
			"last_error":       errText,
			"last_error_code":  errorCode,
			"last_attempt_at":  now,
			"dead_lettered_at": now,
			"retained_until":   retainedUntil,
		}).Error
}

func (r *TaskOutboxRelay) shouldDeadLetter(message domainpkg.InspectionTaskOutboxMessage, errorCode string) bool {
	if errorCode == domainpkg.TaskOutboxErrorPayloadDecode || errorCode == domainpkg.TaskOutboxErrorUnsupportedMessage {
		return true
	}
	nextAttemptCount := message.AttemptCount + 1
	return r.settings.MaxAttempts > 0 && int(nextAttemptCount) >= r.settings.MaxAttempts
}

func classifyDispatchError(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrTaskOutboxDispatcherUnavailable):
		return domainpkg.TaskOutboxErrorDispatcherUnavailable
	default:
		return domainpkg.TaskOutboxErrorDispatch
	}
}

func backoffForAttempt(attemptCount int64) time.Duration {
	switch {
	case attemptCount <= 3:
		return 15 * time.Second
	case attemptCount <= 10:
		return time.Minute
	default:
		return 5 * time.Minute
	}
}

func (r *TaskOutboxRelay) logError(message string, args ...any) {
	if r == nil || r.logger == nil {
		return
	}
	r.logger.Error(message, args...)
}

func (r *TaskOutboxRelay) logInfo(message string, args ...any) {
	if r == nil || r.logger == nil {
		return
	}
	r.logger.Info(message, args...)
}

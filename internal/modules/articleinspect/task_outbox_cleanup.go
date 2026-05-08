package articleinspect

import (
	"context"
	"time"
)

func (r *TaskOutboxRelay) CleanupArticleInspectTaskOutbox(ctx context.Context, limit int) (int, error) {
	if r == nil || r.db == nil {
		return 0, ErrInvalidTaskInput
	}
	if limit <= 0 {
		limit = 100
	}

	now := time.Now().UTC()
	ids := make([]uint64, 0, limit)
	if err := r.db.WithContext(ctx).
		Model(&InspectionTaskOutboxMessage{}).
		Where(
			`(status = ? OR status = ?) AND retained_until IS NOT NULL AND retained_until < ?`,
			TaskOutboxStatusDispatched,
			TaskOutboxStatusDeadLetter,
			now,
		).
		Order("retained_until ASC").
		Order("id ASC").
		Limit(limit).
		Pluck("id", &ids).Error; err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}

	result := r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&InspectionTaskOutboxMessage{})
	if result.Error != nil {
		return 0, result.Error
	}
	r.logInfo("article inspect outbox cleanup finished", "deleted", result.RowsAffected)
	return int(result.RowsAffected), nil
}

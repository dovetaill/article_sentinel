package articleinspect

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ActionRepository struct {
	db *gorm.DB
}

func NewActionRepository(db *gorm.DB) *ActionRepository {
	return &ActionRepository{db: db}
}

func (r *ActionRepository) CreateAction(ctx context.Context, action *InspectionAction) error {
	if r == nil || r.db == nil || action == nil {
		return ErrInvalidActionInput
	}
	return r.db.WithContext(ctx).Create(action).Error
}

func (r *ActionRepository) UpdateActionSummary(ctx context.Context, actionID uint64, status string, successCount, failCount, skipCount int64) error {
	if r == nil || r.db == nil || actionID == 0 {
		return ErrInvalidActionInput
	}
	return r.db.WithContext(ctx).Model(&InspectionAction{}).
		Where("id = ?", actionID).
		Updates(map[string]any{
			"status":        status,
			"success_count": successCount,
			"fail_count":    failCount,
			"skip_count":    skipCount,
			"finished_at":   time.Now().UTC(),
		}).Error
}

func (r *ActionRepository) CreateOperationLog(ctx context.Context, log *InspectionOperationLog) error {
	if r == nil || r.db == nil || log == nil {
		return ErrInvalidActionInput
	}
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *ActionRepository) CreateFieldChangeLogs(ctx context.Context, logs []InspectionFieldChangeLog) error {
	if r == nil || r.db == nil {
		return ErrInvalidActionInput
	}
	if len(logs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&logs).Error
}

func buildActionNumber(now time.Time) string {
	return fmt.Sprintf("action-%s", now.Format("20060102150405.000000000"))
}

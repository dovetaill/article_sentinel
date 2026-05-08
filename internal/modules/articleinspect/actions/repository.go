package actions

import (
	"context"
	"fmt"
	"time"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
	"gorm.io/gorm"
)

type ActionRepository struct {
	db *gorm.DB
}

func NewActionRepository(db *gorm.DB) *ActionRepository {
	return &ActionRepository{db: db}
}

func (r *ActionRepository) CreateAction(ctx context.Context, action *domainpkg.InspectionAction) error {
	if r == nil || r.db == nil || action == nil {
		return sharedpkg.ErrInvalidActionInput
	}
	return r.db.WithContext(ctx).Create(action).Error
}

func (r *ActionRepository) UpdateActionSummary(ctx context.Context, actionID uint64, status string, successCount, failCount, skipCount int64) error {
	if r == nil || r.db == nil || actionID == 0 {
		return sharedpkg.ErrInvalidActionInput
	}
	return r.db.WithContext(ctx).Model(&domainpkg.InspectionAction{}).
		Where("id = ?", actionID).
		Updates(map[string]any{
			"status":        status,
			"success_count": successCount,
			"fail_count":    failCount,
			"skip_count":    skipCount,
			"finished_at":   time.Now().UTC(),
		}).Error
}

func buildActionNumber(now time.Time) string {
	return fmt.Sprintf("action-%s", now.Format("20060102150405.000000000"))
}

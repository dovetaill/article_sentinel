package articleinspect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrInvalidTaskInput = errors.New("invalid task input")

type taskKeywordRepository interface {
	ListByIDs(ctx context.Context, orgID uint64, ids []uint64) ([]InspectionKeyword, map[uint64][]InspectionKeywordScope, error)
}

type taskArticleRepository interface {
	ListCandidateArticles(ctx context.Context, filter CandidateArticleFilter) ([]CandidateArticle, uint64, error)
}

type TaskService struct {
	db       *gorm.DB
	keywords taskKeywordRepository
	articles taskArticleRepository
}

func NewTaskService(db *gorm.DB, keywords taskKeywordRepository, articles taskArticleRepository) *TaskService {
	return &TaskService{db: db, keywords: keywords, articles: articles}
}

func (s *TaskService) Create(ctx context.Context, input CreateInspectionTaskInput) (*InspectionTask, error) {
	if s == nil || s.db == nil || s.keywords == nil {
		return nil, ErrInvalidTaskInput
	}
	if input.OrgID == 0 {
		return nil, fmt.Errorf("%w: orgid is required", ErrInvalidTaskInput)
	}

	keywordIDs := uniqueUint64s(input.KeywordIDs)
	if len(keywordIDs) == 0 {
		return nil, fmt.Errorf("%w: keyword_ids are required", ErrInvalidTaskInput)
	}

	keywords, scopesByKeyword, err := s.keywords.ListByIDs(ctx, input.OrgID, keywordIDs)
	if err != nil {
		return nil, err
	}
	if len(keywords) != len(keywordIDs) {
		return nil, fmt.Errorf("%w: keyword_ids not found", ErrInvalidTaskInput)
	}

	articleState := input.ArticleState
	if articleState == 0 {
		articleState = ArticleStateOnline
	}

	requestSnapshot, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	ruleSnapshot, err := json.Marshal(buildKeywordSnapshots(keywords, scopesByKeyword))
	if err != nil {
		return nil, err
	}

	operatorID, operatorName := auditOperatorFromContext(ctx)
	now := time.Now().UTC()
	task := &InspectionTask{
		OrgID:              input.OrgID,
		TaskNo:             buildTaskNumber(now),
		Status:             TaskStatusPending,
		ArticleStateFilter: strconv.Itoa(int(articleState)),
		PublishTimeStart:   input.PublishTimeStart,
		PublishTimeEnd:     input.PublishTimeEnd,
		ArticleID:          input.ArticleID,
		TitleLike:          strings.TrimSpace(input.TitleLike),
		IncludeBody:        input.IncludeBody,
		RequestSnapshot:    string(requestSnapshot),
		RuleSnapshot:       string(ruleSnapshot),
		CreatorID:          operatorID,
		CreatorName:        operatorName,
	}

	taskKeywords := make([]InspectionTaskKeyword, 0, len(keywordIDs))
	for _, keywordID := range keywordIDs {
		taskKeywords = append(taskKeywords, InspectionTaskKeyword{OrgID: input.OrgID, KeywordID: keywordID})
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		for index := range taskKeywords {
			taskKeywords[index].TaskID = task.ID
		}
		return tx.Create(&taskKeywords).Error
	}); err != nil {
		return nil, err
	}
	return task, nil
}

func buildKeywordSnapshots(keywords []InspectionKeyword, scopesByKeyword map[uint64][]InspectionKeywordScope) []KeywordDTO {
	items := make([]KeywordDTO, 0, len(keywords))
	for _, keyword := range keywords {
		items = append(items, *buildKeywordDTO(&keyword, scopesByKeyword[keyword.ID]))
	}
	return items
}

func buildTaskNumber(now time.Time) string {
	return fmt.Sprintf("inspect-%s", now.Format("20060102150405.000000000"))
}

func uniqueUint64s(values []uint64) []uint64 {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[uint64]struct{}, len(values))
	result := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

package testutil

import (
	"encoding/json"
	"testing"
	"time"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	scanpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/scan"
	"gorm.io/gorm"
)

func SeedCandidateArticles(t *testing.T, db *gorm.DB) {
	t.Helper()

	articles := []domainpkg.Article{
		{ID: 1, OrgID: 100, Title: "Alpha news", State: domainpkg.ArticleStateOnline, PublishAtUnix: MustTime(t, "2026-04-20T10:00:00Z").Unix()},
		{ID: 2, OrgID: 100, Title: "Beta update", State: domainpkg.ArticleStateOnline, PublishAtUnix: MustTime(t, "2026-04-20T11:00:00Z").Unix()},
		{ID: 3, OrgID: 100, Title: "Gamma draft", State: domainpkg.ArticleStateDraft, PublishAtUnix: MustTime(t, "2026-04-20T12:00:00Z").Unix()},
		{ID: 4, OrgID: 200, Title: "Other org", State: domainpkg.ArticleStateOnline, PublishAtUnix: MustTime(t, "2026-04-20T10:30:00Z").Unix()},
	}
	if err := db.Create(&articles).Error; err != nil {
		t.Fatalf("seed articles error = %v", err)
	}

	infos := []domainpkg.ArticleInfo{
		{ID: 1, OrgID: 100, Body: "body one"},
		{ID: 2, OrgID: 100, Body: "body two"},
		{ID: 3, OrgID: 100, Body: "body three"},
		{ID: 4, OrgID: 200, Body: "body four"},
	}
	if err := db.Create(&infos).Error; err != nil {
		t.Fatalf("seed article infos error = %v", err)
	}
}

func SeedInspectionTaskForWorker(t *testing.T, db *gorm.DB, rules []scanpkg.KeywordRule) *domainpkg.InspectionTask {
	t.Helper()

	start := MustTime(t, "2026-04-20T09:00:00Z")
	end := MustTime(t, "2026-04-20T13:00:00Z")
	ruleSnapshot, err := marshalJSON(rules)
	if err != nil {
		t.Fatalf("marshal rule snapshot error = %v", err)
	}
	requestSnapshot, err := marshalJSON(map[string]any{
		"orgid":              uint64(100),
		"publish_time_start": start,
		"publish_time_end":   end,
		"include_body":       true,
	})
	if err != nil {
		t.Fatalf("marshal request snapshot error = %v", err)
	}

	task := &domainpkg.InspectionTask{
		OrgID:              100,
		TaskNo:             "inspect-test",
		Status:             domainpkg.TaskStatusPending,
		ArticleStateFilter: "9",
		PublishTimeStart:   timePointer(start),
		PublishTimeEnd:     timePointer(end),
		IncludeBody:        true,
		RequestSnapshot:    requestSnapshot,
		RuleSnapshot:       ruleSnapshot,
	}
	if err := db.Create(task).Error; err != nil {
		t.Fatalf("create task error = %v", err)
	}
	return task
}

func marshalJSON(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func timePointer(value time.Time) *time.Time {
	return &value
}

package testutil

import (
	"path/filepath"
	"testing"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewArticleInspectTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "articleinspect.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open() error = %v", err)
	}

	if err := db.AutoMigrate(
		&domainpkg.ChuangqiOrg{},
		&domainpkg.InspectionCategory{},
		&domainpkg.InspectionKeyword{},
		&domainpkg.InspectionKeywordScope{},
		&domainpkg.InspectionTask{},
		&domainpkg.InspectionTaskKeyword{},
		&domainpkg.InspectionTaskOutboxMessage{},
		&domainpkg.InspectionResult{},
		&domainpkg.InspectionResultHit{},
		&domainpkg.InspectionAction{},
		&domainpkg.InspectionOperationLog{},
		&domainpkg.InspectionFieldChangeLog{},
		&domainpkg.Article{},
		&domainpkg.ArticleInfo{},
	); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	return db
}

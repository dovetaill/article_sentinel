package articles

import (
	"context"
	"testing"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	taskspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/tasks"
	"github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
)

func TestArticleRepositoryListCandidateArticles(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedCandidateArticles(t, db)

	repo := NewArticleRepository(db)
	start := testutil.MustTime(t, "2026-04-20T09:00:00Z")
	end := testutil.MustTime(t, "2026-04-20T13:00:00Z")

	items, nextID, err := repo.ListCandidateArticles(context.Background(), taskspkg.CandidateArticleFilter{
		OrgID:            100,
		ArticleState:     domainpkg.ArticleStateOnline,
		PublishTimeStart: &start,
		PublishTimeEnd:   &end,
		Limit:            10,
	})
	if err != nil {
		t.Fatalf("ListCandidateArticles() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("ListCandidateArticles() len = %d, want %d", len(items), 2)
	}
	if nextID != 2 {
		t.Fatalf("ListCandidateArticles() nextID = %d, want %d", nextID, 2)
	}
	if items[0].Body != "body one" {
		t.Fatalf("ListCandidateArticles() first body = %q, want %q", items[0].Body, "body one")
	}
}

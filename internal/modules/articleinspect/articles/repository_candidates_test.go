package articles

import (
	"context"
	"testing"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	scanpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/scan"
	taskspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/tasks"
	"github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
)

func TestArticleRepositoryListCandidateArticles(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedCandidateArticles(t, db)

	repo := NewArticleRepository(db)
	start := testutil.MustTime(t, "2026-04-20T09:00:00Z")
	end := testutil.MustTime(t, "2026-04-20T13:00:00Z")

	firstPage, nextID, err := repo.ListCandidateArticles(context.Background(), taskspkg.CandidateArticleFilter{
		OrgID:            100,
		ArticleState:     domainpkg.ArticleStateOnline,
		PublishTimeStart: &start,
		PublishTimeEnd:   &end,
		Limit:            1,
	})
	if err != nil {
		t.Fatalf("ListCandidateArticles(first page) error = %v", err)
	}
	if len(firstPage) != 1 || firstPage[0].ID != 1 {
		t.Fatalf("first page ids = %#v, want first article only", extractCandidateArticleIDs(firstPage))
	}
	if firstPage[0].Body != "body one" {
		t.Fatalf("first page body = %q, want %q", firstPage[0].Body, "body one")
	}

	secondPage, _, err := repo.ListCandidateArticles(context.Background(), taskspkg.CandidateArticleFilter{
		OrgID:            100,
		ArticleState:     domainpkg.ArticleStateOnline,
		PublishTimeStart: &start,
		PublishTimeEnd:   &end,
		AfterID:          nextID,
		Limit:            1,
	})
	if err != nil {
		t.Fatalf("ListCandidateArticles(second page) error = %v", err)
	}
	if len(secondPage) != 1 || secondPage[0].ID != 2 {
		t.Fatalf("second page ids = %#v, want second article only", extractCandidateArticleIDs(secondPage))
	}

	exact, _, err := repo.ListCandidateArticles(context.Background(), taskspkg.CandidateArticleFilter{
		OrgID:        100,
		ArticleState: domainpkg.ArticleStateOnline,
		ArticleID:    2,
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("ListCandidateArticles(exact id) error = %v", err)
	}
	if len(exact) != 1 || exact[0].ID != 2 {
		t.Fatalf("exact filter ids = %#v, want article 2 only", extractCandidateArticleIDs(exact))
	}

	fuzzy, _, err := repo.ListCandidateArticles(context.Background(), taskspkg.CandidateArticleFilter{
		OrgID:        100,
		ArticleState: domainpkg.ArticleStateOnline,
		TitleLike:    "Alpha",
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("ListCandidateArticles(title like) error = %v", err)
	}
	if len(fuzzy) != 1 || fuzzy[0].ID != 1 {
		t.Fatalf("title filter ids = %#v, want article 1 only", extractCandidateArticleIDs(fuzzy))
	}
}

func extractCandidateArticleIDs(items []scanpkg.CandidateArticle) []uint64 {
	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

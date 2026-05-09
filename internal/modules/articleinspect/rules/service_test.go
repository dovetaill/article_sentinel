package rules

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/identity"
	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	testutil "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
)

func TestKeywordService(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	repo := NewKeywordRepository(db)
	service := NewKeywordService(repo)
	ctx := identity.ContextWithActor(context.Background(), identity.NewActor(7, "alice", "ops", "active"))

	created, err := service.Create(ctx, CreateKeywordInput{
		OrgID:         100,
		Name:          "spam",
		CategoryID:    1001,
		MatchType:     domainpkg.MatchTypeContains,
		RiskLevel:     domainpkg.RiskLevelHigh,
		SuggestAction: domainpkg.SuggestActionOffline,
		Enabled:       true,
		Remark:        "watch closely",
		Scopes:        []string{domainpkg.KeywordScopeBody, domainpkg.KeywordScopeTitle, domainpkg.KeywordScopeTitle},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.OrgID != 100 {
		t.Fatalf("Create().OrgID = %d, want %d", created.OrgID, 100)
	}
	if created.CategoryID != 1001 || created.CategoryName != "政策红线" {
		t.Fatalf("Create().Category = %d/%q, want %d/%q", created.CategoryID, created.CategoryName, 1001, "政策红线")
	}
	if !created.Enabled {
		t.Fatal("Create().Enabled = false, want true")
	}
	if created.CreatorID != 7 || created.UpdaterID != 7 {
		t.Fatalf("creator/updater ids = %d/%d, want %d/%d", created.CreatorID, created.UpdaterID, 7, 7)
	}
	if created.CreatorName != "alice" || created.UpdaterName != "alice" {
		t.Fatalf("creator/updater names = %q/%q, want %q/%q", created.CreatorName, created.UpdaterName, "alice", "alice")
	}
	if !reflect.DeepEqual(created.Scopes, []string{domainpkg.KeywordScopeBody, domainpkg.KeywordScopeTitle}) {
		t.Fatalf("Create().Scopes = %#v, want %#v", created.Scopes, []string{domainpkg.KeywordScopeBody, domainpkg.KeywordScopeTitle})
	}

	if _, err := service.Create(ctx, CreateKeywordInput{
		OrgID:         200,
		Name:          "spam",
		CategoryID:    2001,
		MatchType:     domainpkg.MatchTypeContains,
		RiskLevel:     domainpkg.RiskLevelLow,
		SuggestAction: domainpkg.SuggestActionIgnore,
		Enabled:       true,
		Scopes:        []string{domainpkg.KeywordScopeTitle},
	}); err != nil {
		t.Fatalf("Create(other org) error = %v", err)
	}

	listed, err := service.List(ctx, KeywordListInput{OrgID: 100, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if listed.Total != 1 {
		t.Fatalf("List().Total = %d, want %d", listed.Total, 1)
	}
	if len(listed.Items) != 1 {
		t.Fatalf("List().Items len = %d, want %d", len(listed.Items), 1)
	}
	if listed.Items[0].Name != "spam" {
		t.Fatalf("List().Items[0].Name = %q, want %q", listed.Items[0].Name, "spam")
	}

	disabled, err := service.PatchEnabled(ctx, PatchKeywordStatusInput{OrgID: 100, KeywordID: created.ID, Enabled: false})
	if err != nil {
		t.Fatalf("PatchEnabled(false) error = %v", err)
	}
	if disabled.Enabled {
		t.Fatal("PatchEnabled(false).Enabled = true, want false")
	}

	enabled, err := service.PatchEnabled(ctx, PatchKeywordStatusInput{OrgID: 100, KeywordID: created.ID, Enabled: true})
	if err != nil {
		t.Fatalf("PatchEnabled(true) error = %v", err)
	}
	if !enabled.Enabled {
		t.Fatal("PatchEnabled(true).Enabled = false, want true")
	}
}

func TestKeywordValidation(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	service := NewKeywordService(NewKeywordRepository(db))

	tests := []struct {
		name  string
		input CreateKeywordInput
	}{
		{name: "missing orgid", input: CreateKeywordInput{Name: "spam", CategoryID: 1001, MatchType: domainpkg.MatchTypeContains, RiskLevel: domainpkg.RiskLevelHigh, SuggestAction: domainpkg.SuggestActionOffline, Enabled: true, Scopes: []string{domainpkg.KeywordScopeTitle}}},
		{name: "missing category id", input: CreateKeywordInput{OrgID: 100, Name: "spam", MatchType: domainpkg.MatchTypeContains, RiskLevel: domainpkg.RiskLevelHigh, SuggestAction: domainpkg.SuggestActionOffline, Enabled: true, Scopes: []string{domainpkg.KeywordScopeTitle}}},
		{name: "unsupported scope", input: CreateKeywordInput{OrgID: 100, Name: "spam", CategoryID: 1001, MatchType: domainpkg.MatchTypeContains, RiskLevel: domainpkg.RiskLevelHigh, SuggestAction: domainpkg.SuggestActionOffline, Enabled: true, Scopes: []string{"summary"}}},
		{name: "unsupported risk", input: CreateKeywordInput{OrgID: 100, Name: "spam", CategoryID: 1001, MatchType: domainpkg.MatchTypeContains, RiskLevel: "critical", SuggestAction: domainpkg.SuggestActionOffline, Enabled: true, Scopes: []string{domainpkg.KeywordScopeTitle}}},
		{name: "unsupported action", input: CreateKeywordInput{OrgID: 100, Name: "spam", CategoryID: 1001, MatchType: domainpkg.MatchTypeContains, RiskLevel: domainpkg.RiskLevelHigh, SuggestAction: "ban", Enabled: true, Scopes: []string{domainpkg.KeywordScopeTitle}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Create(context.Background(), tt.input)
			if !errors.Is(err, ErrInvalidKeywordInput) {
				t.Fatalf("Create() error = %v, want %v", err, ErrInvalidKeywordInput)
			}
		})
	}
}

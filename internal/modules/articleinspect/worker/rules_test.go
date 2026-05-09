package worker

import (
	"encoding/json"
	"testing"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
)

func TestDecodeTaskRulesFromLegacyKeywordSnapshotJSON(t *testing.T) {
	type legacyKeywordSnapshotPayload struct {
		ID            uint64   `json:"id"`
		OrgID         uint64   `json:"orgid"`
		Name          string   `json:"name"`
		CategoryID    uint64   `json:"category_id"`
		CategoryName  string   `json:"category_name"`
		MatchType     string   `json:"match_type"`
		RiskLevel     string   `json:"risk_level"`
		SuggestAction string   `json:"suggest_action"`
		Enabled       bool     `json:"enabled"`
		Scopes        []string `json:"scopes"`
	}

	snapshot, err := json.Marshal([]legacyKeywordSnapshotPayload{
		{
			ID:            9101013,
			OrgID:         29,
			Name:          "svg",
			CategoryID:    502,
			CategoryName:  "高频违规",
			MatchType:     domainpkg.MatchTypeContains,
			RiskLevel:     domainpkg.RiskLevelHigh,
			SuggestAction: domainpkg.SuggestActionOffline,
			Enabled:       true,
			Scopes:        []string{domainpkg.KeywordScopeTitle},
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	rules, err := DecodeTaskRules(string(snapshot))
	if err != nil {
		t.Fatalf("DecodeTaskRules() error = %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("DecodeTaskRules() len = %d, want %d", len(rules), 1)
	}
	if rules[0].Name != "svg" || rules[0].MatchType != domainpkg.MatchTypeContains || rules[0].RiskLevel != domainpkg.RiskLevelHigh || rules[0].SuggestAction != domainpkg.SuggestActionOffline {
		t.Fatalf("DecodeTaskRules() rule = %+v, want populated match metadata", rules[0])
	}
	if len(rules[0].Scopes) != 1 || rules[0].Scopes[0] != domainpkg.KeywordScopeTitle {
		t.Fatalf("DecodeTaskRules() scopes = %#v, want %#v", rules[0].Scopes, []string{domainpkg.KeywordScopeTitle})
	}
}

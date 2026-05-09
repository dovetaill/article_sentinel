package articleinspect

import (
	"encoding/json"
	"testing"
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
			MatchType:     MatchTypeContains,
			RiskLevel:     RiskLevelHigh,
			SuggestAction: SuggestActionOffline,
			Enabled:       true,
			Scopes:        []string{KeywordScopeTitle},
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	rules, err := decodeTaskRules(string(snapshot))
	if err != nil {
		t.Fatalf("decodeTaskRules() error = %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("decodeTaskRules() len = %d, want %d", len(rules), 1)
	}
	if rules[0].Name != "svg" || rules[0].MatchType != MatchTypeContains || rules[0].RiskLevel != RiskLevelHigh || rules[0].SuggestAction != SuggestActionOffline {
		t.Fatalf("decodeTaskRules() rule = %+v, want populated match metadata", rules[0])
	}
	if len(rules[0].Scopes) != 1 || rules[0].Scopes[0] != KeywordScopeTitle {
		t.Fatalf("decodeTaskRules() scopes = %#v, want %#v", rules[0].Scopes, []string{KeywordScopeTitle})
	}
}

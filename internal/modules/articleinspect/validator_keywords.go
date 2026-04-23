package articleinspect

import (
	"fmt"
	"sort"
	"strings"
)

type normalizedKeywordInput struct {
	orgID         uint64
	name          string
	categoryID    uint64
	matchType     string
	riskLevel     string
	suggestAction string
	enabled       bool
	remark        string
	scopes        []string
}

func normalizeCreateKeywordInput(input CreateKeywordInput) (*normalizedKeywordInput, error) {
	if input.OrgID == 0 {
		return nil, fmt.Errorf("%w: orgid is required", ErrInvalidKeywordInput)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidKeywordInput)
	}

	if input.CategoryID == 0 {
		return nil, fmt.Errorf("%w: category_id is required", ErrInvalidKeywordInput)
	}

	matchType := strings.TrimSpace(strings.ToLower(input.MatchType))
	if !containsString(InspectionMatchTypes(), matchType) {
		return nil, fmt.Errorf("%w: unsupported match type %q", ErrInvalidKeywordInput, input.MatchType)
	}

	riskLevel := strings.TrimSpace(strings.ToLower(input.RiskLevel))
	if !containsString(InspectionRiskLevels(), riskLevel) {
		return nil, fmt.Errorf("%w: unsupported risk level %q", ErrInvalidKeywordInput, input.RiskLevel)
	}

	suggestAction := strings.TrimSpace(strings.ToLower(input.SuggestAction))
	if !containsString(InspectionSuggestActions(), suggestAction) {
		return nil, fmt.Errorf("%w: unsupported suggest action %q", ErrInvalidKeywordInput, input.SuggestAction)
	}

	scopes, err := normalizeKeywordScopes(input.Scopes)
	if err != nil {
		return nil, err
	}

	return &normalizedKeywordInput{
		orgID:         input.OrgID,
		name:          name,
		categoryID:    input.CategoryID,
		matchType:     matchType,
		riskLevel:     riskLevel,
		suggestAction: suggestAction,
		enabled:       input.Enabled,
		remark:        strings.TrimSpace(input.Remark),
		scopes:        scopes,
	}, nil
}

func normalizeKeywordScopes(scopes []string) ([]string, error) {
	if len(scopes) == 0 {
		return nil, fmt.Errorf("%w: scopes are required", ErrInvalidKeywordInput)
	}

	seen := make(map[string]struct{}, len(scopes))
	normalized := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(strings.ToLower(scope))
		if !containsString(InspectionKeywordScopes(), scope) {
			return nil, fmt.Errorf("%w: unsupported scope %q", ErrInvalidKeywordInput, scope)
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		normalized = append(normalized, scope)
	}
	sort.Strings(normalized)
	return normalized, nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

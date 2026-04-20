package articleinspect

import (
	"context"
	"errors"
	"sort"
	"strings"
)

var (
	ErrKeywordNotFound     = errors.New("keyword not found")
	ErrInvalidKeywordInput = errors.New("invalid keyword input")
)

type keywordRepository interface {
	Create(ctx context.Context, keyword *InspectionKeyword, scopes []InspectionKeywordScope) error
	Update(ctx context.Context, keyword *InspectionKeyword, scopes []InspectionKeywordScope) error
	Delete(ctx context.Context, orgID, id uint64) error
	Get(ctx context.Context, orgID, id uint64) (*InspectionKeyword, []InspectionKeywordScope, error)
	List(ctx context.Context, filter KeywordListFilter) ([]InspectionKeyword, map[uint64][]InspectionKeywordScope, int64, error)
	PatchEnabled(ctx context.Context, orgID, id uint64, enabled bool, updaterID uint64, updaterName string) error
	ListByIDs(ctx context.Context, orgID uint64, ids []uint64) ([]InspectionKeyword, map[uint64][]InspectionKeywordScope, error)
}

type KeywordService struct {
	repo keywordRepository
}

func NewKeywordService(repo keywordRepository) *KeywordService {
	return &KeywordService{repo: repo}
}

func (s *KeywordService) Create(ctx context.Context, input CreateKeywordInput) (*KeywordDTO, error) {
	if s == nil || s.repo == nil {
		return nil, ErrInvalidKeywordInput
	}

	normalized, err := normalizeCreateKeywordInput(input)
	if err != nil {
		return nil, err
	}
	operatorID, operatorName := auditOperatorFromContext(ctx)
	keyword := &InspectionKeyword{
		OrgID:         normalized.orgID,
		Name:          normalized.name,
		Category:      normalized.category,
		MatchType:     normalized.matchType,
		RiskLevel:     normalized.riskLevel,
		SuggestAction: normalized.suggestAction,
		Enabled:       normalized.enabled,
		Remark:        normalized.remark,
		CreatorID:     operatorID,
		CreatorName:   operatorName,
		UpdaterID:     operatorID,
		UpdaterName:   operatorName,
	}
	if err := s.repo.Create(ctx, keyword, buildKeywordScopes(normalized.orgID, normalized.scopes)); err != nil {
		return nil, err
	}
	return buildKeywordDTO(keyword, buildKeywordScopes(normalized.orgID, normalized.scopes)), nil
}

func (s *KeywordService) Update(ctx context.Context, input UpdateKeywordInput) (*KeywordDTO, error) {
	if s == nil || s.repo == nil || input.ID == 0 {
		return nil, ErrInvalidKeywordInput
	}

	normalized, err := normalizeCreateKeywordInput(input.CreateKeywordInput)
	if err != nil {
		return nil, err
	}
	operatorID, operatorName := auditOperatorFromContext(ctx)
	keyword := &InspectionKeyword{
		ID:            input.ID,
		OrgID:         normalized.orgID,
		Name:          normalized.name,
		Category:      normalized.category,
		MatchType:     normalized.matchType,
		RiskLevel:     normalized.riskLevel,
		SuggestAction: normalized.suggestAction,
		Enabled:       normalized.enabled,
		Remark:        normalized.remark,
		UpdaterID:     operatorID,
		UpdaterName:   operatorName,
	}
	if err := s.repo.Update(ctx, keyword, buildKeywordScopes(normalized.orgID, normalized.scopes)); err != nil {
		return nil, err
	}
	stored, scopes, err := s.repo.Get(ctx, normalized.orgID, input.ID)
	if err != nil {
		return nil, err
	}
	return buildKeywordDTO(stored, scopes), nil
}

func (s *KeywordService) Delete(ctx context.Context, orgID, id uint64) error {
	if s == nil || s.repo == nil || orgID == 0 || id == 0 {
		return ErrInvalidKeywordInput
	}
	return s.repo.Delete(ctx, orgID, id)
}

func (s *KeywordService) Get(ctx context.Context, orgID, id uint64) (*KeywordDTO, error) {
	if s == nil || s.repo == nil || orgID == 0 || id == 0 {
		return nil, ErrInvalidKeywordInput
	}
	keyword, scopes, err := s.repo.Get(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	return buildKeywordDTO(keyword, scopes), nil
}

func (s *KeywordService) List(ctx context.Context, input KeywordListInput) (*KeywordListResult, error) {
	if s == nil || s.repo == nil || input.OrgID == 0 {
		return nil, ErrInvalidKeywordInput
	}
	page, pageSize := normalizePage(input.Page, input.PageSize)
	items, scopesByKeyword, total, err := s.repo.List(ctx, KeywordListFilter{
		OrgID:    input.OrgID,
		Page:     page,
		PageSize: pageSize,
		Enabled:  input.Enabled,
		Category: strings.TrimSpace(input.Category),
		Query:    strings.TrimSpace(input.Query),
	})
	if err != nil {
		return nil, err
	}

	result := &KeywordListResult{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Items:    make([]KeywordDTO, 0, len(items)),
	}
	for _, item := range items {
		result.Items = append(result.Items, *buildKeywordDTO(&item, scopesByKeyword[item.ID]))
	}
	return result, nil
}

func (s *KeywordService) PatchEnabled(ctx context.Context, input PatchKeywordStatusInput) (*KeywordDTO, error) {
	if s == nil || s.repo == nil || input.OrgID == 0 || input.KeywordID == 0 {
		return nil, ErrInvalidKeywordInput
	}
	operatorID, operatorName := auditOperatorFromContext(ctx)
	if err := s.repo.PatchEnabled(ctx, input.OrgID, input.KeywordID, input.Enabled, operatorID, operatorName); err != nil {
		return nil, err
	}
	keyword, scopes, err := s.repo.Get(ctx, input.OrgID, input.KeywordID)
	if err != nil {
		return nil, err
	}
	return buildKeywordDTO(keyword, scopes), nil
}

func buildKeywordScopes(orgID uint64, scopes []string) []InspectionKeywordScope {
	items := make([]InspectionKeywordScope, 0, len(scopes))
	for _, scope := range scopes {
		items = append(items, InspectionKeywordScope{OrgID: orgID, Scope: scope})
	}
	return items
}

func buildKeywordDTO(keyword *InspectionKeyword, scopes []InspectionKeywordScope) *KeywordDTO {
	if keyword == nil {
		return nil
	}
	values := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		values = append(values, scope.Scope)
	}
	sort.Strings(values)
	return &KeywordDTO{
		ID:            keyword.ID,
		OrgID:         keyword.OrgID,
		Name:          keyword.Name,
		Category:      keyword.Category,
		MatchType:     keyword.MatchType,
		RiskLevel:     keyword.RiskLevel,
		SuggestAction: keyword.SuggestAction,
		Enabled:       keyword.Enabled,
		Remark:        keyword.Remark,
		Scopes:        values,
		CreatorID:     keyword.CreatorID,
		CreatorName:   keyword.CreatorName,
		UpdaterID:     keyword.UpdaterID,
		UpdaterName:   keyword.UpdaterName,
	}
}

func auditOperatorFromContext(ctx context.Context) (uint64, string) {
	operator := ResolveOperator(ctx)
	return operator.ID, operator.Name
}

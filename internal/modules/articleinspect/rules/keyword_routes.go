package rules

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

type keywordQueryRequest struct {
	OrgID      sharedpkg.Uint64Param         `query:"orgid"`
	Page       sharedpkg.OptionalIntParam    `query:"page"`
	PageSize   sharedpkg.OptionalIntParam    `query:"page_size"`
	CategoryID sharedpkg.OptionalUint64Param `query:"category_id"`
	Query      string                        `query:"keyword"`
	Enabled    sharedpkg.OptionalBoolParam   `query:"enabled"`
}

type keywordBody struct {
	OrgID         uint64   `json:"orgid,omitempty"`
	Name          string   `json:"name,omitempty"`
	CategoryID    uint64   `json:"category_id,omitempty"`
	MatchType     string   `json:"match_type,omitempty"`
	RiskLevel     string   `json:"risk_level,omitempty"`
	SuggestAction string   `json:"suggest_action,omitempty"`
	Enabled       bool     `json:"enabled,omitempty"`
	Remark        string   `json:"remark,omitempty"`
	Scopes        []string `json:"scopes,omitempty"`
}

type keywordCreateRequest struct {
	Body keywordBody
}

type keywordDetailRequest struct {
	ID    sharedpkg.Uint64Param `path:"id"`
	OrgID sharedpkg.Uint64Param `query:"orgid"`
}

type keywordUpdateRequest struct {
	ID   sharedpkg.Uint64Param `path:"id"`
	Body keywordBody
}

type keywordPatchStatusBody struct {
	OrgID   uint64 `json:"orgid,omitempty"`
	Enabled bool   `json:"enabled"`
}

type keywordPatchStatusRequest struct {
	ID   sharedpkg.Uint64Param `path:"id"`
	Body keywordPatchStatusBody
}

func RegisterKeywordRoutes(api huma.API, service *KeywordService) {
	huma.Register(api, huma.Operation{
		OperationID:   "article-inspect-keyword-create",
		Method:        http.MethodPost,
		Path:          "/keywords",
		Summary:       "create inspection keyword",
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, input *keywordCreateRequest) (*sharedpkg.CreatedEnvelopeOutput, error) {
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureCreatedFromError(err)
		}
		item, err := service.Create(ctx, CreateKeywordInput{
			OrgID:         orgID,
			Name:          input.Body.Name,
			CategoryID:    input.Body.CategoryID,
			MatchType:     input.Body.MatchType,
			RiskLevel:     input.Body.RiskLevel,
			SuggestAction: input.Body.SuggestAction,
			Enabled:       input.Body.Enabled,
			Remark:        input.Body.Remark,
			Scopes:        input.Body.Scopes,
		})
		if err != nil {
			return failureCreatedFromError(err)
		}
		return sharedpkg.SuccessCreatedEnvelope("keyword created", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-keyword-list",
		Method:             http.MethodGet,
		Path:               "/keywords",
		Summary:            "list inspection keywords",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *keywordQueryRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid keyword input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		page, err := input.Page.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid keyword input"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid keyword input"), nil
		}
		categoryID, err := input.CategoryID.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid keyword input"), nil
		}
		enabled, err := input.Enabled.Ptr()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid enabled filter"), nil
		}
		result, err := service.List(ctx, KeywordListInput{
			OrgID:      orgID,
			Page:       page,
			PageSize:   pageSize,
			Enabled:    enabled,
			CategoryID: categoryID,
			Query:      input.Query,
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "keyword list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-keyword-detail",
		Method:             http.MethodGet,
		Path:               "/keywords/{id}",
		Summary:            "get inspection keyword",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *keywordDetailRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid keyword input"), nil
		}
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid keyword input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		item, err := service.Get(ctx, orgID, id)
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "keyword detail", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-keyword-update",
		Method:             http.MethodPut,
		Path:               "/keywords/{id}",
		Summary:            "update inspection keyword",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *keywordUpdateRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid keyword input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		item, err := service.Update(ctx, UpdateKeywordInput{
			ID: id,
			CreateKeywordInput: CreateKeywordInput{
				OrgID:         orgID,
				Name:          input.Body.Name,
				CategoryID:    input.Body.CategoryID,
				MatchType:     input.Body.MatchType,
				RiskLevel:     input.Body.RiskLevel,
				SuggestAction: input.Body.SuggestAction,
				Enabled:       input.Body.Enabled,
				Remark:        input.Body.Remark,
				Scopes:        input.Body.Scopes,
			},
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "keyword updated", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-keyword-status-patch",
		Method:             http.MethodPatch,
		Path:               "/keywords/{id}/status",
		Summary:            "patch inspection keyword status",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *keywordPatchStatusRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid keyword input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		item, err := service.PatchEnabled(ctx, PatchKeywordStatusInput{
			OrgID:     orgID,
			KeywordID: id,
			Enabled:   input.Body.Enabled,
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "keyword status updated", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-keyword-delete",
		Method:             http.MethodDelete,
		Path:               "/keywords/{id}",
		Summary:            "delete inspection keyword",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *keywordDetailRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid keyword input"), nil
		}
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid keyword input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		if err := service.Delete(ctx, orgID, id); err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "keyword deleted", map[string]uint64{"id": id}), nil
	})
}

package rules

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

type orgListRequest struct{}

type categoryQueryRequest struct {
	OrgID    sharedpkg.Uint64Param       `query:"orgid"`
	Page     sharedpkg.OptionalIntParam  `query:"page"`
	PageSize sharedpkg.OptionalIntParam  `query:"page_size"`
	Query    string                      `query:"name"`
	Enabled  sharedpkg.OptionalBoolParam `query:"enabled"`
}

type categoryDetailRequest struct {
	ID    sharedpkg.Uint64Param `path:"id"`
	OrgID sharedpkg.Uint64Param `query:"orgid"`
}

type categoryBody struct {
	OrgID   uint64 `json:"orgid,omitempty"`
	Name    string `json:"name,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
	Sort    int64  `json:"sort,omitempty"`
}

type categoryCreateRequest struct {
	Body categoryBody
}

type categoryUpdateRequest struct {
	ID   sharedpkg.Uint64Param `path:"id"`
	Body categoryBody
}

type categoryPatchStatusBody struct {
	OrgID   uint64 `json:"orgid,omitempty"`
	Enabled bool   `json:"enabled"`
}

type categoryPatchStatusRequest struct {
	ID   sharedpkg.Uint64Param `path:"id"`
	Body categoryPatchStatusBody
}

func RegisterCategoryRoutes(api huma.API, service *CategoryService) {
	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-org-list",
		Method:      http.MethodGet,
		Path:        "/orgs",
		Summary:     "list inspection organizations",
	}, func(ctx context.Context, input *orgListRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		_ = input
		result, err := service.ListOrgs(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "org list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-category-list",
		Method:             http.MethodGet,
		Path:               "/categories",
		Summary:            "list inspection categories",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *categoryQueryRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		page, err := input.Page.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		enabled, err := input.Enabled.Ptr()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid enabled filter"), nil
		}
		result, err := service.List(ctx, CategoryListInput{
			OrgID:    orgID,
			Page:     page,
			PageSize: pageSize,
			Enabled:  enabled,
			Query:    input.Query,
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "category list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "article-inspect-category-create",
		Method:        http.MethodPost,
		Path:          "/categories",
		Summary:       "create inspection category",
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, input *categoryCreateRequest) (*sharedpkg.CreatedEnvelopeOutput, error) {
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureCreatedFromError(err)
		}
		item, err := service.Create(ctx, CreateCategoryInput{
			OrgID:   orgID,
			Name:    input.Body.Name,
			Enabled: input.Body.Enabled,
			Sort:    input.Body.Sort,
		})
		if err != nil {
			return failureCreatedFromError(err)
		}
		return sharedpkg.SuccessCreatedEnvelope("category created", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-category-detail",
		Method:             http.MethodGet,
		Path:               "/categories/{id}",
		Summary:            "get inspection category",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *categoryDetailRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		item, err := service.Get(ctx, orgID, id)
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "category detail", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-category-update",
		Method:             http.MethodPut,
		Path:               "/categories/{id}",
		Summary:            "update inspection category",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *categoryUpdateRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		item, err := service.Update(ctx, UpdateCategoryInput{
			ID: id,
			CreateCategoryInput: CreateCategoryInput{
				OrgID:   orgID,
				Name:    input.Body.Name,
				Enabled: input.Body.Enabled,
				Sort:    input.Body.Sort,
			},
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "category updated", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-category-status-patch",
		Method:             http.MethodPatch,
		Path:               "/categories/{id}/status",
		Summary:            "patch inspection category status",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *categoryPatchStatusRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		item, err := service.PatchEnabled(ctx, PatchCategoryStatusInput{
			OrgID:      orgID,
			CategoryID: id,
			Enabled:    input.Body.Enabled,
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "category status updated", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-category-delete",
		Method:             http.MethodDelete,
		Path:               "/categories/{id}",
		Summary:            "delete inspection category",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *categoryDetailRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		if err := service.Delete(ctx, orgID, id); err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "category deleted", map[string]uint64{"id": id}), nil
	})
}

package articleinspect

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type orgListRequest struct{}

type categoryQueryRequest struct {
	OrgID    uint64Param       `query:"orgid"`
	Page     optionalIntParam  `query:"page"`
	PageSize optionalIntParam  `query:"page_size"`
	Query    string            `query:"name"`
	Enabled  optionalBoolParam `query:"enabled"`
}

type categoryDetailRequest struct {
	ID    uint64Param `path:"id"`
	OrgID uint64Param `query:"orgid"`
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
	ID   uint64Param `path:"id"`
	Body categoryBody
}

type categoryPatchStatusBody struct {
	OrgID   uint64 `json:"orgid"`
	Enabled bool   `json:"enabled"`
}

type categoryPatchStatusRequest struct {
	ID   uint64Param `path:"id"`
	Body categoryPatchStatusBody
}

func registerCategoryRoutes(api huma.API, service *CategoryService) {
	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-org-list",
		Method:      http.MethodGet,
		Path:        "/orgs",
		Summary:     "list inspection organizations",
	}, func(ctx context.Context, input *orgListRequest) (*okEnvelopeOutput, error) {
		_ = input
		result, err := service.ListOrgs(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		return successOKEnvelope(http.StatusOK, "org list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-category-list",
		Method:             http.MethodGet,
		Path:               "/categories",
		Summary:            "list inspection categories",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *categoryQueryRequest) (*okEnvelopeOutput, error) {
		if err := validateOptionalOrgID(input.OrgID); err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		orgID, err := currentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		page, err := input.Page.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		enabled, err := input.Enabled.Ptr()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid enabled filter"), nil
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
		return successOKEnvelope(http.StatusOK, "category list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "article-inspect-category-create",
		Method:        http.MethodPost,
		Path:          "/categories",
		Summary:       "create inspection category",
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, input *categoryCreateRequest) (*createdEnvelopeOutput, error) {
		orgID, err := currentOrgID(ctx)
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
		return successCreatedEnvelope("category created", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-category-detail",
		Method:             http.MethodGet,
		Path:               "/categories/{id}",
		Summary:            "get inspection category",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *categoryDetailRequest) (*okEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		if err := validateOptionalOrgID(input.OrgID); err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		orgID, err := currentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		item, err := service.Get(ctx, orgID, id)
		if err != nil {
			return failureOKFromError(err)
		}
		return successOKEnvelope(http.StatusOK, "category detail", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-category-update",
		Method:             http.MethodPut,
		Path:               "/categories/{id}",
		Summary:            "update inspection category",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *categoryUpdateRequest) (*okEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		orgID, err := currentOrgID(ctx)
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
		return successOKEnvelope(http.StatusOK, "category updated", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-category-status-patch",
		Method:             http.MethodPatch,
		Path:               "/categories/{id}/status",
		Summary:            "patch inspection category status",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *categoryPatchStatusRequest) (*okEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		orgID, err := currentOrgID(ctx)
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
		return successOKEnvelope(http.StatusOK, "category status updated", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-category-delete",
		Method:             http.MethodDelete,
		Path:               "/categories/{id}",
		Summary:            "delete inspection category",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *categoryDetailRequest) (*okEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		if err := validateOptionalOrgID(input.OrgID); err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		orgID, err := currentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		if err := service.Delete(ctx, orgID, id); err != nil {
			return failureOKFromError(err)
		}
		return successOKEnvelope(http.StatusOK, "category deleted", map[string]uint64{"id": id}), nil
	})
}

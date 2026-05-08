package results

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

type resultListRequest struct {
	OrgID             sharedpkg.Uint64Param         `query:"orgid"`
	TaskID            sharedpkg.OptionalUint64Param `query:"task_id"`
	RiskLevel         string                        `query:"risk_level"`
	DispositionStatus string                        `query:"disposition_status"`
	TitleLike         string                        `query:"title"`
	ArticleID         sharedpkg.OptionalUint64Param `query:"article_id"`
	Page              sharedpkg.OptionalIntParam    `query:"page"`
	PageSize          sharedpkg.OptionalIntParam    `query:"page_size"`
}

type resultDetailRequest struct {
	ID    sharedpkg.Uint64Param `path:"id"`
	OrgID sharedpkg.Uint64Param `query:"orgid"`
}

func RegisterResultRoutes(api huma.API, service *ResultService) {
	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-result-list",
		Method:             http.MethodGet,
		Path:               "/results",
		Summary:            "list article inspection results",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *resultListRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid result query"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		taskID, err := input.TaskID.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid result query"), nil
		}
		articleID, err := input.ArticleID.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid result query"), nil
		}
		page, err := input.Page.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid result query"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid result query"), nil
		}
		result, err := service.List(ctx, ResultListInput{
			OrgID:             orgID,
			TaskID:            taskID,
			RiskLevel:         input.RiskLevel,
			DispositionStatus: input.DispositionStatus,
			TitleLike:         input.TitleLike,
			ArticleID:         articleID,
			Page:              page,
			PageSize:          pageSize,
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "result list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-result-detail",
		Method:             http.MethodGet,
		Path:               "/results/{id}",
		Summary:            "get article inspection result detail",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *resultDetailRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid result input"), nil
		}
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid result input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		result, err := service.GetDetail(ctx, orgID, id)
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "result detail", result), nil
	})
}

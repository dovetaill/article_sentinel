package articleinspect

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type resultListRequest struct {
	OrgID             uint64Param         `query:"orgid"`
	TaskID            optionalUint64Param `query:"task_id"`
	RiskLevel         string              `query:"risk_level"`
	DispositionStatus string              `query:"disposition_status"`
	TitleLike         string              `query:"title"`
	ArticleID         optionalUint64Param `query:"article_id"`
	Page              optionalIntParam    `query:"page"`
	PageSize          optionalIntParam    `query:"page_size"`
}

type resultDetailRequest struct {
	ID    uint64Param `path:"id"`
	OrgID uint64Param `query:"orgid"`
}

func registerResultRoutes(api huma.API, service *ResultService) {
	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-result-list",
		Method:             http.MethodGet,
		Path:               "/results",
		Summary:            "list article inspection results",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *resultListRequest) (*okEnvelopeOutput, error) {
		orgID, err := input.OrgID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid result query"), nil
		}
		taskID, err := input.TaskID.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid result query"), nil
		}
		articleID, err := input.ArticleID.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid result query"), nil
		}
		page, err := input.Page.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid result query"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid result query"), nil
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
		return successOKEnvelope(http.StatusOK, "result list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-result-detail",
		Method:             http.MethodGet,
		Path:               "/results/{id}",
		Summary:            "get article inspection result detail",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *resultDetailRequest) (*okEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid result input"), nil
		}
		orgID, err := input.OrgID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid result input"), nil
		}
		result, err := service.GetDetail(ctx, orgID, id)
		if err != nil {
			return failureOKFromError(err)
		}
		return successOKEnvelope(http.StatusOK, "result detail", result), nil
	})
}

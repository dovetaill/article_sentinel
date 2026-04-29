package articleinspect

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type articleLogRequest struct {
	ArticleID uint64Param      `path:"article_id"`
	OrgID     uint64Param      `query:"orgid"`
	Page      optionalIntParam `query:"page"`
	PageSize  optionalIntParam `query:"page_size"`
}

type operationLogListRequest struct {
	OrgID        uint64Param         `query:"orgid"`
	ArticleID    optionalUint64Param `query:"article_id"`
	TaskID       optionalUint64Param `query:"task_id"`
	OperatorName string              `query:"operator_name"`
	StartAt      optionalTimeParam   `query:"start_at"`
	EndAt        optionalTimeParam   `query:"end_at"`
	Page         optionalIntParam    `query:"page"`
	PageSize     optionalIntParam    `query:"page_size"`
}

type fieldChangeLogListRequest struct {
	OrgID     uint64Param         `query:"orgid"`
	ArticleID optionalUint64Param `query:"article_id"`
	FieldName string              `query:"field_name"`
	StartAt   optionalTimeParam   `query:"start_at"`
	EndAt     optionalTimeParam   `query:"end_at"`
	Page      optionalIntParam    `query:"page"`
	PageSize  optionalIntParam    `query:"page_size"`
}

func registerLogRoutes(api huma.API, service *LogService) {
	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-operation-log-list",
		Method:             http.MethodGet,
		Path:               "/logs/operations",
		Summary:            "list inspection operation logs",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *operationLogListRequest) (*okEnvelopeOutput, error) {
		orgID, err := input.OrgID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		articleID, err := input.ArticleID.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		taskID, err := input.TaskID.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		startAt, err := input.StartAt.Ptr()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid start_at"), nil
		}
		endAt, err := input.EndAt.Ptr()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid end_at"), nil
		}
		page, err := input.Page.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		result, err := service.ListOperationLogs(ctx, OperationLogListInput{
			OrgID:        orgID,
			ArticleID:    articleID,
			TaskID:       taskID,
			OperatorName: input.OperatorName,
			StartAt:      startAt,
			EndAt:        endAt,
			Page:         page,
			PageSize:     pageSize,
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return successOKEnvelope(http.StatusOK, "operation log list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-field-change-log-list",
		Method:             http.MethodGet,
		Path:               "/logs/field-changes",
		Summary:            "list inspection field change logs",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *fieldChangeLogListRequest) (*okEnvelopeOutput, error) {
		orgID, err := input.OrgID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		articleID, err := input.ArticleID.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		startAt, err := input.StartAt.Ptr()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid start_at"), nil
		}
		endAt, err := input.EndAt.Ptr()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid end_at"), nil
		}
		page, err := input.Page.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		result, err := service.ListFieldChangeLogs(ctx, FieldChangeLogListInput{
			OrgID:     orgID,
			ArticleID: articleID,
			FieldName: input.FieldName,
			StartAt:   startAt,
			EndAt:     endAt,
			Page:      page,
			PageSize:  pageSize,
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return successOKEnvelope(http.StatusOK, "field change log list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-article-operation-log-list",
		Method:             http.MethodGet,
		Path:               "/articles/{article_id}/operation-logs",
		Summary:            "list operation logs for an article",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *articleLogRequest) (*okEnvelopeOutput, error) {
		articleID, err := input.ArticleID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		orgID, err := input.OrgID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		page, err := input.Page.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		result, err := service.ListOperationLogs(ctx, OperationLogListInput{
			OrgID:     orgID,
			ArticleID: articleID,
			Page:      page,
			PageSize:  pageSize,
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return successOKEnvelope(http.StatusOK, "operation log list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-article-field-change-log-list",
		Method:             http.MethodGet,
		Path:               "/articles/{article_id}/change-logs",
		Summary:            "list field change logs for an article",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *articleLogRequest) (*okEnvelopeOutput, error) {
		articleID, err := input.ArticleID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		orgID, err := input.OrgID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		page, err := input.Page.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		result, err := service.ListFieldChangeLogs(ctx, FieldChangeLogListInput{
			OrgID:     orgID,
			ArticleID: articleID,
			Page:      page,
			PageSize:  pageSize,
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return successOKEnvelope(http.StatusOK, "field change log list", result), nil
	})
}

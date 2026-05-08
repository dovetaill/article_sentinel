package audit

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

type articleLogRequest struct {
	ArticleID sharedpkg.Uint64Param      `path:"article_id"`
	OrgID     sharedpkg.Uint64Param      `query:"orgid"`
	Page      sharedpkg.OptionalIntParam `query:"page"`
	PageSize  sharedpkg.OptionalIntParam `query:"page_size"`
}

type operationLogListRequest struct {
	OrgID        sharedpkg.Uint64Param         `query:"orgid"`
	ArticleID    sharedpkg.OptionalUint64Param `query:"article_id"`
	TaskID       sharedpkg.OptionalUint64Param `query:"task_id"`
	OperatorName string                        `query:"operator_name"`
	StartAt      sharedpkg.OptionalTimeParam   `query:"start_at"`
	EndAt        sharedpkg.OptionalTimeParam   `query:"end_at"`
	Page         sharedpkg.OptionalIntParam    `query:"page"`
	PageSize     sharedpkg.OptionalIntParam    `query:"page_size"`
}

type fieldChangeLogListRequest struct {
	OrgID     sharedpkg.Uint64Param         `query:"orgid"`
	ArticleID sharedpkg.OptionalUint64Param `query:"article_id"`
	FieldName string                        `query:"field_name"`
	StartAt   sharedpkg.OptionalTimeParam   `query:"start_at"`
	EndAt     sharedpkg.OptionalTimeParam   `query:"end_at"`
	Page      sharedpkg.OptionalIntParam    `query:"page"`
	PageSize  sharedpkg.OptionalIntParam    `query:"page_size"`
}

func RegisterLogRoutes(api huma.API, service *LogService) {
	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-operation-log-list",
		Method:             http.MethodGet,
		Path:               "/logs/operations",
		Summary:            "list inspection operation logs",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *operationLogListRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		articleID, err := input.ArticleID.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		taskID, err := input.TaskID.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		startAt, err := input.StartAt.Ptr()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid start_at"), nil
		}
		endAt, err := input.EndAt.Ptr()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid end_at"), nil
		}
		page, err := input.Page.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
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
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "operation log list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-field-change-log-list",
		Method:             http.MethodGet,
		Path:               "/logs/field-changes",
		Summary:            "list inspection field change logs",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *fieldChangeLogListRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		articleID, err := input.ArticleID.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		startAt, err := input.StartAt.Ptr()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid start_at"), nil
		}
		endAt, err := input.EndAt.Ptr()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid end_at"), nil
		}
		page, err := input.Page.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid log query"), nil
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
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "field change log list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-article-operation-log-list",
		Method:             http.MethodGet,
		Path:               "/articles/{article_id}/operation-logs",
		Summary:            "list operation logs for an article",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *articleLogRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		articleID, err := input.ArticleID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		page, err := input.Page.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
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
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "operation log list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-article-field-change-log-list",
		Method:             http.MethodGet,
		Path:               "/articles/{article_id}/change-logs",
		Summary:            "list field change logs for an article",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *articleLogRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		articleID, err := input.ArticleID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		page, err := input.Page.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
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
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "field change log list", result), nil
	})
}

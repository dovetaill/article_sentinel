package articleinspect

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type batchActionBody struct {
	OrgID     uint64   `json:"orgid,omitempty"`
	TaskID    uint64   `json:"task_id,omitempty"`
	ResultIDs []uint64 `json:"result_ids,omitempty"`
	Reason    string   `json:"reason,omitempty"`
}

type batchActionRequest struct {
	Body batchActionBody
}

func registerActionRoutes(api huma.API, service *ActionService) {
	registerBatchActionRoute(api, service, "/actions/batch-offline", "article-inspect-batch-offline", "batch offline results", service.BatchOffline)
	registerBatchActionRoute(api, service, "/actions/batch-ignore", "article-inspect-batch-ignore", "batch ignore results", service.BatchIgnore)
	registerBatchActionRoute(api, service, "/actions/batch-process", "article-inspect-batch-process", "batch process results", service.BatchProcess)
}

func registerBatchActionRoute(api huma.API, service *ActionService, path, operationID, summary string, handler func(context.Context, BatchActionInput) (*BatchActionSummary, error)) {
	huma.Register(api, huma.Operation{
		OperationID: operationID,
		Method:      http.MethodPost,
		Path:        path,
		Summary:     summary,
	}, func(ctx context.Context, input *batchActionRequest) (*okEnvelopeOutput, error) {
		if len(input.Body.ResultIDs) == 0 {
			return failureOKEnvelope(http.StatusBadRequest, "result_ids are required"), nil
		}
		operator := ResolveOperator(ctx)
		result, err := handler(ctx, BatchActionInput{
			OrgID:        input.Body.OrgID,
			TaskID:       input.Body.TaskID,
			ResultIDs:    input.Body.ResultIDs,
			Reason:       input.Body.Reason,
			OperatorID:   operator.ID,
			OperatorName: operator.Name,
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return successOKEnvelope(http.StatusOK, "batch action applied", result), nil
	})
}

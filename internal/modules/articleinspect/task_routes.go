package articleinspect

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type taskCreateRequest struct {
	Body CreateInspectionTaskInput
}

type taskQueryRequest struct {
	OrgID    uint64Param      `query:"orgid"`
	Page     optionalIntParam `query:"page"`
	PageSize optionalIntParam `query:"page_size"`
	Status   string           `query:"status"`
	TaskNo   string           `query:"task_no"`
}

type taskDetailRequest struct {
	ID    uint64Param `path:"id"`
	OrgID uint64Param `query:"orgid"`
}

func registerTaskRoutes(api huma.API, service *TaskService, dispatcher TaskDispatcher, logger *slog.Logger) {
	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-task-list",
		Method:             http.MethodGet,
		Path:               "/tasks",
		Summary:            "list article inspection tasks",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *taskQueryRequest) (*okEnvelopeOutput, error) {
		orgID, err := input.OrgID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		page, err := input.Page.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		result, err := service.List(ctx, TaskListInput{
			OrgID:    orgID,
			Page:     page,
			PageSize: pageSize,
			Status:   input.Status,
			TaskNo:   input.TaskNo,
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return successOKEnvelope(http.StatusOK, "task list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "article-inspect-task-create",
		Method:        http.MethodPost,
		Path:          "/tasks",
		Summary:       "create article inspection task",
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, input *taskCreateRequest) (*createdEnvelopeOutput, error) {
		task, err := service.CreateAndEnqueue(ctx, input.Body, NewTaskOutboxRelay(service.db, dispatcher, logger))
		if err != nil {
			return failureCreatedFromError(err)
		}
		return successCreatedEnvelope("task created", task), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-task-detail",
		Method:             http.MethodGet,
		Path:               "/tasks/{id}",
		Summary:            "get article inspection task detail",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *taskDetailRequest) (*okEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		orgID, err := input.OrgID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		result, err := service.Get(ctx, orgID, id)
		if err != nil {
			return failureOKFromError(err)
		}
		return successOKEnvelope(http.StatusOK, "task detail", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-task-delete",
		Method:             http.MethodDelete,
		Path:               "/tasks/{id}",
		Summary:            "delete article inspection task",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *taskDetailRequest) (*okEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		orgID, err := input.OrgID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		if err := service.Delete(ctx, orgID, id); err != nil {
			return failureOKFromError(err)
		}
		return successOKEnvelope(http.StatusOK, "task deleted", map[string]uint64{"id": id}), nil
	})
}

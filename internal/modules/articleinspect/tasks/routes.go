package tasks

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	outboxpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/outbox"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

type taskCreateRequest struct {
	Body CreateInspectionTaskInput
}

type taskQueryRequest struct {
	OrgID    sharedpkg.Uint64Param      `query:"orgid"`
	Page     sharedpkg.OptionalIntParam `query:"page"`
	PageSize sharedpkg.OptionalIntParam `query:"page_size"`
	Status   string                     `query:"status"`
	TaskNo   string                     `query:"task_no"`
}

type taskDetailRequest struct {
	ID    sharedpkg.Uint64Param `path:"id"`
	OrgID sharedpkg.Uint64Param `query:"orgid"`
}

func RegisterTaskRoutes(api huma.API, service *TaskService, dispatcher outboxpkg.TaskDispatcher, logger *slog.Logger, outboxSettings outboxpkg.TaskOutboxSettings) {
	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-task-list",
		Method:             http.MethodGet,
		Path:               "/tasks",
		Summary:            "list article inspection tasks",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *taskQueryRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		page, err := input.Page.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		result, err := service.List(ctx, TaskListInput{OrgID: orgID, Page: page, PageSize: pageSize, Status: input.Status, TaskNo: input.TaskNo})
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "task list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "article-inspect-task-create",
		Method:        http.MethodPost,
		Path:          "/tasks",
		Summary:       "create article inspection task",
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, input *taskCreateRequest) (*sharedpkg.CreatedEnvelopeOutput, error) {
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureCreatedFromError(err)
		}
		relay := outboxpkg.NewTaskOutboxRelay(service.db, dispatcher, logger)
		if relay != nil {
			relay = relay.WithSettings(outboxSettings)
		}
		createInput := input.Body
		createInput.OrgID = orgID
		task, err := service.CreateAndEnqueue(ctx, createInput, relay)
		if err != nil {
			return failureCreatedFromError(err)
		}
		return sharedpkg.SuccessCreatedEnvelope("task created", task), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-task-detail",
		Method:             http.MethodGet,
		Path:               "/tasks/{id}",
		Summary:            "get article inspection task detail",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *taskDetailRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		result, err := service.Get(ctx, orgID, id)
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "task detail", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-task-delete",
		Method:             http.MethodDelete,
		Path:               "/tasks/{id}",
		Summary:            "delete article inspection task",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *taskDetailRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		id, err := input.ID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		if err := service.Delete(ctx, orgID, id); err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "task deleted", map[string]uint64{"id": id}), nil
	})
}

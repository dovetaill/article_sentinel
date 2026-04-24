package articleinspect

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dovetaill/article-sentinel/internal/api/response"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
	"gorm.io/gorm"
)

type TaskDispatcher interface {
	DispatchArticleInspectTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) error
}

type Routes struct {
	Categories *CategoryService
	Keywords   *KeywordService
	Tasks      *TaskService
	Results    *ResultService
	Actions    *ActionService
	Lifecycle  *LifecycleService
	Logs       *LogService
	Articles   *ArticleService
	Dispatcher TaskDispatcher
}

type envelopeOutput struct {
	Status int `status:"200"`
	Body   response.Envelope
}

type keywordIDRequest struct {
	ID string `path:"id"`
}

type keywordQueryRequest struct {
	OrgID      uint64 `query:"orgid"`
	Page       int    `query:"page"`
	PageSize   int    `query:"page_size"`
	CategoryID uint64 `query:"category_id"`
	Query      string `query:"keyword"`
	Enabled    string `query:"enabled"`
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
	ID    string `path:"id"`
	OrgID uint64 `query:"orgid"`
}

type keywordUpdateRequest struct {
	ID   string `path:"id"`
	Body keywordBody
}

type keywordPatchStatusBody struct {
	OrgID   uint64 `json:"orgid"`
	Enabled bool   `json:"enabled"`
}

type keywordPatchStatusRequest struct {
	ID   string `path:"id"`
	Body keywordPatchStatusBody
}

type orgListRequest struct{}

type categoryQueryRequest struct {
	OrgID    uint64 `query:"orgid"`
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	Query    string `query:"name"`
	Enabled  string `query:"enabled"`
}

type categoryDetailRequest struct {
	ID    string `path:"id"`
	OrgID uint64 `query:"orgid"`
}

type categoryBody struct {
	OrgID   uint64 `json:"orgid,omitempty"`
	Name    string `json:"name,omitempty"`
	Code    string `json:"code,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
	Sort    int64  `json:"sort,omitempty"`
}

type categoryCreateRequest struct {
	Body categoryBody
}

type categoryUpdateRequest struct {
	ID   string `path:"id"`
	Body categoryBody
}

type categoryPatchStatusBody struct {
	OrgID   uint64 `json:"orgid"`
	Enabled bool   `json:"enabled"`
}

type categoryPatchStatusRequest struct {
	ID   string `path:"id"`
	Body categoryPatchStatusBody
}

type taskCreateRequest struct {
	Body CreateInspectionTaskInput
}

type taskQueryRequest struct {
	OrgID    uint64 `query:"orgid"`
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	Status   string `query:"status"`
	TaskNo   string `query:"task_no"`
}

type taskDetailRequest struct {
	ID    string `path:"id"`
	OrgID uint64 `query:"orgid"`
}

type resultListRequest struct {
	OrgID             uint64 `query:"orgid"`
	TaskID            uint64 `query:"task_id"`
	RiskLevel         string `query:"risk_level"`
	DispositionStatus string `query:"disposition_status"`
	TitleLike         string `query:"title"`
	ArticleID         uint64 `query:"article_id"`
	Page              int    `query:"page"`
	PageSize          int    `query:"page_size"`
}

type resultDetailRequest struct {
	ID    string `path:"id"`
	OrgID uint64 `query:"orgid"`
}

type batchActionBody struct {
	OrgID     uint64   `json:"orgid,omitempty"`
	TaskID    uint64   `json:"task_id,omitempty"`
	ResultIDs []uint64 `json:"result_ids,omitempty"`
	Reason    string   `json:"reason,omitempty"`
}

type batchActionRequest struct {
	Body batchActionBody
}

type articleIDRequest struct {
	ArticleID string `path:"article_id"`
}

type articleListRequest struct {
	OrgID    uint64 `query:"orgid"`
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
	State    string `query:"state"`
	Query    string `query:"keyword"`
}

type articleDetailRequest struct {
	ArticleID string `path:"article_id"`
	OrgID     uint64 `query:"orgid"`
}

type articleLifecycleBody struct {
	OrgID    uint64 `json:"orgid,omitempty"`
	TaskID   uint64 `json:"task_id,omitempty"`
	ResultID uint64 `json:"result_id,omitempty"`
	ActionID uint64 `json:"action_id,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type articleOfflineRequest struct {
	ArticleID string `path:"article_id"`
	Body      articleLifecycleBody
}

type articleRectifyBody struct {
	OrgID      uint64 `json:"orgid,omitempty"`
	TaskID     uint64 `json:"task_id,omitempty"`
	ResultID   uint64 `json:"result_id,omitempty"`
	ActionID   uint64 `json:"action_id,omitempty"`
	Reason     string `json:"reason,omitempty"`
	Title      string `json:"title,omitempty"`
	ShortTitle string `json:"short_title,omitempty"`
	RichTitle  string `json:"rich_title,omitempty"`
	Keyword    string `json:"keyword,omitempty"`
	Desc       string `json:"desc,omitempty"`
	Body       string `json:"body,omitempty"`
}

type articleRectifyRequest struct {
	ArticleID string `path:"article_id"`
	Body      articleRectifyBody
}

type articleRepublishRequest struct {
	ArticleID string `path:"article_id"`
	Body      articleLifecycleBody
}

type articleLogRequest struct {
	ArticleID string `path:"article_id"`
	OrgID     uint64 `query:"orgid"`
	Page      int    `query:"page"`
	PageSize  int    `query:"page_size"`
}

type operationLogListRequest struct {
	OrgID        uint64 `query:"orgid"`
	ArticleID    uint64 `query:"article_id"`
	TaskID       uint64 `query:"task_id"`
	OperatorName string `query:"operator_name"`
	StartAt      string `query:"start_at"`
	EndAt        string `query:"end_at"`
	Page         int    `query:"page"`
	PageSize     int    `query:"page_size"`
}

type fieldChangeLogListRequest struct {
	OrgID     uint64 `query:"orgid"`
	ArticleID uint64 `query:"article_id"`
	FieldName string `query:"field_name"`
	StartAt   string `query:"start_at"`
	EndAt     string `query:"end_at"`
	Page      int    `query:"page"`
	PageSize  int    `query:"page_size"`
}

func RegisterRoutes(api huma.API, routes Routes) {
	if api == nil {
		return
	}

	if routes.Categories != nil {
		registerCategoryRoutes(api, routes.Categories)
	}
	if routes.Keywords != nil {
		registerKeywordRoutes(api, routes.Keywords)
	}
	if routes.Tasks != nil {
		registerTaskRoutes(api, routes.Tasks, routes.Dispatcher)
	}
	if routes.Results != nil {
		registerResultRoutes(api, routes.Results)
	}
	if routes.Actions != nil {
		registerActionRoutes(api, routes.Actions)
	}
	if routes.Lifecycle != nil {
		registerLifecycleRoutes(api, routes.Lifecycle)
	}
	if routes.Logs != nil {
		registerLogRoutes(api, routes.Logs)
	}
	if routes.Articles != nil {
		registerArticleRoutes(api, routes.Articles)
	}
}

func registerCategoryRoutes(api huma.API, service *CategoryService) {
	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-org-list",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/orgs",
		Summary:     "list inspection organizations",
	}, func(ctx context.Context, input *orgListRequest) (*envelopeOutput, error) {
		_ = input
		result, err := service.ListOrgs(ctx)
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "org list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-category-list",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/categories",
		Summary:     "list inspection categories",
	}, func(ctx context.Context, input *categoryQueryRequest) (*envelopeOutput, error) {
		enabled, err := parseOptionalBool(input.Enabled)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid enabled filter"), nil
		}
		result, err := service.List(ctx, CategoryListInput{
			OrgID:    input.OrgID,
			Page:     input.Page,
			PageSize: input.PageSize,
			Enabled:  enabled,
			Query:    input.Query,
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "category list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-category-create",
		Method:      http.MethodPost,
		Path:        "/api/v1/article-inspect/categories",
		Summary:     "create inspection category",
	}, func(ctx context.Context, input *categoryCreateRequest) (*envelopeOutput, error) {
		item, err := service.Create(ctx, CreateCategoryInput{
			OrgID:   input.Body.OrgID,
			Name:    input.Body.Name,
			Code:    input.Body.Code,
			Enabled: input.Body.Enabled,
			Sort:    input.Body.Sort,
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusCreated, "category created", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-category-detail",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/categories/{id}",
		Summary:     "get inspection category",
	}, func(ctx context.Context, input *categoryDetailRequest) (*envelopeOutput, error) {
		id, err := parseUint64ID(input.ID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		item, err := service.Get(ctx, input.OrgID, id)
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "category detail", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-category-update",
		Method:      http.MethodPut,
		Path:        "/api/v1/article-inspect/categories/{id}",
		Summary:     "update inspection category",
	}, func(ctx context.Context, input *categoryUpdateRequest) (*envelopeOutput, error) {
		id, err := parseUint64ID(input.ID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		item, err := service.Update(ctx, UpdateCategoryInput{
			ID: id,
			CreateCategoryInput: CreateCategoryInput{
				OrgID:   input.Body.OrgID,
				Name:    input.Body.Name,
				Code:    input.Body.Code,
				Enabled: input.Body.Enabled,
				Sort:    input.Body.Sort,
			},
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "category updated", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-category-status-patch",
		Method:      http.MethodPatch,
		Path:        "/api/v1/article-inspect/categories/{id}/status",
		Summary:     "patch inspection category status",
	}, func(ctx context.Context, input *categoryPatchStatusRequest) (*envelopeOutput, error) {
		id, err := parseUint64ID(input.ID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		item, err := service.PatchEnabled(ctx, PatchCategoryStatusInput{
			OrgID:      input.Body.OrgID,
			CategoryID: id,
			Enabled:    input.Body.Enabled,
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "category status updated", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-category-delete",
		Method:      http.MethodDelete,
		Path:        "/api/v1/article-inspect/categories/{id}",
		Summary:     "delete inspection category",
	}, func(ctx context.Context, input *categoryDetailRequest) (*envelopeOutput, error) {
		id, err := parseUint64ID(input.ID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid category input"), nil
		}
		if err := service.Delete(ctx, input.OrgID, id); err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "category deleted", map[string]uint64{"id": id}), nil
	})
}

func registerKeywordRoutes(api huma.API, service *KeywordService) {
	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-keyword-create",
		Method:      http.MethodPost,
		Path:        "/api/v1/article-inspect/keywords",
		Summary:     "create inspection keyword",
	}, func(ctx context.Context, input *keywordCreateRequest) (*envelopeOutput, error) {
		item, err := service.Create(ctx, CreateKeywordInput{
			OrgID:         input.Body.OrgID,
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
			return failureFromError(err)
		}
		return successEnvelope(http.StatusCreated, "keyword created", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-keyword-list",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/keywords",
		Summary:     "list inspection keywords",
	}, func(ctx context.Context, input *keywordQueryRequest) (*envelopeOutput, error) {
		enabled, err := parseOptionalBool(input.Enabled)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid enabled filter"), nil
		}
		result, err := service.List(ctx, KeywordListInput{
			OrgID:      input.OrgID,
			Page:       input.Page,
			PageSize:   input.PageSize,
			Enabled:    enabled,
			CategoryID: input.CategoryID,
			Query:      input.Query,
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "keyword list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-keyword-detail",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/keywords/{id}",
		Summary:     "get inspection keyword",
	}, func(ctx context.Context, input *keywordDetailRequest) (*envelopeOutput, error) {
		id, err := parseUint64ID(input.ID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid keyword input"), nil
		}
		item, err := service.Get(ctx, input.OrgID, id)
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "keyword detail", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-keyword-update",
		Method:      http.MethodPut,
		Path:        "/api/v1/article-inspect/keywords/{id}",
		Summary:     "update inspection keyword",
	}, func(ctx context.Context, input *keywordUpdateRequest) (*envelopeOutput, error) {
		id, err := parseUint64ID(input.ID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid keyword input"), nil
		}
		item, err := service.Update(ctx, UpdateKeywordInput{
			ID: id,
			CreateKeywordInput: CreateKeywordInput{
				OrgID:         input.Body.OrgID,
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
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "keyword updated", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-keyword-status-patch",
		Method:      http.MethodPatch,
		Path:        "/api/v1/article-inspect/keywords/{id}/status",
		Summary:     "patch inspection keyword status",
	}, func(ctx context.Context, input *keywordPatchStatusRequest) (*envelopeOutput, error) {
		id, err := parseUint64ID(input.ID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid keyword input"), nil
		}
		item, err := service.PatchEnabled(ctx, PatchKeywordStatusInput{
			OrgID:     input.Body.OrgID,
			KeywordID: id,
			Enabled:   input.Body.Enabled,
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "keyword status updated", item), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-keyword-delete",
		Method:      http.MethodDelete,
		Path:        "/api/v1/article-inspect/keywords/{id}",
		Summary:     "delete inspection keyword",
	}, func(ctx context.Context, input *keywordDetailRequest) (*envelopeOutput, error) {
		id, err := parseUint64ID(input.ID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid keyword input"), nil
		}
		if err := service.Delete(ctx, input.OrgID, id); err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "keyword deleted", map[string]uint64{"id": id}), nil
	})
}

func registerTaskRoutes(api huma.API, service *TaskService, dispatcher TaskDispatcher) {
	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-task-list",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/tasks",
		Summary:     "list article inspection tasks",
	}, func(ctx context.Context, input *taskQueryRequest) (*envelopeOutput, error) {
		result, err := service.List(ctx, TaskListInput{
			OrgID:    input.OrgID,
			Page:     input.Page,
			PageSize: input.PageSize,
			Status:   input.Status,
			TaskNo:   input.TaskNo,
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "task list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-task-create",
		Method:      http.MethodPost,
		Path:        "/api/v1/article-inspect/tasks",
		Summary:     "create article inspection task",
	}, func(ctx context.Context, input *taskCreateRequest) (*envelopeOutput, error) {
		if dispatcher == nil {
			return failureEnvelope(http.StatusServiceUnavailable, "article inspect queue unavailable"), nil
		}

		task, err := service.Create(ctx, input.Body)
		if err != nil {
			return failureFromError(err)
		}

		operator := ResolveOperator(ctx)
		payload := queuetasks.ArticleInspectTaskPayload{
			TaskID:        task.ID,
			OrgID:         task.OrgID,
			TriggerSource: "api",
			OperatorID:    operator.ID,
			OperatorName:  operator.Name,
		}
		if err := dispatcher.DispatchArticleInspectTask(ctx, payload); err != nil {
			return failureEnvelope(http.StatusInternalServerError, "enqueue inspection task failed"), nil
		}
		return successEnvelope(http.StatusCreated, "task created", task), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-task-detail",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/tasks/{id}",
		Summary:     "get article inspection task detail",
	}, func(ctx context.Context, input *taskDetailRequest) (*envelopeOutput, error) {
		id, err := parseUint64ID(input.ID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid task input"), nil
		}
		result, err := service.Get(ctx, input.OrgID, id)
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "task detail", result), nil
	})
}

func registerResultRoutes(api huma.API, service *ResultService) {
	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-result-list",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/results",
		Summary:     "list article inspection results",
	}, func(ctx context.Context, input *resultListRequest) (*envelopeOutput, error) {
		result, err := service.List(ctx, ResultListInput{
			OrgID:             input.OrgID,
			TaskID:            input.TaskID,
			RiskLevel:         input.RiskLevel,
			DispositionStatus: input.DispositionStatus,
			TitleLike:         input.TitleLike,
			ArticleID:         input.ArticleID,
			Page:              input.Page,
			PageSize:          input.PageSize,
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "result list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-result-detail",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/results/{id}",
		Summary:     "get article inspection result detail",
	}, func(ctx context.Context, input *resultDetailRequest) (*envelopeOutput, error) {
		id, err := parseUint64ID(input.ID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid result input"), nil
		}
		result, err := service.GetDetail(ctx, input.OrgID, id)
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "result detail", result), nil
	})
}

func registerActionRoutes(api huma.API, service *ActionService) {
	registerBatchActionRoute(api, service, "/api/v1/article-inspect/actions/batch-ignore", "article-inspect-batch-ignore", "batch ignore results", service.BatchIgnore)
	registerBatchActionRoute(api, service, "/api/v1/article-inspect/actions/batch-process", "article-inspect-batch-process", "batch process results", service.BatchProcess)
}

func registerBatchActionRoute(api huma.API, service *ActionService, path, operationID, summary string, handler func(context.Context, BatchActionInput) (*BatchActionSummary, error)) {
	huma.Register(api, huma.Operation{
		OperationID: operationID,
		Method:      http.MethodPost,
		Path:        path,
		Summary:     summary,
	}, func(ctx context.Context, input *batchActionRequest) (*envelopeOutput, error) {
		if len(input.Body.ResultIDs) == 0 {
			return failureEnvelope(http.StatusBadRequest, "result_ids are required"), nil
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
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "batch action applied", result), nil
	})
}

func registerLifecycleRoutes(api huma.API, service *LifecycleService) {
	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-article-offline",
		Method:      http.MethodPost,
		Path:        "/api/v1/article-inspect/articles/{article_id}/offline",
		Summary:     "offline a matched article",
	}, func(ctx context.Context, input *articleOfflineRequest) (*envelopeOutput, error) {
		articleID, err := parseUint64ID(input.ArticleID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		if input.Body.OrgID == 0 {
			return failureEnvelope(http.StatusBadRequest, "orgid is required"), nil
		}
		operator := ResolveOperator(ctx)
		result, err := service.OfflineArticle(ctx, OfflineArticleInput{
			OrgID:        input.Body.OrgID,
			ArticleID:    articleID,
			TaskID:       input.Body.TaskID,
			ResultID:     input.Body.ResultID,
			ActionID:     input.Body.ActionID,
			Reason:       input.Body.Reason,
			OperatorID:   operator.ID,
			OperatorName: operator.Name,
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "article offlined", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-article-rectify",
		Method:      http.MethodPut,
		Path:        "/api/v1/article-inspect/articles/{article_id}/rectify",
		Summary:     "rectify matched article fields",
	}, func(ctx context.Context, input *articleRectifyRequest) (*envelopeOutput, error) {
		articleID, err := parseUint64ID(input.ArticleID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		if input.Body.OrgID == 0 {
			return failureEnvelope(http.StatusBadRequest, "orgid is required"), nil
		}
		operator := ResolveOperator(ctx)
		result, err := service.UpdateArticleFields(ctx, UpdateArticleFieldsInput{
			OrgID:        input.Body.OrgID,
			ArticleID:    articleID,
			TaskID:       input.Body.TaskID,
			ResultID:     input.Body.ResultID,
			ActionID:     input.Body.ActionID,
			Reason:       input.Body.Reason,
			OperatorID:   operator.ID,
			OperatorName: operator.Name,
			Fields: EditableArticleFields{
				Title:      input.Body.Title,
				ShortTitle: input.Body.ShortTitle,
				RichTitle:  input.Body.RichTitle,
				Keyword:    input.Body.Keyword,
				Desc:       input.Body.Desc,
				Body:       input.Body.Body,
			},
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "article rectified", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-article-republish",
		Method:      http.MethodPost,
		Path:        "/api/v1/article-inspect/articles/{article_id}/republish",
		Summary:     "republish a rectified article",
	}, func(ctx context.Context, input *articleRepublishRequest) (*envelopeOutput, error) {
		articleID, err := parseUint64ID(input.ArticleID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		if input.Body.OrgID == 0 {
			return failureEnvelope(http.StatusBadRequest, "orgid is required"), nil
		}
		operator := ResolveOperator(ctx)
		result, err := service.RepublishArticle(ctx, RepublishArticleInput{
			OrgID:        input.Body.OrgID,
			ArticleID:    articleID,
			TaskID:       input.Body.TaskID,
			ResultID:     input.Body.ResultID,
			ActionID:     input.Body.ActionID,
			Reason:       input.Body.Reason,
			OperatorID:   operator.ID,
			OperatorName: operator.Name,
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "article republished", result), nil
	})
}

func registerLogRoutes(api huma.API, service *LogService) {
	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-operation-log-list",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/logs/operations",
		Summary:     "list inspection operation logs",
	}, func(ctx context.Context, input *operationLogListRequest) (*envelopeOutput, error) {
		startAt, err := parseOptionalTime(input.StartAt)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid start_at"), nil
		}
		endAt, err := parseOptionalTime(input.EndAt)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid end_at"), nil
		}
		result, err := service.ListOperationLogs(ctx, OperationLogListInput{
			OrgID:        input.OrgID,
			ArticleID:    input.ArticleID,
			TaskID:       input.TaskID,
			OperatorName: input.OperatorName,
			StartAt:      startAt,
			EndAt:        endAt,
			Page:         input.Page,
			PageSize:     input.PageSize,
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "operation log list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-field-change-log-list",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/logs/field-changes",
		Summary:     "list inspection field change logs",
	}, func(ctx context.Context, input *fieldChangeLogListRequest) (*envelopeOutput, error) {
		startAt, err := parseOptionalTime(input.StartAt)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid start_at"), nil
		}
		endAt, err := parseOptionalTime(input.EndAt)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid end_at"), nil
		}
		result, err := service.ListFieldChangeLogs(ctx, FieldChangeLogListInput{
			OrgID:     input.OrgID,
			ArticleID: input.ArticleID,
			FieldName: input.FieldName,
			StartAt:   startAt,
			EndAt:     endAt,
			Page:      input.Page,
			PageSize:  input.PageSize,
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "field change log list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-article-operation-log-list",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/articles/{article_id}/operation-logs",
		Summary:     "list operation logs for an article",
	}, func(ctx context.Context, input *articleLogRequest) (*envelopeOutput, error) {
		articleID, err := parseUint64ID(input.ArticleID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		result, err := service.ListOperationLogs(ctx, OperationLogListInput{
			OrgID:     input.OrgID,
			ArticleID: articleID,
			Page:      input.Page,
			PageSize:  input.PageSize,
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "operation log list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-article-field-change-log-list",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/articles/{article_id}/change-logs",
		Summary:     "list field change logs for an article",
	}, func(ctx context.Context, input *articleLogRequest) (*envelopeOutput, error) {
		articleID, err := parseUint64ID(input.ArticleID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		result, err := service.ListFieldChangeLogs(ctx, FieldChangeLogListInput{
			OrgID:     input.OrgID,
			ArticleID: articleID,
			Page:      input.Page,
			PageSize:  input.PageSize,
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "field change log list", result), nil
	})
}

func registerArticleRoutes(api huma.API, service *ArticleService) {
	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-article-list",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/articles",
		Summary:     "list real articles for the article center",
	}, func(ctx context.Context, input *articleListRequest) (*envelopeOutput, error) {
		state, err := parseOptionalInt8(input.State)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid state"), nil
		}
		result, err := service.List(ctx, ArticleListInput{
			OrgID:    input.OrgID,
			Page:     input.Page,
			PageSize: input.PageSize,
			State:    state,
			Query:    input.Query,
		})
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "article list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "article-inspect-article-detail",
		Method:      http.MethodGet,
		Path:        "/api/v1/article-inspect/articles/{article_id}",
		Summary:     "get real article detail with inspect summary",
	}, func(ctx context.Context, input *articleDetailRequest) (*envelopeOutput, error) {
		articleID, err := parseUint64ID(input.ArticleID)
		if err != nil {
			return failureEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		result, err := service.Get(ctx, input.OrgID, articleID)
		if err != nil {
			return failureFromError(err)
		}
		return successEnvelope(http.StatusOK, "article detail", result), nil
	})
}

func successEnvelope(status int, message string, data any) *envelopeOutput {
	return &envelopeOutput{Status: status, Body: response.OK(message, data)}
}

func failureFromError(err error) (*envelopeOutput, error) {
	status, message := articleInspectStatusFromError(err)
	return failureEnvelope(status, message), nil
}

func failureEnvelope(status int, message string) *envelopeOutput {
	return &envelopeOutput{Status: status, Body: response.Fail(status, message)}
}

func articleInspectStatusFromError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, "ok"
	case errors.Is(err, ErrCategoryNotFound):
		return http.StatusNotFound, "resource not found"
	case errors.Is(err, ErrKeywordNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		return http.StatusNotFound, "resource not found"
	case errors.Is(err, ErrArticleNotFound):
		return http.StatusNotFound, "resource not found"
	case errors.Is(err, ErrTaskNotFound):
		return http.StatusNotFound, "resource not found"
	case errors.Is(err, ErrInvalidCategoryInput):
		return http.StatusBadRequest, "invalid category input"
	case errors.Is(err, ErrInvalidKeywordInput):
		return http.StatusBadRequest, "invalid keyword input"
	case errors.Is(err, ErrInvalidTaskInput):
		return http.StatusBadRequest, "invalid task input"
	case errors.Is(err, ErrInvalidResultQuery):
		return http.StatusBadRequest, "invalid result query"
	case errors.Is(err, ErrInvalidActionInput):
		return http.StatusBadRequest, "invalid action input"
	case errors.Is(err, ErrInvalidLogQuery):
		return http.StatusBadRequest, "invalid log query"
	case errors.Is(err, ErrInvalidArticleQuery):
		return http.StatusBadRequest, "invalid article query"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func parseUint64ID(raw string) (uint64, error) {
	return strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
}

func parseOptionalTime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalBool(value string) (*bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalInt8(value string) (*int8, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 8)
	if err != nil {
		return nil, err
	}
	result := int8(parsed)
	return &result, nil
}

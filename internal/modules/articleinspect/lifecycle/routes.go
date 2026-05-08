package lifecycle

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

type articleLifecycleBody struct {
	OrgID    uint64 `json:"orgid,omitempty"`
	TaskID   uint64 `json:"task_id,omitempty"`
	ResultID uint64 `json:"result_id,omitempty"`
	ActionID uint64 `json:"action_id,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type articleOfflineRequest struct {
	ArticleID sharedpkg.Uint64Param `path:"article_id"`
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
	ArticleID sharedpkg.Uint64Param `path:"article_id"`
	Body      articleRectifyBody
}

type articleRepublishRequest struct {
	ArticleID sharedpkg.Uint64Param `path:"article_id"`
	Body      articleLifecycleBody
}

func RegisterLifecycleRoutes(api huma.API, service *LifecycleService) {
	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-article-offline",
		Method:             http.MethodPost,
		Path:               "/articles/{article_id}/offline",
		Summary:            "offline a matched article",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *articleOfflineRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		articleID, err := input.ArticleID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		operator := sharedpkg.ResolveOperator(ctx)
		result, err := service.OfflineArticle(ctx, OfflineArticleInput{
			OrgID:        orgID,
			ArticleID:    articleID,
			TaskID:       input.Body.TaskID,
			ResultID:     input.Body.ResultID,
			ActionID:     input.Body.ActionID,
			Reason:       input.Body.Reason,
			OperatorID:   operator.ID,
			OperatorName: operator.Name,
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "article offlined", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-article-rectify",
		Method:             http.MethodPut,
		Path:               "/articles/{article_id}/rectify",
		Summary:            "rectify matched article fields",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *articleRectifyRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		articleID, err := input.ArticleID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		operator := sharedpkg.ResolveOperator(ctx)
		result, err := service.UpdateArticleFields(ctx, UpdateArticleFieldsInput{
			OrgID:        orgID,
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
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "article rectified", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-article-republish",
		Method:             http.MethodPost,
		Path:               "/articles/{article_id}/republish",
		Summary:            "republish a rectified article",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *articleRepublishRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		articleID, err := input.ArticleID.Parse()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		operator := sharedpkg.ResolveOperator(ctx)
		result, err := service.RepublishArticle(ctx, RepublishArticleInput{
			OrgID:        orgID,
			ArticleID:    articleID,
			TaskID:       input.Body.TaskID,
			ResultID:     input.Body.ResultID,
			ActionID:     input.Body.ActionID,
			Reason:       input.Body.Reason,
			OperatorID:   operator.ID,
			OperatorName: operator.Name,
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "article republished", result), nil
	})
}

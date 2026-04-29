package articleinspect

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type articleListRequest struct {
	OrgID     uint64Param         `query:"orgid"`
	Page      optionalIntParam    `query:"page"`
	PageSize  optionalIntParam    `query:"page_size"`
	State     optionalInt8Param   `query:"state"`
	ArticleID optionalUint64Param `query:"article_id"`
	Title     string              `query:"title"`
}

type articleDetailRequest struct {
	ArticleID uint64Param `path:"article_id"`
	OrgID     uint64Param `query:"orgid"`
}

func registerArticleRoutes(api huma.API, service *ArticleService) {
	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-article-list",
		Method:             http.MethodGet,
		Path:               "/articles",
		Summary:            "list real articles for the article center",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *articleListRequest) (*okEnvelopeOutput, error) {
		orgID, err := input.OrgID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid article query"), nil
		}
		page, err := input.Page.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid article query"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid article query"), nil
		}
		state, err := input.State.Ptr()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid state"), nil
		}
		articleID, err := input.ArticleID.Value()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid article query"), nil
		}
		result, err := service.List(ctx, ArticleListInput{
			OrgID:     orgID,
			Page:      page,
			PageSize:  pageSize,
			State:     state,
			ArticleID: articleID,
			TitleLike: input.Title,
		})
		if err != nil {
			return failureOKFromError(err)
		}
		return successOKEnvelope(http.StatusOK, "article list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-article-detail",
		Method:             http.MethodGet,
		Path:               "/articles/{article_id}",
		Summary:            "get real article detail with inspect summary",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *articleDetailRequest) (*okEnvelopeOutput, error) {
		articleID, err := input.ArticleID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		orgID, err := input.OrgID.Parse()
		if err != nil {
			return failureOKEnvelope(http.StatusBadRequest, "invalid article input"), nil
		}
		result, err := service.Get(ctx, orgID, articleID)
		if err != nil {
			return failureOKFromError(err)
		}
		return successOKEnvelope(http.StatusOK, "article detail", result), nil
	})
}

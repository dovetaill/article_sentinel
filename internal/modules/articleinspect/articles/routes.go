package articles

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

type articleListRequest struct {
	OrgID     sharedpkg.Uint64Param         `query:"orgid"`
	Page      sharedpkg.OptionalIntParam    `query:"page"`
	PageSize  sharedpkg.OptionalIntParam    `query:"page_size"`
	State     sharedpkg.OptionalInt8Param   `query:"state"`
	ArticleID sharedpkg.OptionalUint64Param `query:"article_id"`
	Title     string                        `query:"title"`
}

type articleDetailRequest struct {
	ArticleID sharedpkg.Uint64Param `path:"article_id"`
	OrgID     sharedpkg.Uint64Param `query:"orgid"`
}

func RegisterArticleRoutes(api huma.API, service *ArticleService) {
	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-article-list",
		Method:             http.MethodGet,
		Path:               "/articles",
		Summary:            "list real articles for the article center",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *articleListRequest) (*sharedpkg.OKEnvelopeOutput, error) {
		if err := sharedpkg.ValidateOptionalOrgID(input.OrgID); err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article query"), nil
		}
		orgID, err := sharedpkg.CurrentOrgID(ctx)
		if err != nil {
			return failureOKFromError(err)
		}
		page, err := input.Page.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article query"), nil
		}
		pageSize, err := input.PageSize.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article query"), nil
		}
		state, err := input.State.Ptr()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid state"), nil
		}
		articleID, err := input.ArticleID.Value()
		if err != nil {
			return sharedpkg.FailureOKEnvelope(http.StatusBadRequest, "invalid article query"), nil
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
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "article list", result), nil
	})

	huma.Register(api, huma.Operation{
		OperationID:        "article-inspect-article-detail",
		Method:             http.MethodGet,
		Path:               "/articles/{article_id}",
		Summary:            "get real article detail with inspect summary",
		SkipValidateParams: true,
	}, func(ctx context.Context, input *articleDetailRequest) (*sharedpkg.OKEnvelopeOutput, error) {
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
		result, err := service.Get(ctx, orgID, articleID)
		if err != nil {
			return failureOKFromError(err)
		}
		return sharedpkg.SuccessOKEnvelope(http.StatusOK, "article detail", result), nil
	})
}

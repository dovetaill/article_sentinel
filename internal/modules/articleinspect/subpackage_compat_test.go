package articleinspect

import (
	"net/http"
	"testing"

	articlespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/articles"
	auditpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/audit"
	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	resultspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/results"
	rulespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/rules"
	scanpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/scan"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
)

func TestRootAliasesMatchExtractedSubpackages(t *testing.T) {
	var rootTask InspectionTask
	var domainTask domainpkg.InspectionTask
	rootTask = domainTask
	domainTask = rootTask

	var rootParam uint64Param
	var sharedParam sharedpkg.Uint64Param
	rootParam = sharedParam
	sharedParam = rootParam

	var rootCandidate CandidateArticle
	var scanCandidate scanpkg.CandidateArticle
	rootCandidate = scanCandidate
	scanCandidate = rootCandidate

	var rootRule KeywordRule
	var scanRule scanpkg.KeywordRule
	rootRule = scanRule
	scanRule = rootRule

	var rootHit Hit
	var scanHit scanpkg.Hit
	rootHit = scanHit
	scanHit = rootHit

	var rootKeywordDTO KeywordDTO
	var rulesKeywordDTO rulespkg.KeywordDTO
	rootKeywordDTO = rulesKeywordDTO
	rulesKeywordDTO = rootKeywordDTO

	var rootCategoryDTO CategoryDTO
	var rulesCategoryDTO rulespkg.CategoryDTO
	rootCategoryDTO = rulesCategoryDTO
	rulesCategoryDTO = rootCategoryDTO

	var rootKeywordRepo *KeywordRepository
	var rulesKeywordRepo *rulespkg.KeywordRepository
	rootKeywordRepo = rulesKeywordRepo
	rulesKeywordRepo = rootKeywordRepo

	var rootKeywordService *KeywordService
	var rulesKeywordService *rulespkg.KeywordService
	rootKeywordService = rulesKeywordService
	rulesKeywordService = rootKeywordService

	var rootArticleDetail ArticleDetail
	var articlesArticleDetail articlespkg.ArticleDetail
	rootArticleDetail = articlesArticleDetail
	articlesArticleDetail = rootArticleDetail

	var rootArticleService *ArticleService
	var articlesArticleService *articlespkg.ArticleService
	rootArticleService = articlesArticleService
	articlesArticleService = rootArticleService

	var rootResultListInput ResultListInput
	var resultsResultListInput resultspkg.ResultListInput
	rootResultListInput = resultsResultListInput
	resultsResultListInput = rootResultListInput

	var rootResultService *ResultService
	var resultsResultService *resultspkg.ResultService
	rootResultService = resultsResultService
	resultsResultService = rootResultService

	var rootLogInput OperationLogListInput
	var auditLogInput auditpkg.OperationLogListInput
	rootLogInput = auditLogInput
	auditLogInput = rootLogInput

	var rootLogService *LogService
	var auditLogService *auditpkg.LogService
	rootLogService = auditLogService
	auditLogService = rootLogService

	var rootScannerImpl KeywordScanner
	var scanScannerImpl scanpkg.KeywordScanner
	rootScannerImpl = scanScannerImpl
	scanScannerImpl = rootScannerImpl

	if ArticleStateOnline != domainpkg.ArticleStateOnline {
		t.Fatalf("ArticleStateOnline = %d, want %d", ArticleStateOnline, domainpkg.ArticleStateOnline)
	}

	rootScanner := NewKeywordScanner()
	if rootScanner == nil {
		t.Fatal("NewKeywordScanner() = nil")
	}
	if scanpkg.NewKeywordScanner() == nil {
		t.Fatal("scan.NewKeywordScanner() = nil")
	}
	var rootScannerContract Scanner = scanpkg.NewKeywordScanner()
	var scanScannerContract scanpkg.Scanner = NewKeywordScanner()
	if rootScannerContract == nil || scanScannerContract == nil {
		t.Fatal("scanner alias contracts = nil")
	}

	envelope := sharedpkg.SuccessOKEnvelope(http.StatusOK, "ok", map[string]string{"status": "ok"})
	if envelope.Status != http.StatusOK {
		t.Fatalf("shared SuccessOKEnvelope status = %d, want %d", envelope.Status, http.StatusOK)
	}
}

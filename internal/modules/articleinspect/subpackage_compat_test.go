package articleinspect

import (
	"context"
	"net/http"
	"testing"

	actionspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/actions"
	articlespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/articles"
	auditpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/audit"
	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	lifecyclepkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/lifecycle"
	outboxpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/outbox"
	resultspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/results"
	rulespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/rules"
	scanpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/scan"
	sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"
	taskspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/tasks"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
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

	var rootTaskFilter CandidateArticleFilter
	var tasksTaskFilter taskspkg.CandidateArticleFilter
	rootTaskFilter = tasksTaskFilter
	tasksTaskFilter = rootTaskFilter

	var rootTaskListInput TaskListInput
	var tasksTaskListInput taskspkg.TaskListInput
	rootTaskListInput = tasksTaskListInput
	tasksTaskListInput = rootTaskListInput

	var rootTaskCreateInput CreateInspectionTaskInput
	var tasksTaskCreateInput taskspkg.CreateInspectionTaskInput
	rootTaskCreateInput = tasksTaskCreateInput
	tasksTaskCreateInput = rootTaskCreateInput

	var rootTaskService *TaskService
	var tasksTaskService *taskspkg.TaskService
	rootTaskService = tasksTaskService
	tasksTaskService = rootTaskService

	var rootOutboxSettings TaskOutboxSettings
	var outboxSettings outboxpkg.TaskOutboxSettings
	rootOutboxSettings = outboxSettings
	outboxSettings = rootOutboxSettings

	var rootOutboxReport TaskOutboxDispatchReport
	var outboxReport outboxpkg.TaskOutboxDispatchReport
	rootOutboxReport = outboxReport
	outboxReport = rootOutboxReport

	var rootActionInput BatchActionInput
	var actionsInput actionspkg.BatchActionInput
	rootActionInput = actionsInput
	actionsInput = rootActionInput

	var rootActionSummary BatchActionSummary
	var actionsSummary actionspkg.BatchActionSummary
	rootActionSummary = actionsSummary
	actionsSummary = rootActionSummary

	var rootActionRepo *ActionRepository
	var actionsRepo *actionspkg.ActionRepository
	rootActionRepo = actionsRepo
	actionsRepo = rootActionRepo

	var rootActionService *ActionService
	var actionsService *actionspkg.ActionService
	rootActionService = actionsService
	actionsService = rootActionService

	var rootEditableFields EditableArticleFields
	var lifecycleFields lifecyclepkg.EditableArticleFields
	rootEditableFields = lifecycleFields
	lifecycleFields = rootEditableFields

	var rootFieldChange FieldChange
	var lifecycleFieldChange lifecyclepkg.FieldChange
	rootFieldChange = lifecycleFieldChange
	lifecycleFieldChange = rootFieldChange

	var rootOfflineInput OfflineArticleInput
	var lifecycleOfflineInput lifecyclepkg.OfflineArticleInput
	rootOfflineInput = lifecycleOfflineInput
	lifecycleOfflineInput = rootOfflineInput

	var rootUpdateFieldsInput UpdateArticleFieldsInput
	var lifecycleUpdateFieldsInput lifecyclepkg.UpdateArticleFieldsInput
	rootUpdateFieldsInput = lifecycleUpdateFieldsInput
	lifecycleUpdateFieldsInput = rootUpdateFieldsInput

	var rootRepublishInput RepublishArticleInput
	var lifecycleRepublishInput lifecyclepkg.RepublishArticleInput
	rootRepublishInput = lifecycleRepublishInput
	lifecycleRepublishInput = rootRepublishInput

	var rootLifecycleResult LifecycleActionResult
	var lifecycleResult lifecyclepkg.LifecycleActionResult
	rootLifecycleResult = lifecycleResult
	lifecycleResult = rootLifecycleResult

	var rootLifecycleService *LifecycleService
	var lifecycleService *lifecyclepkg.LifecycleService
	rootLifecycleService = lifecycleService
	lifecycleService = rootLifecycleService

	var rootRelay *TaskOutboxRelay
	var outboxRelay *outboxpkg.TaskOutboxRelay
	rootRelay = outboxRelay
	outboxRelay = rootRelay

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
	if ErrInvalidTaskInput != taskspkg.ErrInvalidTaskInput {
		t.Fatal("ErrInvalidTaskInput does not match tasks.ErrInvalidTaskInput")
	}
	if ErrTaskOutboxDispatcherUnavailable != outboxpkg.ErrTaskOutboxDispatcherUnavailable {
		t.Fatal("ErrTaskOutboxDispatcherUnavailable does not match outbox.ErrTaskOutboxDispatcherUnavailable")
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

	dispatcher := compatTaskDispatcher{}
	var rootDispatcher TaskDispatcher = dispatcher
	var tasksDispatcher taskspkg.TaskDispatcher = rootDispatcher
	var outboxDispatcher outboxpkg.TaskDispatcher = rootDispatcher
	rootDispatcher = tasksDispatcher
	rootDispatcher = outboxDispatcher
	if rootDispatcher == nil || tasksDispatcher == nil || outboxDispatcher == nil {
		t.Fatal("task dispatcher compatibility contracts = nil")
	}

	envelope := sharedpkg.SuccessOKEnvelope(http.StatusOK, "ok", map[string]string{"status": "ok"})
	if envelope.Status != http.StatusOK {
		t.Fatalf("shared SuccessOKEnvelope status = %d, want %d", envelope.Status, http.StatusOK)
	}
}

type compatTaskDispatcher struct{}

func (compatTaskDispatcher) DispatchArticleInspectTask(_ context.Context, _ queuetasks.ArticleInspectTaskPayload) error {
	return nil
}

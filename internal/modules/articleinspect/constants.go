package articleinspect

import domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"

const (
	ArticleStateDeleted      = domainpkg.ArticleStateDeleted
	ArticleStateAuditPending = domainpkg.ArticleStateAuditPending
	ArticleStateAuditBack    = domainpkg.ArticleStateAuditBack
	ArticleStateDraft        = domainpkg.ArticleStateDraft
	ArticleStateStep         = domainpkg.ArticleStateStep
	ArticleStateOfflineSync  = domainpkg.ArticleStateOfflineSync
	ArticleStateOffline      = domainpkg.ArticleStateOffline
	ArticleStateOnline       = domainpkg.ArticleStateOnline
)

var ArticleLifecycleStates = domainpkg.ArticleLifecycleStates

const (
	MatchTypeContains = domainpkg.MatchTypeContains
	MatchTypeExact    = domainpkg.MatchTypeExact
	MatchTypeRegex    = domainpkg.MatchTypeRegex
)

const (
	RiskLevelLow    = domainpkg.RiskLevelLow
	RiskLevelMedium = domainpkg.RiskLevelMedium
	RiskLevelHigh   = domainpkg.RiskLevelHigh
)

const (
	SuggestActionIgnore  = domainpkg.SuggestActionIgnore
	SuggestActionProcess = domainpkg.SuggestActionProcess
	SuggestActionOffline = domainpkg.SuggestActionOffline
)

const (
	KeywordScopeTitle      = domainpkg.KeywordScopeTitle
	KeywordScopeShortTitle = domainpkg.KeywordScopeShortTitle
	KeywordScopeRichTitle  = domainpkg.KeywordScopeRichTitle
	KeywordScopeKeyword    = domainpkg.KeywordScopeKeyword
	KeywordScopeDesc       = domainpkg.KeywordScopeDesc
	KeywordScopeBody       = domainpkg.KeywordScopeBody
)

const (
	TaskStatusPending        = domainpkg.TaskStatusPending
	TaskStatusRunning        = domainpkg.TaskStatusRunning
	TaskStatusSuccess        = domainpkg.TaskStatusSuccess
	TaskStatusFailed         = domainpkg.TaskStatusFailed
	TaskStatusPartialSuccess = domainpkg.TaskStatusPartialSuccess
)

const (
	TaskOutboxStatusPending    = domainpkg.TaskOutboxStatusPending
	TaskOutboxStatusClaimed    = domainpkg.TaskOutboxStatusClaimed
	TaskOutboxStatusDispatched = domainpkg.TaskOutboxStatusDispatched
	TaskOutboxStatusDeadLetter = domainpkg.TaskOutboxStatusDeadLetter
)

const (
	TaskOutboxMessageTypeRunTask = domainpkg.TaskOutboxMessageTypeRunTask
)

const (
	TaskOutboxErrorDispatch              = domainpkg.TaskOutboxErrorDispatch
	TaskOutboxErrorDispatcherUnavailable = domainpkg.TaskOutboxErrorDispatcherUnavailable
	TaskOutboxErrorPayloadDecode         = domainpkg.TaskOutboxErrorPayloadDecode
	TaskOutboxErrorUnsupportedMessage    = domainpkg.TaskOutboxErrorUnsupportedMessage
	TaskOutboxErrorDBUpdate              = domainpkg.TaskOutboxErrorDBUpdate
)

const (
	ResultDispositionPending     = domainpkg.ResultDispositionPending
	ResultDispositionIgnored     = domainpkg.ResultDispositionIgnored
	ResultDispositionProcessed   = domainpkg.ResultDispositionProcessed
	ResultDispositionOfflined    = domainpkg.ResultDispositionOfflined
	ResultDispositionRepublished = domainpkg.ResultDispositionRepublished
	ResultDispositionFailed      = domainpkg.ResultDispositionFailed
)

func InspectionTaskStatuses() []string {
	return domainpkg.InspectionTaskStatuses()
}

func InspectionResultDispositionStatuses() []string {
	return domainpkg.InspectionResultDispositionStatuses()
}

const (
	ActionStatusRunning = domainpkg.ActionStatusRunning
	ActionStatusSuccess = domainpkg.ActionStatusSuccess
	ActionStatusSkipped = domainpkg.ActionStatusSkipped
	ActionStatusFailed  = domainpkg.ActionStatusFailed
)

const (
	ActionTypeBatchIgnore  = domainpkg.ActionTypeBatchIgnore
	ActionTypeBatchProcess = domainpkg.ActionTypeBatchProcess
	ActionTypeOffline      = domainpkg.ActionTypeOffline
	ActionTypeRectify      = domainpkg.ActionTypeRectify
	ActionTypeRepublish    = domainpkg.ActionTypeRepublish
)

func InspectionKeywordScopes() []string {
	return domainpkg.InspectionKeywordScopes()
}

func InspectionMatchTypes() []string {
	return domainpkg.InspectionMatchTypes()
}

func InspectionRiskLevels() []string {
	return domainpkg.InspectionRiskLevels()
}

func InspectionSuggestActions() []string {
	return domainpkg.InspectionSuggestActions()
}

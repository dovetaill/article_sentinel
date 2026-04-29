package articleinspect

const (
	ArticleStateDeleted      int8 = 0
	ArticleStateAuditPending int8 = 1
	ArticleStateAuditBack    int8 = 2
	ArticleStateDraft        int8 = 3
	ArticleStateStep         int8 = 5
	ArticleStateOfflineSync  int8 = 7
	ArticleStateOffline      int8 = 8
	ArticleStateOnline       int8 = 9
)

var ArticleLifecycleStates = map[int8]string{
	ArticleStateDeleted:      "del",
	ArticleStateAuditPending: "audit",
	ArticleStateAuditBack:    "back",
	ArticleStateDraft:        "draft",
	ArticleStateStep:         "step",
	ArticleStateOfflineSync:  "offline_sync",
	ArticleStateOffline:      "offline",
	ArticleStateOnline:       "online",
}

const (
	MatchTypeContains = "contains"
	MatchTypeExact    = "exact"
	MatchTypeRegex    = "regex"
)

const (
	RiskLevelLow    = "low"
	RiskLevelMedium = "medium"
	RiskLevelHigh   = "high"
)

const (
	SuggestActionIgnore  = "ignore"
	SuggestActionProcess = "process"
	SuggestActionOffline = "offline"
)

const (
	KeywordScopeTitle      = "title"
	KeywordScopeShortTitle = "short_title"
	KeywordScopeRichTitle  = "rich_title"
	KeywordScopeKeyword    = "keyword"
	KeywordScopeDesc       = "desc"
	KeywordScopeBody       = "body"
)

const (
	TaskStatusPending        = "pending"
	TaskStatusRunning        = "running"
	TaskStatusSuccess        = "success"
	TaskStatusFailed         = "failed"
	TaskStatusPartialSuccess = "partial_success"
)

const (
	TaskOutboxStatusPending    = "pending"
	TaskOutboxStatusClaimed    = "claimed"
	TaskOutboxStatusDispatched = "dispatched"
	TaskOutboxStatusDeadLetter = "dead_letter"
)

const (
	TaskOutboxMessageTypeRunTask = "articleinspect.task.run"
)

const (
	TaskOutboxErrorDispatch              = "dispatch_error"
	TaskOutboxErrorDispatcherUnavailable = "dispatcher_unavailable"
	TaskOutboxErrorPayloadDecode         = "payload_decode_error"
	TaskOutboxErrorUnsupportedMessage    = "unsupported_message_type"
	TaskOutboxErrorDBUpdate              = "db_update_error"
)

const (
	ResultDispositionPending     = "pending"
	ResultDispositionIgnored     = "ignored"
	ResultDispositionProcessed   = "processed"
	ResultDispositionOfflined    = "offlined"
	ResultDispositionRepublished = "republished"
	ResultDispositionFailed      = "failed"
)

func InspectionTaskStatuses() []string {
	return []string{
		TaskStatusPending,
		TaskStatusRunning,
		TaskStatusSuccess,
		TaskStatusFailed,
		TaskStatusPartialSuccess,
	}
}

func InspectionResultDispositionStatuses() []string {
	return []string{
		ResultDispositionPending,
		ResultDispositionIgnored,
		ResultDispositionProcessed,
		ResultDispositionOfflined,
		ResultDispositionRepublished,
		ResultDispositionFailed,
	}
}

const (
	ActionStatusRunning = "running"
	ActionStatusSuccess = "success"
	ActionStatusSkipped = "skipped"
	ActionStatusFailed  = "failed"
)

const (
	ActionTypeBatchIgnore  = "batch_ignore"
	ActionTypeBatchProcess = "batch_process"
	ActionTypeOffline      = "offline"
	ActionTypeRectify      = "rectify"
	ActionTypeRepublish    = "republish"
)

func InspectionKeywordScopes() []string {
	return []string{
		KeywordScopeTitle,
		KeywordScopeShortTitle,
		KeywordScopeRichTitle,
		KeywordScopeKeyword,
		KeywordScopeDesc,
		KeywordScopeBody,
	}
}

func InspectionMatchTypes() []string {
	return []string{
		MatchTypeContains,
		MatchTypeExact,
		MatchTypeRegex,
	}
}

func InspectionRiskLevels() []string {
	return []string{
		RiskLevelLow,
		RiskLevelMedium,
		RiskLevelHigh,
	}
}

func InspectionSuggestActions() []string {
	return []string{
		SuggestActionIgnore,
		SuggestActionProcess,
		SuggestActionOffline,
	}
}

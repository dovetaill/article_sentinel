package articleinspect

import "time"

type InspectionTimestamps struct {
	CreateAt time.Time `gorm:"column:create_at;not null;autoCreateTime" json:"created_at"`
	UpdateAt time.Time `gorm:"column:update_at;not null;autoUpdateTime" json:"updated_at"`
}

type InspectionCategory struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrgID       uint64 `gorm:"column:orgid;not null;index" json:"orgid"`
	Name        string `gorm:"column:name;size:128;not null" json:"name"`
	Enabled     bool   `gorm:"column:enabled;not null;default:true" json:"enabled"`
	Sort        int64  `gorm:"column:sort;not null;default:0" json:"sort"`
	CreatorID   uint64 `gorm:"column:creator_id;not null;default:0" json:"creator_id"`
	CreatorName string `gorm:"column:creator_name;size:128;not null;default:''" json:"creator_name"`
	UpdaterID   uint64 `gorm:"column:updater_id;not null;default:0" json:"updater_id"`
	UpdaterName string `gorm:"column:updater_name;size:128;not null;default:''" json:"updater_name"`
	InspectionTimestamps
}

func (InspectionCategory) TableName() string {
	return "xt_article_inspect_categories"
}

type InspectionKeyword struct {
	ID            uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrgID         uint64 `gorm:"column:orgid;not null;index" json:"orgid"`
	Name          string `gorm:"column:name;size:255;not null" json:"name"`
	CategoryID    uint64 `gorm:"column:category_id;not null;index" json:"category_id"`
	MatchType     string `gorm:"column:match_type;size:32;not null" json:"match_type"`
	RiskLevel     string `gorm:"column:risk_level;size:32;not null" json:"risk_level"`
	SuggestAction string `gorm:"column:suggest_action;size:32;not null" json:"suggest_action"`
	Enabled       bool   `gorm:"column:enabled;not null;default:true" json:"enabled"`
	Remark        string `gorm:"column:remark;type:text" json:"remark"`
	CreatorID     uint64 `gorm:"column:creator_id;not null;default:0" json:"creator_id"`
	CreatorName   string `gorm:"column:creator_name;size:128;not null;default:''" json:"creator_name"`
	UpdaterID     uint64 `gorm:"column:updater_id;not null;default:0" json:"updater_id"`
	UpdaterName   string `gorm:"column:updater_name;size:128;not null;default:''" json:"updater_name"`
	InspectionTimestamps
}

func (InspectionKeyword) TableName() string {
	return "xt_article_inspect_keywords"
}

type InspectionKeywordScope struct {
	ID        uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrgID     uint64 `gorm:"column:orgid;not null;index" json:"orgid"`
	KeywordID uint64 `gorm:"column:keyword_id;not null;index" json:"keyword_id"`
	Scope     string `gorm:"column:scope;size:32;not null" json:"scope"`
	InspectionTimestamps
}

func (InspectionKeywordScope) TableName() string {
	return "xt_article_inspect_keyword_scopes"
}

type InspectionTask struct {
	ID                 uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrgID              uint64     `gorm:"column:orgid;not null;index" json:"orgid"`
	TaskNo             string     `gorm:"column:task_no;size:64;not null;uniqueIndex" json:"task_no"`
	Status             string     `gorm:"column:status;size:32;not null" json:"status"`
	ArticleStateFilter string     `gorm:"column:article_state_filter;size:64;not null;default:''" json:"article_state_filter"`
	PublishTimeStart   *time.Time `gorm:"column:publish_time_start" json:"publish_time_start"`
	PublishTimeEnd     *time.Time `gorm:"column:publish_time_end" json:"publish_time_end"`
	ArticleID          uint64     `gorm:"column:article_id;not null;default:0" json:"article_id"`
	TitleLike          string     `gorm:"column:title_like;size:255;not null;default:''" json:"title_like"`
	IncludeBody        bool       `gorm:"column:include_body;not null;default:true" json:"include_body"`
	ScopeOverrideMode  string     `gorm:"column:scope_override_mode;size:32;not null;default:''" json:"scope_override_mode"`
	ScopeSnapshot      string     `gorm:"column:scope_snapshot;type:longtext" json:"scope_snapshot"`
	RequestSnapshot    string     `gorm:"column:request_snapshot;type:longtext" json:"request_snapshot"`
	RuleSnapshot       string     `gorm:"column:rule_snapshot;type:longtext" json:"rule_snapshot"`
	TotalScanned       int64      `gorm:"column:total_scanned;not null;default:0" json:"total_scanned"`
	HitArticles        int64      `gorm:"column:hit_articles;not null;default:0" json:"hit_articles"`
	HitCount           int64      `gorm:"column:hit_count;not null;default:0" json:"hit_count"`
	SkipCount          int64      `gorm:"column:skip_count;not null;default:0" json:"skip_count"`
	FailCount          int64      `gorm:"column:fail_count;not null;default:0" json:"fail_count"`
	BatchCount         int64      `gorm:"column:batch_count;not null;default:0" json:"batch_count"`
	StartedAt          *time.Time `gorm:"column:started_at" json:"started_at"`
	FinishedAt         *time.Time `gorm:"column:finished_at" json:"finished_at"`
	DurationMS         int64      `gorm:"column:duration_ms;not null;default:0" json:"duration_ms"`
	ErrorMessage       string     `gorm:"column:error_message;type:text" json:"error_message"`
	CreatorID          uint64     `gorm:"column:creator_id;not null;default:0" json:"creator_id"`
	CreatorName        string     `gorm:"column:creator_name;size:128;not null;default:''" json:"creator_name"`
	InspectionTimestamps
}

func (InspectionTask) TableName() string {
	return "xt_article_inspect_tasks"
}

type InspectionTaskKeyword struct {
	ID        uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrgID     uint64 `gorm:"column:orgid;not null;index" json:"orgid"`
	TaskID    uint64 `gorm:"column:task_id;not null;index" json:"task_id"`
	KeywordID uint64 `gorm:"column:keyword_id;not null;index" json:"keyword_id"`
	InspectionTimestamps
}

func (InspectionTaskKeyword) TableName() string {
	return "xt_article_inspect_task_keywords"
}

type InspectionTaskOutboxMessage struct {
	ID             uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrgID          uint64     `gorm:"column:orgid;not null;index" json:"orgid"`
	TaskID         uint64     `gorm:"column:task_id;not null;index" json:"task_id"`
	MessageType    string     `gorm:"column:message_type;size:64;not null" json:"message_type"`
	Status         string     `gorm:"column:status;size:32;not null;index" json:"status"`
	Payload        string     `gorm:"column:payload;type:longtext" json:"payload"`
	AttemptCount   int64      `gorm:"column:attempt_count;not null;default:0" json:"attempt_count"`
	ClaimedBy      string     `gorm:"column:claimed_by;size:64;not null;default:''" json:"claimed_by"`
	ClaimedAt      *time.Time `gorm:"column:claimed_at" json:"claimed_at"`
	ClaimUntil     *time.Time `gorm:"column:claim_until" json:"claim_until"`
	NextAttemptAt  *time.Time `gorm:"column:next_attempt_at" json:"next_attempt_at"`
	LastError      string     `gorm:"column:last_error;type:text" json:"last_error"`
	LastErrorCode  string     `gorm:"column:last_error_code;size:64;not null;default:''" json:"last_error_code"`
	LastAttemptAt  *time.Time `gorm:"column:last_attempt_at" json:"last_attempt_at"`
	DeadLetteredAt *time.Time `gorm:"column:dead_lettered_at" json:"dead_lettered_at"`
	DispatchedAt   *time.Time `gorm:"column:dispatched_at" json:"dispatched_at"`
	RetainedUntil  *time.Time `gorm:"column:retained_until" json:"retained_until"`
	InspectionTimestamps
}

func (InspectionTaskOutboxMessage) TableName() string {
	return "xt_article_inspect_task_outbox"
}

type InspectionResult struct {
	ID                 uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrgID              uint64     `gorm:"column:orgid;not null;index" json:"orgid"`
	TaskID             uint64     `gorm:"column:task_id;not null;index" json:"task_id"`
	ArticleID          uint64     `gorm:"column:article_id;not null;index" json:"article_id"`
	ArticleTitle       string     `gorm:"column:article_title;size:255;not null;default:''" json:"article_title"`
	ArticleState       int8       `gorm:"column:article_state;not null;default:0" json:"article_state"`
	PublishAtTime      *time.Time `gorm:"column:publish_at_time" json:"publish_at_time"`
	RiskLevel          string     `gorm:"column:risk_level;size:32;not null" json:"risk_level"`
	SuggestAction      string     `gorm:"column:suggest_action;size:32;not null" json:"suggest_action"`
	HitFieldsCount     int64      `gorm:"column:hit_fields_count;not null;default:0" json:"hit_fields_count"`
	HitKeywordsCount   int64      `gorm:"column:hit_keywords_count;not null;default:0" json:"hit_keywords_count"`
	HitCount           int64      `gorm:"column:hit_count;not null;default:0" json:"hit_count"`
	DispositionStatus  string     `gorm:"column:disposition_status;size:32;not null" json:"disposition_status"`
	LatestActionID     uint64     `gorm:"column:latest_action_id;not null;default:0" json:"latest_action_id"`
	LatestOperatorID   uint64     `gorm:"column:latest_operator_id;not null;default:0" json:"latest_operator_id"`
	LatestOperatorName string     `gorm:"column:latest_operator_name;size:128;not null;default:''" json:"latest_operator_name"`
	LatestActionAt     *time.Time `gorm:"column:latest_action_at" json:"latest_action_at"`
	InspectionTimestamps
}

func (InspectionResult) TableName() string {
	return "xt_article_inspect_results"
}

type InspectionResultHit struct {
	ID            uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrgID         uint64 `gorm:"column:orgid;not null;index" json:"orgid"`
	TaskID        uint64 `gorm:"column:task_id;not null;index" json:"task_id"`
	ResultID      uint64 `gorm:"column:result_id;not null;index" json:"result_id"`
	ArticleID     uint64 `gorm:"column:article_id;not null;index" json:"article_id"`
	KeywordID     uint64 `gorm:"column:keyword_id;not null;index" json:"keyword_id"`
	KeywordText   string `gorm:"column:keyword_text;size:255;not null;default:''" json:"keyword_text"`
	Category      string `gorm:"column:category;size:64;not null;default:''" json:"category"`
	FieldName     string `gorm:"column:field_name;size:64;not null;default:''" json:"field_name"`
	MatchType     string `gorm:"column:match_type;size:32;not null" json:"match_type"`
	RiskLevel     string `gorm:"column:risk_level;size:32;not null" json:"risk_level"`
	SuggestAction string `gorm:"column:suggest_action;size:32;not null" json:"suggest_action"`
	MatchedText   string `gorm:"column:matched_text;type:text" json:"matched_text"`
	Snippet       string `gorm:"column:snippet;type:text" json:"snippet"`
	PositionStart int64  `gorm:"column:position_start;not null;default:0" json:"position_start"`
	PositionEnd   int64  `gorm:"column:position_end;not null;default:0" json:"position_end"`
	RuleSnapshot  string `gorm:"column:rule_snapshot;type:longtext" json:"rule_snapshot"`
	InspectionTimestamps
}

func (InspectionResultHit) TableName() string {
	return "xt_article_inspect_result_hits"
}

type InspectionAction struct {
	ID              uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrgID           uint64     `gorm:"column:orgid;not null;index" json:"orgid"`
	ActionNo        string     `gorm:"column:action_no;size:64;not null;uniqueIndex" json:"action_no"`
	ActionType      string     `gorm:"column:action_type;size:32;not null" json:"action_type"`
	TaskID          uint64     `gorm:"column:task_id;not null;default:0;index" json:"task_id"`
	BatchScope      string     `gorm:"column:batch_scope;size:32;not null;default:''" json:"batch_scope"`
	TargetCount     int64      `gorm:"column:target_count;not null;default:0" json:"target_count"`
	SuccessCount    int64      `gorm:"column:success_count;not null;default:0" json:"success_count"`
	FailCount       int64      `gorm:"column:fail_count;not null;default:0" json:"fail_count"`
	SkipCount       int64      `gorm:"column:skip_count;not null;default:0" json:"skip_count"`
	Status          string     `gorm:"column:status;size:32;not null" json:"status"`
	Reason          string     `gorm:"column:reason;type:text" json:"reason"`
	RequestSnapshot string     `gorm:"column:request_snapshot;type:longtext" json:"request_snapshot"`
	OperatorID      uint64     `gorm:"column:operator_id;not null;default:0" json:"operator_id"`
	OperatorName    string     `gorm:"column:operator_name;size:128;not null;default:''" json:"operator_name"`
	RequestID       string     `gorm:"column:request_id;size:128;not null;default:''" json:"request_id"`
	SourceIP        string     `gorm:"column:source_ip;size:64;not null;default:''" json:"source_ip"`
	StartedAt       *time.Time `gorm:"column:started_at" json:"started_at"`
	FinishedAt      *time.Time `gorm:"column:finished_at" json:"finished_at"`
	InspectionTimestamps
}

func (InspectionAction) TableName() string {
	return "xt_article_inspect_actions"
}

type InspectionOperationLog struct {
	ID              uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrgID           uint64 `gorm:"column:orgid;not null;index" json:"orgid"`
	ActionID        uint64 `gorm:"column:action_id;not null;default:0;index" json:"action_id"`
	TaskID          uint64 `gorm:"column:task_id;not null;default:0;index" json:"task_id"`
	ResultID        uint64 `gorm:"column:result_id;not null;default:0;index" json:"result_id"`
	ArticleID       uint64 `gorm:"column:article_id;not null;default:0;index" json:"article_id"`
	OperationType   string `gorm:"column:operation_type;size:32;not null" json:"operation_type"`
	BeforeState     string `gorm:"column:before_state;size:32;not null;default:''" json:"before_state"`
	AfterState      string `gorm:"column:after_state;size:32;not null;default:''" json:"after_state"`
	Status          string `gorm:"column:status;size:32;not null" json:"status"`
	Reason          string `gorm:"column:reason;type:text" json:"reason"`
	Summary         string `gorm:"column:summary;type:text" json:"summary"`
	RequestSnapshot string `gorm:"column:request_snapshot;type:longtext" json:"request_snapshot"`
	OperatorID      uint64 `gorm:"column:operator_id;not null;default:0" json:"operator_id"`
	OperatorName    string `gorm:"column:operator_name;size:128;not null;default:''" json:"operator_name"`
	RequestID       string `gorm:"column:request_id;size:128;not null;default:''" json:"request_id"`
	SourceIP        string `gorm:"column:source_ip;size:64;not null;default:''" json:"source_ip"`
	InspectionTimestamps
}

func (InspectionOperationLog) TableName() string {
	return "xt_article_inspect_operation_logs"
}

type InspectionFieldChangeLog struct {
	ID           uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrgID        uint64 `gorm:"column:orgid;not null;index" json:"orgid"`
	ActionID     uint64 `gorm:"column:action_id;not null;default:0;index" json:"action_id"`
	TaskID       uint64 `gorm:"column:task_id;not null;default:0;index" json:"task_id"`
	ResultID     uint64 `gorm:"column:result_id;not null;default:0;index" json:"result_id"`
	ArticleID    uint64 `gorm:"column:article_id;not null;default:0;index" json:"article_id"`
	FieldName    string `gorm:"column:field_name;size:64;not null" json:"field_name"`
	BeforeValue  string `gorm:"column:before_value;type:longtext" json:"before_value"`
	AfterValue   string `gorm:"column:after_value;type:longtext" json:"after_value"`
	DiffSummary  string `gorm:"column:diff_summary;type:text" json:"diff_summary"`
	OperatorID   uint64 `gorm:"column:operator_id;not null;default:0" json:"operator_id"`
	OperatorName string `gorm:"column:operator_name;size:128;not null;default:''" json:"operator_name"`
	RequestID    string `gorm:"column:request_id;size:128;not null;default:''" json:"request_id"`
	SourceIP     string `gorm:"column:source_ip;size:64;not null;default:''" json:"source_ip"`
	InspectionTimestamps
}

func (InspectionFieldChangeLog) TableName() string {
	return "xt_article_inspect_field_change_logs"
}

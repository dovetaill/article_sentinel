package lifecycle

type OfflineArticleInput struct {
	OrgID        uint64
	ArticleID    uint64
	TaskID       uint64
	ResultID     uint64
	ActionID     uint64
	OperatorID   uint64
	OperatorName string
	Reason       string
}

type UpdateArticleFieldsInput struct {
	OrgID        uint64
	ArticleID    uint64
	TaskID       uint64
	ResultID     uint64
	ActionID     uint64
	OperatorID   uint64
	OperatorName string
	Reason       string
	Fields       EditableArticleFields
}

type RepublishArticleInput struct {
	OrgID        uint64
	ArticleID    uint64
	TaskID       uint64
	ResultID     uint64
	ActionID     uint64
	OperatorID   uint64
	OperatorName string
	Reason       string
}

type EditableArticleFields struct {
	Title      string
	ShortTitle string
	RichTitle  string
	Keyword    string
	Desc       string
	Body       string
}

type FieldChange struct {
	FieldName   string `json:"field_name"`
	BeforeValue string `json:"before_value"`
	AfterValue  string `json:"after_value"`
	DiffSummary string `json:"diff_summary"`
}

type LifecycleActionResult struct {
	Status      string `json:"status"`
	ArticleID   uint64 `json:"article_id"`
	BeforeState int8   `json:"before_state"`
	AfterState  int8   `json:"after_state"`
}

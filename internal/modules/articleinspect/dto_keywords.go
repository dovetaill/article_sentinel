package articleinspect

type CreateKeywordInput struct {
	OrgID         uint64
	Name          string
	CategoryID    uint64
	MatchType     string
	RiskLevel     string
	SuggestAction string
	Enabled       bool
	Remark        string
	Scopes        []string
}

type UpdateKeywordInput struct {
	ID uint64
	CreateKeywordInput
}

type PatchKeywordStatusInput struct {
	OrgID     uint64
	KeywordID uint64
	Enabled   bool
}

type KeywordListInput struct {
	OrgID      uint64
	Page       int
	PageSize   int
	Enabled    *bool
	CategoryID uint64
	Query      string
}

type KeywordDTO struct {
	ID            uint64   `json:"id"`
	OrgID         uint64   `json:"orgid"`
	Name          string   `json:"name"`
	CategoryID    uint64   `json:"category_id"`
	CategoryName  string   `json:"category_name"`
	MatchType     string   `json:"match_type"`
	RiskLevel     string   `json:"risk_level"`
	SuggestAction string   `json:"suggest_action"`
	Enabled       bool     `json:"enabled"`
	Remark        string   `json:"remark"`
	Scopes        []string `json:"scopes"`
	CreatorID     uint64   `json:"creator_id"`
	CreatorName   string   `json:"creator_name"`
	UpdaterID     uint64   `json:"updater_id"`
	UpdaterName   string   `json:"updater_name"`
}

type KeywordListResult struct {
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Total    int64        `json:"total"`
	Items    []KeywordDTO `json:"items"`
}

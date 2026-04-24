package articleinspect

import "time"

type ArticleListInput struct {
	OrgID    uint64
	Page     int
	PageSize int
	State    *int8
	Query    string
}

type ArticleListItem struct {
	ID                  uint64     `json:"id"`
	OrgID               uint64     `json:"orgid"`
	Title               string     `json:"title"`
	Thumbnail           string     `json:"thumbnail,omitempty"`
	State               int8       `json:"state"`
	PublishAtTime       *time.Time `json:"publish_at_time"`
	LatestRiskLevel     string     `json:"latest_risk_level,omitempty"`
	LatestTaskID        uint64     `json:"latest_task_id,omitempty"`
	LatestResultID      uint64     `json:"latest_result_id,omitempty"`
	LatestSuggestAction string     `json:"latest_suggest_action,omitempty"`
	LatestDisposition   string     `json:"latest_disposition_status,omitempty"`
	LatestOperatorName  string     `json:"latest_operator_name,omitempty"`
	LatestActionAt      *time.Time `json:"latest_action_at,omitempty"`
}

type ArticleListResult struct {
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int64             `json:"total"`
	Items    []ArticleListItem `json:"items"`
}

type ArticleDetail struct {
	ID                  uint64     `json:"id"`
	OrgID               uint64     `json:"orgid"`
	Title               string     `json:"title"`
	ShortTitle          string     `json:"short_title"`
	RichTitle           string     `json:"rich_title"`
	Keyword             string     `json:"keyword"`
	Desc                string     `json:"desc"`
	Body                string     `json:"body"`
	Thumbnail           string     `json:"thumbnail,omitempty"`
	State               int8       `json:"state"`
	PublishAtTime       *time.Time `json:"publish_at_time"`
	LatestRiskLevel     string     `json:"latest_risk_level,omitempty"`
	LatestTaskID        uint64     `json:"latest_task_id,omitempty"`
	LatestResultID      uint64     `json:"latest_result_id,omitempty"`
	LatestSuggestAction string     `json:"latest_suggest_action,omitempty"`
	LatestDisposition   string     `json:"latest_disposition_status,omitempty"`
	LatestOperatorName  string     `json:"latest_operator_name,omitempty"`
	LatestActionAt      *time.Time `json:"latest_action_at,omitempty"`
}

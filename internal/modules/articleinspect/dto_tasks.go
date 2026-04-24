package articleinspect

import "time"

type CandidateArticleFilter struct {
	OrgID            uint64
	ArticleState     int8
	PublishTimeStart *time.Time
	PublishTimeEnd   *time.Time
	ArticleID        uint64
	TitleLike        string
	AfterID          uint64
	Limit            int
}

type TaskListInput struct {
	OrgID    uint64
	Page     int
	PageSize int
	Status   string
	TaskNo   string
}

type TaskListResult struct {
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int64            `json:"total"`
	Items    []InspectionTask `json:"items"`
}

type CreateInspectionTaskInput struct {
	OrgID            uint64     `json:"orgid"`
	KeywordIDs       []uint64   `json:"keyword_ids"`
	PublishTimeStart *time.Time `json:"publish_time_start,omitempty"`
	PublishTimeEnd   *time.Time `json:"publish_time_end,omitempty"`
	ArticleID        uint64     `json:"article_id,omitempty"`
	TitleLike        string     `json:"title_like,omitempty"`
	IncludeBody      bool       `json:"include_body,omitempty"`
	ArticleState     int8       `json:"article_state,omitempty"`
}

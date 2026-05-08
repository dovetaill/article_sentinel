package results

import domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"

type ResultListInput struct {
	OrgID             uint64
	TaskID            uint64
	RiskLevel         string
	DispositionStatus string
	TitleLike         string
	ArticleID         uint64
	Page              int
	PageSize          int
}

type ResultListItem struct {
	domainpkg.InspectionResult
	PreviewFieldName   string `json:"preview_field_name,omitempty"`
	PreviewKeywordText string `json:"preview_keyword_text,omitempty"`
	PreviewMatchedText string `json:"preview_matched_text,omitempty"`
	PreviewSnippet     string `json:"preview_snippet,omitempty"`
	ExtraHitCount      int64  `json:"extra_hit_count"`
}

type ResultListResult struct {
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int64            `json:"total"`
	Items    []ResultListItem `json:"items"`
}

type ResultDetail struct {
	Result          domainpkg.InspectionResult           `json:"result"`
	Hits            []domainpkg.InspectionResultHit      `json:"hits"`
	OperationLogs   []domainpkg.InspectionOperationLog   `json:"operation_logs"`
	FieldChangeLogs []domainpkg.InspectionFieldChangeLog `json:"field_change_logs"`
}

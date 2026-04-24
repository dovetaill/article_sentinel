package articleinspect

type ResultListItem struct {
	InspectionResult
	PreviewFieldName   string `json:"preview_field_name,omitempty"`
	PreviewKeywordText string `json:"preview_keyword_text,omitempty"`
	PreviewMatchedText string `json:"preview_matched_text,omitempty"`
	PreviewSnippet     string `json:"preview_snippet,omitempty"`
	ExtraHitCount      int64  `json:"extra_hit_count"`
}


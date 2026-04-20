package articleinspect

import "strings"

type EditableArticleFields struct {
	Title      string
	ShortTitle string
	RichTitle  string
	Keyword    string
	Desc       string
	Body       string
}

type FieldChange struct {
	FieldName   string
	BeforeValue string
	AfterValue  string
	DiffSummary string
}

func DiffEditableFields(before, after EditableArticleFields) []FieldChange {
	changes := make([]FieldChange, 0, 6)
	changes = appendFieldChange(changes, KeywordScopeTitle, before.Title, after.Title)
	changes = appendFieldChange(changes, KeywordScopeShortTitle, before.ShortTitle, after.ShortTitle)
	changes = appendFieldChange(changes, KeywordScopeRichTitle, before.RichTitle, after.RichTitle)
	changes = appendFieldChange(changes, KeywordScopeKeyword, before.Keyword, after.Keyword)
	changes = appendFieldChange(changes, KeywordScopeDesc, before.Desc, after.Desc)
	changes = appendFieldChange(changes, KeywordScopeBody, before.Body, after.Body)
	return changes
}

func appendFieldChange(changes []FieldChange, fieldName, beforeValue, afterValue string) []FieldChange {
	if beforeValue == afterValue {
		return changes
	}
	return append(changes, FieldChange{
		FieldName:   fieldName,
		BeforeValue: beforeValue,
		AfterValue:  afterValue,
		DiffSummary: summarizeFieldDiff(fieldName, beforeValue, afterValue),
	})
}

func summarizeFieldDiff(fieldName, beforeValue, afterValue string) string {
	const maxSegment = 48
	beforeSummary := truncateDiffText(beforeValue, maxSegment)
	afterSummary := truncateDiffText(afterValue, maxSegment)
	return fieldName + ": " + beforeSummary + " -> " + afterSummary
}

func truncateDiffText(value string, maxLength int) string {
	value = strings.TrimSpace(value)
	if maxLength <= 0 || len(value) <= maxLength {
		return value
	}
	if maxLength <= 3 {
		return value[:maxLength]
	}
	return value[:maxLength-3] + "..."
}

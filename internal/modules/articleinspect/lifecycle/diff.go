package lifecycle

import (
	"strings"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
)

func DiffEditableFields(before, after EditableArticleFields) []FieldChange {
	changes := make([]FieldChange, 0, 6)
	changes = appendFieldChange(changes, domainpkg.KeywordScopeTitle, before.Title, after.Title)
	changes = appendFieldChange(changes, domainpkg.KeywordScopeShortTitle, before.ShortTitle, after.ShortTitle)
	changes = appendFieldChange(changes, domainpkg.KeywordScopeRichTitle, before.RichTitle, after.RichTitle)
	changes = appendFieldChange(changes, domainpkg.KeywordScopeKeyword, before.Keyword, after.Keyword)
	changes = appendFieldChange(changes, domainpkg.KeywordScopeDesc, before.Desc, after.Desc)
	changes = appendFieldChange(changes, domainpkg.KeywordScopeBody, before.Body, after.Body)
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

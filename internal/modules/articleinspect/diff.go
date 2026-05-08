package articleinspect

import lifecyclepkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/lifecycle"

type EditableArticleFields = lifecyclepkg.EditableArticleFields

type FieldChange = lifecyclepkg.FieldChange

func DiffEditableFields(before, after EditableArticleFields) []FieldChange {
	return lifecyclepkg.DiffEditableFields(before, after)
}

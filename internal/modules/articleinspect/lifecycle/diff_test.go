package lifecycle

import (
	"testing"

	domainpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/domain"
	"github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
)

func TestFieldDiff(t *testing.T) {
	stamp := testutil.MustTime(t, "2026-04-20T10:00:00Z").Format("20060102")
	before := EditableArticleFields{
		Title: "Old title " + stamp,
		Body:  "old body content that is deliberately long so the diff summary needs truncation when rendered",
	}
	after := EditableArticleFields{
		Title: "New title " + stamp,
		Body:  "new body content that is deliberately long so the diff summary needs truncation when rendered differently",
	}

	changes := DiffEditableFields(before, after)
	if len(changes) != 2 {
		t.Fatalf("DiffEditableFields() len = %d, want %d", len(changes), 2)
	}

	var bodyChange *FieldChange
	for index := range changes {
		if changes[index].FieldName == domainpkg.KeywordScopeBody {
			bodyChange = &changes[index]
		}
	}
	if bodyChange == nil {
		t.Fatal("DiffEditableFields() missing body change")
	}
	if len(bodyChange.DiffSummary) == 0 {
		t.Fatal("body diff summary = empty, want non-empty")
	}

	if got := DiffEditableFields(before, before); len(got) != 0 {
		t.Fatalf("DiffEditableFields(no-op) len = %d, want %d", len(got), 0)
	}
}

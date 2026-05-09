package articleinspect

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestArticleInspectRootStaysThin(t *testing.T) {
	entries, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	want := map[string]struct{}{
		"module.go": {},
		"routes.go": {},
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry, "_test.go") {
			continue
		}
		name := filepath.Base(entry)
		if _, ok := want[name]; ok {
			continue
		}
		t.Fatalf("unexpected root go file: %s", name)
	}
}

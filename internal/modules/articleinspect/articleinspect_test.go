package articleinspect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectionDocsArtifactsExist(t *testing.T) {
	requiredArtifacts := []string{
		filepath.Join("..", "..", "..", "docs", "article-inspection-api.md"),
		filepath.Join("..", "..", "..", "docs", "article-inspection-pages.md"),
		filepath.Join("..", "..", "..", "scripts", "article_inspection_seed.sql"),
	}

	for _, artifact := range requiredArtifacts {
		info, err := os.Stat(artifact)
		if err != nil {
			t.Fatalf("Stat(%q) error = %v", artifact, err)
		}
		if info.Size() == 0 {
			t.Fatalf("artifact %q is empty", artifact)
		}
	}
}

package scripts

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestArticleCenterSeedProvides100Org29Articles(t *testing.T) {
	path := filepath.Join("article_center_seed_100.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read seed file: %v", err)
	}

	sql := string(content)
	if !strings.Contains(sql, "DELETE FROM `xt_article_info` WHERE `article_id` BETWEEN 9110001 AND 9110100;") {
		t.Fatalf("seed missing xt_article_info cleanup for 9110001-9110100 range")
	}
	if !strings.Contains(sql, "DELETE FROM `xt_article` WHERE `id` BETWEEN 9110001 AND 9110100;") {
		t.Fatalf("seed missing xt_article cleanup for 9110001-9110100 range")
	}

	articleInsert := regexp.MustCompile("(?s)INSERT INTO `xt_article` .*? VALUES\\s*(.*?);")
	articleMatch := articleInsert.FindStringSubmatch(sql)
	if len(articleMatch) != 2 {
		t.Fatalf("seed missing xt_article insert block")
	}

	infoInsert := regexp.MustCompile("(?s)INSERT INTO `xt_article_info` .*? VALUES\\s*(.*?);")
	infoMatch := infoInsert.FindStringSubmatch(sql)
	if len(infoMatch) != 2 {
		t.Fatalf("seed missing xt_article_info insert block")
	}

	articleRowPattern := regexp.MustCompile(`\(\s*9110\d{3}, 29, `)
	if got := len(articleRowPattern.FindAllString(articleMatch[1], -1)); got != 100 {
		t.Fatalf("xt_article row count = %d, want %d", got, 100)
	}
	infoRowPattern := regexp.MustCompile(`\(\s*9110\d{3}, `)
	if got := len(infoRowPattern.FindAllString(infoMatch[1], -1)); got != 100 {
		t.Fatalf("xt_article_info row count = %d, want %d", got, 100)
	}

	if strings.Count(articleMatch[1], ", 29, ") != 100 {
		t.Fatalf("xt_article orgid 29 row count = %d, want %d", strings.Count(articleMatch[1], ", 29, "), 100)
	}
}

package response_test

import (
	"testing"

	"github.com/dovetaill/article-sentinel/internal/api/response"
)

func TestPagedReturnsStandardShape(t *testing.T) {
	got := response.Paged("user list", 2, 50, 101, []map[string]any{
		{"id": 1, "username": "alice"},
	})

	if got.Code != 0 {
		t.Fatalf("code = %d, want %d", got.Code, 0)
	}
	if got.Message != "user list" {
		t.Fatalf("message = %q, want %q", got.Message, "user list")
	}

	data, ok := got.Data.(map[string]any)
	if !ok {
		t.Fatalf("data type = %T, want map[string]any", got.Data)
	}
	if data["page"] != 2 {
		t.Fatalf("page = %v, want %d", data["page"], 2)
	}
	if data["page_size"] != 50 {
		t.Fatalf("page_size = %v, want %d", data["page_size"], 50)
	}
	if data["total"] != int64(101) {
		t.Fatalf("total = %v, want %d", data["total"], int64(101))
	}

	items, ok := data["items"].([]map[string]any)
	if !ok {
		t.Fatalf("items type = %T, want []map[string]any", data["items"])
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want %d", len(items), 1)
	}
	if items[0]["username"] != "alice" {
		t.Fatalf("items[0].username = %v, want %q", items[0]["username"], "alice")
	}
}

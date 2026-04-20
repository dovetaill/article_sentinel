package post_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/article-sentinel/internal/api/response"
	postmodule "github.com/dovetaill/article-sentinel/internal/modules/post"
)

func TestPostModelUsesPortableContentColumnType(t *testing.T) {
	field, ok := reflect.TypeOf(postmodule.Post{}).FieldByName("Content")
	if !ok {
		t.Fatal("Content field not found")
	}

	if got := field.Tag.Get("gorm"); strings.Contains(got, "longtext") {
		t.Fatalf("gorm tag = %q, must avoid postgres-incompatible longtext", got)
	}
}

type postRepoStub struct {
	nextID uint
	items  map[uint]*postmodule.Post
}

func newPostRepoStub(items ...*postmodule.Post) *postRepoStub {
	stub := &postRepoStub{nextID: 1, items: make(map[uint]*postmodule.Post, len(items))}
	for _, item := range items {
		if item == nil {
			continue
		}
		clone := *item
		stub.items[item.ID] = &clone
		if item.ID >= stub.nextID {
			stub.nextID = item.ID + 1
		}
	}
	return stub
}

func (s *postRepoStub) Create(ctx context.Context, item *postmodule.Post) error {
	for _, existing := range s.items {
		if existing.Slug == item.Slug {
			return postmodule.ErrDuplicatePostSlug
		}
	}
	clone := *item
	if clone.ID == 0 {
		clone.ID = s.nextID
		s.nextID++
	}
	clone.CreatedAt = time.Now()
	clone.UpdatedAt = clone.CreatedAt
	s.items[clone.ID] = &clone
	*item = clone
	return nil
}

func (s *postRepoStub) List(ctx context.Context, page, pageSize int) ([]postmodule.Post, int64, error) {
	ids := make([]int, 0, len(s.items))
	for id := range s.items {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)

	all := make([]postmodule.Post, 0, len(ids))
	for _, id := range ids {
		all = append(all, *s.items[uint(id)])
	}

	total := int64(len(all))
	start := (page - 1) * pageSize
	if start >= len(all) {
		return []postmodule.Post{}, total, nil
	}
	end := start + pageSize
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], total, nil
}

func (s *postRepoStub) FindByID(ctx context.Context, id uint) (*postmodule.Post, error) {
	item, ok := s.items[id]
	if !ok {
		return nil, postmodule.ErrPostNotFound
	}
	clone := *item
	return &clone, nil
}

func (s *postRepoStub) FindBySlug(ctx context.Context, slug string) (*postmodule.Post, error) {
	for _, item := range s.items {
		if item.Slug != slug {
			continue
		}
		clone := *item
		return &clone, nil
	}
	return nil, postmodule.ErrPostNotFound
}

func (s *postRepoStub) Update(ctx context.Context, item *postmodule.Post) error {
	if _, ok := s.items[item.ID]; !ok {
		return postmodule.ErrPostNotFound
	}
	for _, existing := range s.items {
		if existing.ID != item.ID && existing.Slug == item.Slug {
			return postmodule.ErrDuplicatePostSlug
		}
	}
	clone := *item
	clone.UpdatedAt = time.Now()
	s.items[item.ID] = &clone
	*item = clone
	return nil
}

func (s *postRepoStub) Delete(ctx context.Context, id uint) error {
	if _, ok := s.items[id]; !ok {
		return postmodule.ErrPostNotFound
	}
	delete(s.items, id)
	return nil
}

func TestPostModuleStarterContract(t *testing.T) {
	handler, repo := newPostHandler(t)

	created := sendJSONRequest(t, handler, http.MethodPost, "/api/v1/posts", map[string]any{
		"title":   "First Post",
		"summary": "intro",
		"content": "hello world",
	})
	if created.status != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", created.status, http.StatusCreated)
	}
	assertSuccessEnvelope(t, created.envelope, "post created")
	createData := mustDataMap(t, created.envelope.Data)
	createdID := uint(numberField(t, createData, "id"))
	if createData["slug"] != "first-post" {
		t.Fatalf("create slug = %v, want %q", createData["slug"], "first-post")
	}

	listed := sendRequest(t, handler, http.MethodGet, "/api/v1/posts?page=1&page_size=20", nil)
	if listed.status != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listed.status, http.StatusOK)
	}
	assertSuccessEnvelope(t, listed.envelope, "post list")
	listData := mustDataMap(t, listed.envelope.Data)
	if got := numberField(t, listData, "page"); got != 1 {
		t.Fatalf("list page = %v, want %d", got, 1)
	}
	if got := numberField(t, listData, "page_size"); got != 20 {
		t.Fatalf("list page_size = %v, want %d", got, 20)
	}
	if got := numberField(t, listData, "total"); got != 1 {
		t.Fatalf("list total = %v, want %d", got, 1)
	}
	items := mustSlice(t, listData["items"])
	if len(items) != 1 {
		t.Fatalf("list items length = %d, want %d", len(items), 1)
	}
	listItem := mustDataMap(t, items[0])
	if listItem["title"] != "First Post" {
		t.Fatalf("list first title = %v, want %q", listItem["title"], "First Post")
	}

	detail := sendRequest(t, handler, http.MethodGet, fmt.Sprintf("/api/v1/posts/%d", createdID), nil)
	if detail.status != http.StatusOK {
		t.Fatalf("detail status = %d, want %d", detail.status, http.StatusOK)
	}
	assertSuccessEnvelope(t, detail.envelope, "post detail")
	detailData := mustDataMap(t, detail.envelope.Data)
	if got := numberField(t, detailData, "id"); got != float64(createdID) {
		t.Fatalf("detail id = %v, want %d", got, createdID)
	}

	updated := sendJSONRequest(t, handler, http.MethodPatch, fmt.Sprintf("/api/v1/posts/%d", createdID), map[string]any{
		"title":   "First Post Updated",
		"summary": "updated summary",
		"content": "updated body",
	})
	if updated.status != http.StatusOK {
		t.Fatalf("patch status = %d, want %d", updated.status, http.StatusOK)
	}
	assertSuccessEnvelope(t, updated.envelope, "post updated")
	updatedData := mustDataMap(t, updated.envelope.Data)
	if updatedData["title"] != "First Post Updated" {
		t.Fatalf("updated title = %v, want %q", updatedData["title"], "First Post Updated")
	}

	deleted := sendRequest(t, handler, http.MethodDelete, fmt.Sprintf("/api/v1/posts/%d", createdID), nil)
	if deleted.status != http.StatusOK {
		t.Fatalf("delete status = %d, want %d", deleted.status, http.StatusOK)
	}
	assertSuccessEnvelope(t, deleted.envelope, "post deleted")
	if len(repo.items) != 0 {
		t.Fatalf("repo item count after delete = %d, want %d", len(repo.items), 0)
	}

	missing := sendRequest(t, handler, http.MethodGet, fmt.Sprintf("/api/v1/posts/%d", createdID), nil)
	if missing.status != http.StatusNotFound {
		t.Fatalf("detail after delete status = %d, want %d", missing.status, http.StatusNotFound)
	}
	assertFailureEnvelope(t, missing.envelope, http.StatusNotFound, "post not found")
}

func TestPostModuleValidationFailures(t *testing.T) {
	handler, _ := newPostHandler(t, &postmodule.Post{ID: 1, Title: "Seed", Slug: "seed", Summary: "seed", Content: "seed"})

	tests := []struct {
		name       string
		method     string
		path       string
		body       any
		wantStatus int
		wantCode   int
		wantMsg    string
	}{
		{
			name:       "create rejects blank title",
			method:     http.MethodPost,
			path:       "/api/v1/posts",
			body:       map[string]any{"title": "   ", "summary": "s", "content": "c"},
			wantStatus: http.StatusBadRequest,
			wantCode:   http.StatusBadRequest,
			wantMsg:    "invalid post input",
		},
		{
			name:       "create rejects duplicate slug",
			method:     http.MethodPost,
			path:       "/api/v1/posts",
			body:       map[string]any{"title": "Another", "slug": "seed", "summary": "s", "content": "c"},
			wantStatus: http.StatusConflict,
			wantCode:   http.StatusConflict,
			wantMsg:    "post slug already exists",
		},
		{
			name:       "detail rejects non numeric id",
			method:     http.MethodGet,
			path:       "/api/v1/posts/not-a-number",
			wantStatus: http.StatusBadRequest,
			wantCode:   http.StatusBadRequest,
			wantMsg:    "invalid post input",
		},
		{
			name:       "patch rejects non numeric id",
			method:     http.MethodPatch,
			path:       "/api/v1/posts/not-a-number",
			body:       map[string]any{"title": "x"},
			wantStatus: http.StatusBadRequest,
			wantCode:   http.StatusBadRequest,
			wantMsg:    "invalid post input",
		},
		{
			name:       "delete rejects non numeric id",
			method:     http.MethodDelete,
			path:       "/api/v1/posts/not-a-number",
			wantStatus: http.StatusBadRequest,
			wantCode:   http.StatusBadRequest,
			wantMsg:    "invalid post input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got httpResult
			if tt.body != nil {
				got = sendJSONRequest(t, handler, tt.method, tt.path, tt.body)
			} else {
				got = sendRequest(t, handler, tt.method, tt.path, nil)
			}

			if got.status != tt.wantStatus {
				t.Fatalf("status = %d, want %d", got.status, tt.wantStatus)
			}
			assertFailureEnvelope(t, got.envelope, tt.wantCode, tt.wantMsg)
		})
	}
}

type httpResult struct {
	status   int
	envelope response.Envelope
}

func newPostHandler(t *testing.T, items ...*postmodule.Post) (http.Handler, *postRepoStub) {
	t.Helper()
	repo := newPostRepoStub(items...)
	service := postmodule.NewService(repo)

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	postmodule.RegisterRoutes(api, service)

	return mux, repo
}

func sendJSONRequest(t *testing.T, handler http.Handler, method, path string, body any) httpResult {
	t.Helper()

	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	return sendRequest(t, handler, method, path, bytes.NewReader(encoded))
}

func sendRequest(t *testing.T, handler http.Handler, method, path string, body *bytes.Reader) httpResult {
	t.Helper()

	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, body)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return httpResult{status: rec.Code, envelope: decodeEnvelope(t, rec)}
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) response.Envelope {
	t.Helper()
	var got response.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return got
}

func assertSuccessEnvelope(t *testing.T, envelope response.Envelope, wantMessage string) {
	t.Helper()
	if envelope.Code != 0 {
		t.Fatalf("envelope.code = %d, want %d", envelope.Code, 0)
	}
	if envelope.Message != wantMessage {
		t.Fatalf("envelope.message = %q, want %q", envelope.Message, wantMessage)
	}
}

func assertFailureEnvelope(t *testing.T, envelope response.Envelope, wantCode int, wantMessage string) {
	t.Helper()
	if envelope.Code != wantCode {
		t.Fatalf("envelope.code = %d, want %d", envelope.Code, wantCode)
	}
	if envelope.Message != wantMessage {
		t.Fatalf("envelope.message = %q, want %q", envelope.Message, wantMessage)
	}
}

func mustDataMap(t *testing.T, value any) map[string]any {
	t.Helper()
	data, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("data type = %T, want map[string]any", value)
	}
	return data
}

func mustSlice(t *testing.T, value any) []any {
	t.Helper()
	items, ok := value.([]any)
	if !ok {
		t.Fatalf("slice type = %T, want []any", value)
	}
	return items
}

func numberField(t *testing.T, m map[string]any, key string) float64 {
	t.Helper()
	v, ok := m[key]
	if !ok {
		t.Fatalf("missing key %q", key)
	}
	n, ok := v.(float64)
	if !ok {
		t.Fatalf("key %q type = %T, want float64", key, v)
	}
	return n
}

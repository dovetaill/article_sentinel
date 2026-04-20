package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type memorySlogHandler struct {
	mu      sync.Mutex
	records []map[string]any
}

func (h *memorySlogHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *memorySlogHandler) Handle(_ context.Context, record slog.Record) error {
	entry := map[string]any{"msg": record.Message}
	record.Attrs(func(attr slog.Attr) bool {
		entry[attr.Key] = attr.Value.Any()
		return true
	})

	h.mu.Lock()
	h.records = append(h.records, entry)
	h.mu.Unlock()
	return nil
}

func (h *memorySlogHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *memorySlogHandler) WithGroup(_ string) slog.Handler      { return h }

func TestAccessLogCapturesStatusAndRequestID(t *testing.T) {
	memory := &memorySlogHandler{}
	logger := slog.New(memory)

	handler := Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}),
		RequestID(),
		AccessLog(logger),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if len(memory.records) != 1 {
		t.Fatalf("record count = %d, want %d", len(memory.records), 1)
	}

	record := memory.records[0]
	if record["msg"] != "http_access" {
		t.Fatalf("msg = %v, want %q", record["msg"], "http_access")
	}
	if record["method"] != http.MethodPost {
		t.Fatalf("method = %v, want %q", record["method"], http.MethodPost)
	}
	if record["path"] != "/api/v1/posts" {
		t.Fatalf("path = %v, want %q", record["path"], "/api/v1/posts")
	}
	if got, ok := record["status_code"].(int64); !ok || got != http.StatusCreated {
		t.Fatalf("status_code = %v, want %d", record["status_code"], http.StatusCreated)
	}
	if _, ok := record["duration_ms"].(int64); !ok {
		t.Fatalf("duration_ms type = %T, want int64", record["duration_ms"])
	}
	if requestID, ok := record["request_id"].(string); !ok || requestID == "" {
		t.Fatalf("request_id = %v, want non-empty string", record["request_id"])
	}
}

func TestAccessLogDefaultsStatusCodeTo200(t *testing.T) {
	memory := &memorySlogHandler{}
	logger := slog.New(memory)

	handler := AccessLog(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(memory.records) != 1 {
		t.Fatalf("record count = %d, want %d", len(memory.records), 1)
	}
	if got, ok := memory.records[0]["status_code"].(int64); !ok || got != http.StatusOK {
		t.Fatalf("status_code = %v, want %d", memory.records[0]["status_code"], http.StatusOK)
	}
}

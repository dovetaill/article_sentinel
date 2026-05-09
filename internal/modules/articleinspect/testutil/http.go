package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/dovetaill/article-sentinel/internal/api/response"
	"github.com/dovetaill/article-sentinel/internal/identity"
)

type HTTPResult struct {
	Status   int
	Envelope response.Envelope
}

type RequestOptions struct {
	Anonymous bool
	Session   *identity.AdminSession
}

func SendJSONRequest(t *testing.T, handler http.Handler, method, path string, body any) HTTPResult {
	t.Helper()
	return SendJSONRequestWithOptions(t, handler, method, path, body, RequestOptions{})
}

func SendJSONRequestWithOptions(t *testing.T, handler http.Handler, method, path string, body any, options RequestOptions) HTTPResult {
	t.Helper()

	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	return SendRequestWithOptions(t, handler, method, path, bytes.NewReader(encoded), options)
}

func SendRequest(t *testing.T, handler http.Handler, method, path string, body io.Reader) HTTPResult {
	t.Helper()
	return SendRequestWithOptions(t, handler, method, path, body, RequestOptions{})
}

func SendRequestWithOptions(t *testing.T, handler http.Handler, method, path string, body io.Reader, options RequestOptions) HTTPResult {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = io.ReadAll(body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
	}

	var req *http.Request
	if payload != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	if !options.Anonymous {
		session := options.Session
		if session == nil {
			derived := DeriveAdminSession(t, path, payload)
			session = &derived
		}
		actor := session.Actor()
		ctx := identity.ContextWithAdminSession(req.Context(), *session)
		ctx = identity.ContextWithActor(ctx, actor)
		ctx = identity.ContextWithPrincipal(ctx, identity.PrincipalFromActor(actor))
		req = req.WithContext(ctx)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return HTTPResult{Status: rec.Code, Envelope: DecodeEnvelope(t, rec)}
}

func DeriveAdminSession(t *testing.T, requestPath string, payload []byte) identity.AdminSession {
	t.Helper()

	orgID := uint64(29)
	if parsedURL, err := url.Parse(requestPath); err == nil {
		if raw := strings.TrimSpace(parsedURL.Query().Get("orgid")); raw != "" {
			if parsed, parseErr := strconv.ParseUint(raw, 10, 64); parseErr == nil && parsed > 0 {
				orgID = parsed
			}
		}
	}

	if len(payload) > 0 {
		var body map[string]any
		if err := json.Unmarshal(payload, &body); err == nil {
			if parsed := BodyOrgID(body); parsed > 0 {
				orgID = parsed
			}
		}
	}

	return identity.AdminSession{
		UserID:   7,
		OrgID:    orgID,
		OrgName:  fmt.Sprintf("org-%d", orgID),
		Nickname: "alice",
		Priv:     "admin",
		Status:   "active",
	}
}

func BodyOrgID(body map[string]any) uint64 {
	value, ok := body["orgid"]
	if !ok {
		return 0
	}

	switch typed := value.(type) {
	case float64:
		return uint64(typed)
	case string:
		parsed, err := strconv.ParseUint(strings.TrimSpace(typed), 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func DecodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) response.Envelope {
	t.Helper()

	body := bytes.TrimSpace(rec.Body.Bytes())
	if len(body) == 0 || body[0] != '{' {
		return response.Envelope{}
	}

	var got response.Envelope
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return got
}

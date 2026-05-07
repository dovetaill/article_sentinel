package identity

import (
	"testing"
	"time"

	"github.com/dovetaill/go-auth-demo/pkg/config"
	jwt "github.com/golang-jwt/jwt/v5"
)

func TestAdminSessionManagerExchangesLegacyJWT(t *testing.T) {
	now := time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC)
	mgr := NewAdminSessionManager(config.SessionConfig{
		LegacySecret: "legacy-secret",
		Secret:       "session-secret",
		Issuer:       "go-auth-demo",
		TTLHours:     24,
	})
	mgr.now = func() time.Time { return now }

	legacy := signedLegacyJWT(t, "legacy-secret", map[string]any{
		"id":           "90525",
		"orgid":        "29",
		"orgname":      "一县一端测试机构",
		"platform":     "chuangqi",
		"priv":         "super",
		"roleid":       "1",
		"nickname":     "用户A",
		"avatar":       "https://example.com/a.png",
		"departmentid": "3",
		"is_open_edu":  true,
		"exp":          now.Add(time.Hour).Unix(),
		"nbf":          now.Add(-time.Minute).Unix(),
		"iat":          now.Add(-time.Minute).Unix(),
		"iss":          "legacy-admin",
		"sub":          "90525",
	})

	token, session, expiresAt, err := mgr.ExchangeLegacyJWT(legacy)
	if err != nil {
		t.Fatalf("ExchangeLegacyJWT() error = %v", err)
	}
	if session.UserID != 90525 || session.OrgID != 29 || session.Nickname != "用户A" {
		t.Fatalf("session = %+v", session)
	}
	if !expiresAt.Equal(now.Add(24 * time.Hour)) {
		t.Fatalf("expiresAt = %v, want %v", expiresAt, now.Add(24*time.Hour))
	}

	parsed, err := mgr.ParseSessionJWT(token)
	if err != nil {
		t.Fatalf("ParseSessionJWT() error = %v", err)
	}
	if parsed.OrgID != 29 || parsed.DepartmentID != 3 || !parsed.IsOpenEdu {
		t.Fatalf("parsed = %+v", parsed)
	}

	actor := parsed.Actor()
	if actor.ID != 90525 || actor.Username != "用户A" || actor.Role != "super" {
		t.Fatalf("actor = %+v", actor)
	}
}

func TestAdminSessionManagerRejectsExpiredOrInvalidLegacyJWT(t *testing.T) {
	now := time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC)
	mgr := NewAdminSessionManager(config.SessionConfig{
		LegacySecret: "legacy-secret",
		Secret:       "session-secret",
		Issuer:       "go-auth-demo",
		TTLHours:     24,
	})
	mgr.now = func() time.Time { return now }

	expired := signedLegacyJWT(t, "legacy-secret", map[string]any{
		"id":       "90525",
		"orgid":    "29",
		"orgname":  "一县一端测试机构",
		"nickname": "用户A",
		"priv":     "super",
		"exp":      now.Add(-time.Minute).Unix(),
	})
	if _, _, _, err := mgr.ExchangeLegacyJWT(expired); err == nil {
		t.Fatal("ExchangeLegacyJWT() error = nil, want unauthorized")
	}

	wrongSignature := signedLegacyJWT(t, "other-secret", map[string]any{
		"id":       "90525",
		"orgid":    "29",
		"orgname":  "一县一端测试机构",
		"nickname": "用户A",
		"priv":     "super",
		"exp":      now.Add(time.Hour).Unix(),
	})
	if _, _, _, err := mgr.ExchangeLegacyJWT(wrongSignature); err == nil {
		t.Fatal("ExchangeLegacyJWT() error = nil, want unauthorized")
	}
}

func TestAdminSessionManagerAlwaysUsesFixedCookieName(t *testing.T) {
	mgr := NewAdminSessionManager(config.SessionConfig{
		LegacySecret: "legacy-secret",
		Secret:       "session-secret",
		Issuer:       "go-auth-demo",
		TTLHours:     24,
	})

	if got := mgr.CookieName(); got != "as_admin_session" {
		t.Fatalf("CookieName() = %q, want %q", got, "as_admin_session")
	}
}

func TestContextWithAdminSessionStoresSession(t *testing.T) {
	session := AdminSession{
		UserID:   90525,
		OrgID:    29,
		Nickname: "用户A",
		Priv:     "super",
	}
	ctx := ContextWithAdminSession(t.Context(), session)
	got, ok := AdminSessionFromContext(ctx)
	if !ok {
		t.Fatal("AdminSessionFromContext() ok = false")
	}
	if got.OrgID != 29 || got.Nickname != "用户A" {
		t.Fatalf("session = %+v", got)
	}
}

func signedLegacyJWT(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return signed
}

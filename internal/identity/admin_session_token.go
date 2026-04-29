package identity

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dovetaill/article-sentinel/pkg/config"
	jwt "github.com/golang-jwt/jwt/v5"
)

type AdminSessionManager struct {
	legacySecret []byte
	secret       []byte
	issuer       string
	ttl          time.Duration
	secureCookie bool
	now          func() time.Time
	err          error
}

const adminSessionCookieName = "as_admin_session"

type adminSessionClaims struct {
	UserID       uint64 `json:"id"`
	OrgID        uint64 `json:"orgid"`
	OrgName      string `json:"orgname,omitempty"`
	Platform     string `json:"platform,omitempty"`
	Priv         string `json:"priv,omitempty"`
	RoleID       string `json:"roleid,omitempty"`
	Nickname     string `json:"nickname,omitempty"`
	Avatar       string `json:"avatar,omitempty"`
	DepartmentID uint64 `json:"departmentid,omitempty"`
	IsOpenEdu    bool   `json:"is_open_edu,omitempty"`
	Status       string `json:"status,omitempty"`
	jwt.RegisteredClaims
}

func NewAdminSessionManager(cfg config.SessionConfig) *AdminSessionManager {
	ttl := time.Duration(cfg.TTLHours) * time.Hour
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	issuer := strings.TrimSpace(cfg.Issuer)
	if issuer == "" {
		issuer = "article-sentinel-admin"
	}

	manager := &AdminSessionManager{
		legacySecret: []byte(cfg.LegacySecret),
		secret:       []byte(cfg.Secret),
		issuer:       issuer,
		ttl:          ttl,
		secureCookie: cfg.SecureCookie,
		now:          time.Now,
	}

	if strings.TrimSpace(cfg.LegacySecret) == "" || strings.TrimSpace(cfg.Secret) == "" {
		manager.err = fmt.Errorf("admin session secrets are required")
	}

	return manager
}

func (m *AdminSessionManager) CookieName() string {
	return adminSessionCookieName
}

func (m *AdminSessionManager) SecureCookie() bool {
	return m != nil && m.secureCookie
}

func (m *AdminSessionManager) TTL() time.Duration {
	if m == nil {
		return 24 * time.Hour
	}
	return m.ttl
}

func (m *AdminSessionManager) Issuer() string {
	if m == nil {
		return "article-sentinel-admin"
	}
	return m.issuer
}

func (m *AdminSessionManager) ExchangeLegacyJWT(tokenString string) (string, AdminSession, time.Time, error) {
	if m == nil {
		return "", AdminSession{}, time.Time{}, ErrUnauthorized
	}
	if m.err != nil {
		return "", AdminSession{}, time.Time{}, m.err
	}

	legacyClaims := jwt.MapClaims{}
	parser := jwt.NewParser(jwt.WithTimeFunc(m.now))
	parsed, err := parser.ParseWithClaims(tokenString, legacyClaims, func(token *jwt.Token) (any, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
		}
		return m.legacySecret, nil
	})
	if err != nil || !parsed.Valid {
		return "", AdminSession{}, time.Time{}, ErrUnauthorized
	}

	session, err := mapLegacyAdminSession(legacyClaims)
	if err != nil {
		return "", AdminSession{}, time.Time{}, err
	}

	signed, expiresAt, err := m.SignSessionJWT(session)
	if err != nil {
		return "", AdminSession{}, time.Time{}, err
	}
	return signed, session, expiresAt, nil
}

func (m *AdminSessionManager) SignSessionJWT(session AdminSession) (string, time.Time, error) {
	if m == nil || m.err != nil {
		if m != nil && m.err != nil {
			return "", time.Time{}, m.err
		}
		return "", time.Time{}, ErrUnauthorized
	}
	if session.UserID == 0 || session.OrgID == 0 {
		return "", time.Time{}, ErrUnauthorized
	}

	now := m.now()
	expiresAt := now.Add(m.ttl)
	claims := adminSessionClaims{
		UserID:       session.UserID,
		OrgID:        session.OrgID,
		OrgName:      strings.TrimSpace(session.OrgName),
		Platform:     strings.TrimSpace(session.Platform),
		Priv:         strings.TrimSpace(session.Priv),
		RoleID:       strings.TrimSpace(session.RoleID),
		Nickname:     strings.TrimSpace(session.Nickname),
		Avatar:       strings.TrimSpace(session.Avatar),
		DepartmentID: session.DepartmentID,
		IsOpenEdu:    session.IsOpenEdu,
		Status:       strings.TrimSpace(session.Status),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(session.UserID, 10),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (m *AdminSessionManager) ParseSessionJWT(tokenString string) (AdminSession, error) {
	if m == nil || m.err != nil {
		if m != nil && m.err != nil {
			return AdminSession{}, m.err
		}
		return AdminSession{}, ErrUnauthorized
	}

	parser := jwt.NewParser(jwt.WithTimeFunc(m.now))
	parsed, err := parser.ParseWithClaims(tokenString, &adminSessionClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
		}
		return m.secret, nil
	})
	if err != nil {
		return AdminSession{}, ErrUnauthorized
	}

	claims, ok := parsed.Claims.(*adminSessionClaims)
	if !ok || !parsed.Valid {
		return AdminSession{}, ErrUnauthorized
	}
	session := claims.Session()
	if session.UserID == 0 || session.OrgID == 0 {
		return AdminSession{}, ErrUnauthorized
	}
	return session, nil
}

func (c *adminSessionClaims) Session() AdminSession {
	if c == nil {
		return AdminSession{}
	}
	return AdminSession{
		UserID:       c.UserID,
		OrgID:        c.OrgID,
		OrgName:      c.OrgName,
		Platform:     c.Platform,
		Priv:         c.Priv,
		RoleID:       c.RoleID,
		Nickname:     c.Nickname,
		Avatar:       c.Avatar,
		DepartmentID: c.DepartmentID,
		IsOpenEdu:    c.IsOpenEdu,
		Status:       c.Status,
	}
}

func mapLegacyAdminSession(claims jwt.MapClaims) (AdminSession, error) {
	userID, err := legacyUint64Claim(claims, "id")
	if err != nil || userID == 0 {
		return AdminSession{}, ErrUnauthorized
	}
	orgID, err := legacyUint64Claim(claims, "orgid")
	if err != nil || orgID == 0 {
		return AdminSession{}, ErrUnauthorized
	}

	session := AdminSession{
		UserID:       userID,
		OrgID:        orgID,
		OrgName:      legacyStringClaim(claims, "orgname"),
		Platform:     legacyStringClaim(claims, "platform"),
		Priv:         legacyStringClaim(claims, "priv"),
		RoleID:       legacyStringClaim(claims, "roleid"),
		Nickname:     legacyStringClaim(claims, "nickname"),
		Avatar:       legacyStringClaim(claims, "avatar"),
		Status:       legacyStringClaim(claims, "status"),
		DepartmentID: legacyOptionalUint64Claim(claims, "departmentid"),
		IsOpenEdu:    legacyOptionalBoolClaim(claims, "is_open_edu"),
	}
	return session, nil
}

func legacyStringClaim(claims jwt.MapClaims, key string) string {
	value, ok := claims[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func legacyUint64Claim(claims jwt.MapClaims, key string) (uint64, error) {
	value, ok := claims[key]
	if !ok {
		return 0, errors.New("missing claim")
	}
	return coerceUint64(value)
}

func legacyOptionalUint64Claim(claims jwt.MapClaims, key string) uint64 {
	value, ok := claims[key]
	if !ok {
		return 0
	}
	result, err := coerceUint64(value)
	if err != nil {
		return 0
	}
	return result
}

func legacyOptionalBoolClaim(claims jwt.MapClaims, key string) bool {
	value, ok := claims[key]
	if !ok {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(typed))
		return err == nil && parsed
	case float64:
		return typed != 0
	default:
		return false
	}
}

func coerceUint64(value any) (uint64, error) {
	switch typed := value.(type) {
	case uint64:
		return typed, nil
	case uint32:
		return uint64(typed), nil
	case uint:
		return uint64(typed), nil
	case int:
		if typed < 0 {
			return 0, errors.New("negative value")
		}
		return uint64(typed), nil
	case int64:
		if typed < 0 {
			return 0, errors.New("negative value")
		}
		return uint64(typed), nil
	case float64:
		if typed < 0 {
			return 0, errors.New("negative value")
		}
		return uint64(typed), nil
	case string:
		return strconv.ParseUint(strings.TrimSpace(typed), 10, 64)
	default:
		return 0, fmt.Errorf("unsupported uint64 claim type %T", value)
	}
}

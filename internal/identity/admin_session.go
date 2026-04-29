package identity

import (
	"context"
	"strconv"
	"strings"
)

type adminSessionContextKey struct{}

// AdminSession 表示管理台第三方跳转登录后的会话信息。
type AdminSession struct {
	UserID       uint64 `json:"id"`
	OrgID        uint64 `json:"orgid"`
	OrgName      string `json:"orgname"`
	Platform     string `json:"platform,omitempty"`
	Priv         string `json:"priv,omitempty"`
	RoleID       string `json:"roleid,omitempty"`
	Nickname     string `json:"nickname,omitempty"`
	Avatar       string `json:"avatar,omitempty"`
	DepartmentID uint64 `json:"departmentid,omitempty"`
	IsOpenEdu    bool   `json:"is_open_edu,omitempty"`
	Status       string `json:"status,omitempty"`
}

func ContextWithAdminSession(ctx context.Context, session AdminSession) context.Context {
	return context.WithValue(ctx, adminSessionContextKey{}, session)
}

func AdminSessionFromContext(ctx context.Context) (AdminSession, bool) {
	session, ok := ctx.Value(adminSessionContextKey{}).(AdminSession)
	return session, ok
}

func (s AdminSession) Actor() Actor {
	username := strings.TrimSpace(s.Nickname)
	if username == "" {
		username = strconv.FormatUint(s.UserID, 10)
	}

	status := strings.TrimSpace(s.Status)
	if status == "" {
		status = "active"
	}

	return NewActor(uint(s.UserID), username, strings.TrimSpace(s.Priv), status)
}

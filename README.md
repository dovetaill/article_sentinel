# go-auth-demo

`go-auth-demo` is a small Go HTTP starter for apps that need to exchange a legacy HS256 JWT callback for an HTTP-only session cookie. It keeps the runtime intentionally compact: one server command, typed YAML config, structured logs, health checks, a protected demo endpoint, and protected API documentation.

## Purpose

Use this repository as a reusable backend seed when you want:

- A minimal Go server that starts with only a config file and the Go toolchain.
- A legacy JWT bridge at `/auth/login?jwt=...` that issues the fixed `as_admin_session` cookie.
- A simple session model that maps authenticated users into an actor/principal context.
- Huma-powered health, readiness, OpenAPI, schema, and docs routes.
- A clean place to add product modules without carrying app-specific history.

## Features

- `cmd/server` entrypoint with graceful shutdown.
- Strongly typed config loaded from YAML plus environment overrides.
- Structured logging with stdout/file output options and rotation settings.
- Request id, recover, timeout, access log, session parsing, and docs guard layers.
- Public `/healthz` and `/readyz` probes.
- Runtime auth routes for login, current session, and logout.
- Protected `/api/v1/demo/me` example route that returns the current actor.
- Protected documentation and schema browser endpoints.

## Directory Structure

```text
go-auth-demo/
├── cmd/server/                 # server entrypoint
├── configs/                    # example and local YAML config
├── internal/
│   ├── api/                    # route registration, handlers, response envelope
│   ├── app/                    # runtime bootstrap and shutdown helpers
│   ├── identity/               # session, actor, and principal types
│   └── request pipeline helpers
├── pkg/config/                 # config schema and loader
├── pkg/logger/                 # slog setup and rotation support
└── docs/plans/                 # retained init design and implementation notes
```

## Run Locally

Start the server with the checked-in example config:

```bash
go run ./cmd/server -config configs/config.example.yaml
```

The example config listens on `0.0.0.0:8080`, sets `secure_cookie: false` for local HTTP testing, and redirects successful logins to `/docs`.

Quick public checks:

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/readyz
```

## Configuration

The server reads YAML from `-config` or from `CONFIG_PATH` when the flag is omitted. `configs/config.example.yaml` is the canonical starter config; `configs/config.local.yaml` is a local copy you can adjust.

Important sections:

| Section | Keys | Notes |
| --- | --- | --- |
| `app` | `name`, `env`, `host`, `port` | app identity and listen address |
| `http` | `request_timeout_seconds`, `read_timeout_seconds`, `write_timeout_seconds`, `idle_timeout_seconds` | server and handler timeouts |
| `auth.session` | `legacy_secret`, `secret`, `issuer`, `ttl_hours`, `secure_cookie`, `login_url`, `redirect_url` | legacy JWT validation and session cookie signing |
| `docs` | `enabled`, `openapi_path`, `ui_path` | Huma docs and OpenAPI paths |
| `log` | `level`, `format`, `output`, rotation keys | structured logging output |

Useful environment overrides include `APP_PORT`, `AUTH_SESSION_LEGACY_SECRET`, `AUTH_SESSION_SECRET`, `AUTH_SESSION_SECURE_COOKIE`, `AUTH_SESSION_LOGIN_URL`, `AUTH_SESSION_REDIRECT_URL`, `DOCS_ENABLED`, and `LOG_LEVEL`.

For real deployments, replace both sample secrets and prefer environment injection over committed config values.

## Auth Flow

1. A trusted upstream app signs an HS256 legacy JWT with `auth.session.legacy_secret`.
2. The user is sent to `GET /auth/login?jwt=<legacy-jwt>`.
3. The server validates the token and requires `id` and `orgid` claims.
4. Optional claims such as `orgname`, `platform`, `priv`, `roleid`, `nickname`, `avatar`, `departmentid`, `is_open_edu`, and `status` are copied into the session.
5. The server signs a new session JWT with `auth.session.secret` and sets `as_admin_session` as `HttpOnly`, `SameSite=Lax`, and `Secure` according to config.
6. The browser is redirected to `auth.session.redirect_url`.
7. Protected endpoints read the cookie on each request and return `401` when the session is missing or invalid.

## Sign A Local Legacy JWT

Run this from any shell to generate a short-lived HS256 token using only the Go standard library:

```bash
cat >/tmp/sign-legacy-jwt.go <<'EOF_GO'
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	secret := os.Getenv("LEGACY_SECRET")
	if secret == "" {
		secret = "change-me-legacy-secret"
	}

	now := time.Now()
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	claims := map[string]any{
		"id":       1001,
		"orgid":    2001,
		"orgname":  "Local Org",
		"nickname": "Local Admin",
		"priv":     "admin",
		"status":   "active",
		"iat":      now.Unix(),
		"nbf":      now.Unix(),
		"exp":      now.Add(time.Hour).Unix(),
	}

	encoded := strings.Join([]string{mustJSON(header), mustJSON(claims)}, ".")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encoded))
	fmt.Println(encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)))
}

func mustJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}
EOF_GO

LEGACY_JWT=$(LEGACY_SECRET=change-me-legacy-secret go run /tmp/sign-legacy-jwt.go)
echo "$LEGACY_JWT"
```

Exchange it for a session cookie and call protected endpoints:

```bash
curl -i -c /tmp/go-auth-demo.cookies "http://127.0.0.1:8080/auth/login?jwt=${LEGACY_JWT}"
curl -b /tmp/go-auth-demo.cookies http://127.0.0.1:8080/api/v1/auth/session
curl -b /tmp/go-auth-demo.cookies http://127.0.0.1:8080/api/v1/demo/me
curl -b /tmp/go-auth-demo.cookies http://127.0.0.1:8080/docs
```

## API

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| GET | `/healthz` | No | process health |
| GET | `/readyz` | No | app readiness |
| GET | `/auth/login?jwt=...` | No | exchange legacy JWT for session cookie |
| GET | `/api/v1/auth/session` | Cookie | current session |
| POST | `/api/v1/auth/logout` | No | clear cookie |
| GET | `/api/v1/demo/me` | Cookie | current actor demo |
| GET | `/docs` | Cookie | API docs |
| GET | `/openapi.json` | Cookie | OpenAPI 3.1 JSON |
| GET | `/openapi.yaml` | Cookie | OpenAPI 3.1 YAML |
| GET | `/openapi-3.0.json` | Cookie | OpenAPI 3.0 JSON |
| GET | `/openapi-3.0.yaml` | Cookie | OpenAPI 3.0 YAML |
| GET | `/schemas/{schema}` | Cookie | JSON schema browser |

All JSON responses use this envelope shape:

```json
{
  "code": 0,
  "message": "session",
  "data": {}
}
```

Failures use the same shape with an HTTP status code in `code`, for example `401` and `"unauthorized"`.

## Protected Docs Endpoints

Documentation is enabled by `docs.enabled`. The docs UI and schema routes are intentionally private: call `/auth/login` first, keep the `as_admin_session` cookie, then visit `/docs`, `/openapi.json`, `/openapi.yaml`, `/openapi-3.0.json`, `/openapi-3.0.yaml`, or `/schemas/{schema}`.

If you change `docs.openapi_path` or `docs.ui_path`, update any client bookmarks and tests that reference the old paths.

## Extend The Starter

1. Add request/response types and handlers under `internal/api/handlers`.
2. Register Huma operations in `internal/api/register/router.go` so they appear in OpenAPI.
3. Keep public endpoints explicit; require the session actor for private endpoints.
4. Add config fields in `pkg/config` only when the server truly needs them.
5. Add focused tests next to the new package, then run `go test ./...` before committing.
6. Update this README when you add new public routes or required config.

## Test

Run the full test suite:

```bash
go test ./...
```

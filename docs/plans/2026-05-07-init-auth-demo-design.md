# Init Auth Demo Design

## Goal

`go-auth-demo` is a reusable Go HTTP starter that demonstrates a legacy HS256 JWT callback exchanged for an HTTP-only admin session cookie. The branch should stay small, runnable with the Go toolchain alone, and free of product-specific operating notes.

## Runtime Shape

The demo keeps a single server entrypoint:

- `cmd/server` loads config, creates the logger, builds the router, and starts `net/http`.
- `configs/config.example.yaml` provides local defaults that run without extra services.
- `pkg/config` owns the typed config schema and environment overrides.
- `pkg/logger` owns structured log setup and rotation settings.
- `internal/api` owns route registration, handlers, and the shared response envelope.
- `internal/identity` owns sessions, actors, and principals.

## HTTP Surface

Public endpoints:

- `GET /healthz`
- `GET /readyz`
- `GET /auth/login?jwt=...`
- `POST /api/v1/auth/logout`

Cookie-protected endpoints:

- `GET /api/v1/auth/session`
- `GET /api/v1/demo/me`
- `GET /docs`
- `GET /openapi.json`
- `GET /openapi.yaml`
- `GET /openapi-3.0.json`
- `GET /openapi-3.0.yaml`
- `GET /schemas/{schema}`

## Auth Flow

1. A trusted caller signs an HS256 legacy JWT with `auth.session.legacy_secret`.
2. The browser opens `/auth/login?jwt=<token>`.
3. The server validates the token and requires `id` and `orgid`.
4. Optional profile claims are copied into an admin session.
5. The server signs a session JWT with `auth.session.secret`.
6. The session JWT is stored in the fixed `as_admin_session` cookie.
7. Private routes read the cookie and return `401` for missing or invalid sessions.

## Config

The starter config includes only generic runtime sections:

- `app`: name, env, host, port
- `http`: server and request timeouts
- `auth.session`: legacy secret, session secret, issuer, TTL, cookie security, login URL, redirect URL
- `docs`: docs switch and OpenAPI/UI paths
- `log`: log level, format, output, and rotation settings

## Extension Rules

- Add new Huma operations through `internal/api/register/router.go`.
- Keep public routes explicit and make private routes session-aware.
- Add config only when startup or request handling needs it.
- Keep examples generic so this branch remains a clean seed project.
- Verify changes with `go test ./...`.

# Init Auth Demo Design

## Goal

Create an `init` branch that turns `article-sentinel` into a reusable Go HTTP backend demo focused on JWT-based admin session authentication. The demo should start without frontend assets, business modules, queues, schedulers, or database requirements.

## Scope

Keep a runnable HTTP service with:

- configuration loading
- structured logging
- request id, recover, timeout, access log middleware
- JWT legacy-token exchange into an HTTP-only session cookie
- session parsing middleware
- current-session and logout APIs
- health/readiness endpoints
- protected API documentation endpoints
- tests for the retained auth and HTTP behavior

Remove:

- all React/Vite frontend code under `web/`
- article inspection business logic
- starter `post` business demo module
- worker, scheduler, queue, outbox, and migrations tied to business processing
- business DDL, seed scripts, and article-inspection documentation

## Architecture

The resulting backend keeps one executable entrypoint:

- `cmd/server`: loads config, builds runtime resources, registers routes, and starts `net/http`.

The HTTP surface is intentionally small:

- `GET /healthz`: process health check
- `GET /readyz`: app readiness check that does not require a database
- `GET /auth/login?jwt=<legacy-jwt>`: validates an HS256 legacy JWT and sets `as_admin_session`
- `GET /api/v1/auth/session`: returns the current session from the cookie
- `POST /api/v1/auth/logout`: clears the session cookie
- optional `GET /api/v1/demo/me`: protected example endpoint that returns the current actor/session

The router should no longer import business modules or queue dispatchers. It should only wire public utility routes and authentication routes, then wrap the mux with the retained middleware chain. Documentation endpoints must be protected by session authentication, including `/docs`, `/openapi.json`, `/openapi.yaml`, `/openapi-3.0.json`, `/openapi-3.0.yaml`, and `/schemas/*`.

## Configuration

Keep only generic configuration sections:

- `app`: name, env, host, port
- `http`: request/read/write/idle timeouts
- `auth.session`: legacy secret, session secret, issuer, TTL, secure cookie, login URL, redirect URL
- `docs`: OpenAPI/docs switches if the Huma health routes remain documented
- `log`: level, format, output, rotation settings

Drop database, redis, queue, scheduler, and outbox config from the demo unless a retained package still needs it during implementation. The default config should allow:

```bash
go run ./cmd/server -config configs/config.example.yaml
```

without Docker, MySQL, or Redis.

## Data Flow

Login flow:

1. A third-party system redirects to `/auth/login?jwt=<legacy-jwt>`.
2. The handler validates the legacy JWT using `auth.session.legacy_secret`.
3. Required claims (`id`, `orgid`) are mapped into `identity.AdminSession`.
4. The server signs a new session JWT using `auth.session.secret`.
5. The server stores it in an HTTP-only `as_admin_session` cookie and redirects to `auth.session.redirect_url`.

Session flow:

1. `SessionContext` middleware reads `as_admin_session`.
2. It validates the session JWT and stores `AdminSession`, `Actor`, and `Principal` in request context.
3. Authenticated handlers read the context and return session or actor data.
4. Invalid cookies are cleared and treated as anonymous requests.

## Error Handling

- Invalid or missing login JWT clears the session cookie and redirects to `auth.session.login_url`.
- Anonymous requests to documentation endpoints return `401` JSON instead of exposing OpenAPI schemas publicly.
- Missing or invalid session cookie returns `401` on protected/session endpoints.
- Unsupported HTTP methods return `405` JSON envelopes.
- Middleware keeps request metadata and source IP available for downstream code.
- Startup fails fast only for invalid config or logger setup, not for missing database or Redis.

## Testing

Retain or rewrite tests for:

- legacy JWT exchange and session JWT parsing
- fixed session cookie name
- session middleware context population and bad-cookie cleanup
- auth login/session/logout handlers
- router route presence and removed business route absence
- docs/OpenAPI/schema endpoints requiring a valid session cookie
- config loading with the reduced schema
- server-independent health/readiness behavior

Run at least:

```bash
go test ./...
```

## Documentation

Rewrite `README.md` as a generic backend starter guide covering:

- project purpose
- directory structure
- configuration
- local run command
- authentication flow
- API examples
- how to build new modules on top of the template


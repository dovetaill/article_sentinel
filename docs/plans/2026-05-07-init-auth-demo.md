# Init Auth Demo Implementation Notes

This retained plan note records the intended final shape of the `init` branch for `go-auth-demo`.

## Completed Scope

- Runtime reduced to one server command.
- Config reduced to app, HTTP, auth session, docs, and log sections.
- Module path renamed to `github.com/dovetaill/go-auth-demo`.
- Auth bridge added at `/auth/login?jwt=...`.
- Session cookie fixed to `as_admin_session`.
- Current-session and logout routes added.
- Protected demo route added at `/api/v1/demo/me`.
- Health, readiness, OpenAPI, docs UI, and schema routes retained.
- Docs and schema routes require a valid session cookie.
- Tests cover config shape, auth token exchange, session parsing, route protection, and health/readiness behavior.

## Task 6 Documentation Scope

- `README.md` is the canonical starter guide.
- Old product-specific docs are removed from `docs/`.
- Only the 2026-05-07 init design and implementation notes remain under `docs/plans/`.
- The retained docs avoid references to the source product so the branch can be reused as a template.

## Verification

Run before completing the branch:

```bash
go test ./...
```

Run the Task 6 stale-reference check after docs cleanup.

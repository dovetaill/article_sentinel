# Main Starter / Showcase Split Design

**Date:** 2026-04-03

## Goal

Reposition `main` as a minimal, production-leaning starter while preserving the current multi-surface business implementation as a dedicated showcase branch. The starter should remain runnable out of the box, demonstrate article-sentinel's preferred layering and OpenAPI wiring, and avoid forcing consumers to adopt the current `admin/member/article/category/engagement` domain model.

## Current Problems

1. `README.md` reads like a productized business backend instead of a reusable starter.
2. `internal/api/register/router.go` directly assembles concrete business modules, so the core entrypoint is coupled to one opinionated app shape.
3. `internal/identity` currently depends on `internal/modules/auth`, which means the identity layer is not actually reusable.
4. `internal/app/bootstrap/server.go` assumes business schema migration and seed-admin behavior by default.
5. The default branch tells a richer story than many adopters want before they have even replaced the demo domain.

## Product Positioning Decision

### `main`

`main` becomes the canonical starter branch:

- keeps runtime entrypoints: `server`, `worker`, `scheduler`, `migrate`
- keeps infrastructure: config, logger, database, Redis bootstrap, request middleware, response envelope, health/readiness, OpenAPI docs
- keeps a thin authentication/identity foundation, but only in reusable form
- keeps a single demo CRUD module to show how handlers/services/repositories fit together
- stops presenting multi-role, multi-surface business architecture as the default project shape

### `showcase/multisurface`

The current codebase shape becomes the practical showcase branch:

- keeps `public`, `member auth`, `member self`, `admin` surfaces
- keeps `auth`, `user`, `member`, `category`, `article`, `engagement`
- keeps seed-admin behavior and richer onboarding docs
- serves as the reference app for teams that want the full opinionated structure

## Starter Scope for `main`

The minimal branch should still feel real. It should include:

- `GET /healthz`
- `GET /readyz`
- OpenAPI + docs UI
- unified JSON envelope
- a single `post`-style CRUD example module
- optional bearer auth foundation that consumers can wire into their own user model later

Recommended example surface:

- `GET /api/v1/posts`
- `GET /api/v1/posts/{id}`
- `POST /api/v1/posts`
- `PATCH /api/v1/posts/{id}`
- `DELETE /api/v1/posts/{id}`

This keeps the branch understandable while still demonstrating request models, response models, service/repository layering, pagination, and route registration.

## Identity / Auth Boundary

The starter should keep authentication primitives, but only as reusable infrastructure.

### Keep in `main`

- bearer token parsing middleware
- token manager implementation
- request context actor/principal storage
- a generic `RequireAuthenticated()` middleware
- optional generic role/claim guard only if it does not hardcode `admin/member`

### Remove from `main`

- `admin` and `member` role worldview
- `RequireAdmin()` and `RequireMember()` as starter defaults
- identity helpers that depend on `internal/modules/auth`
- default admin seeding
- member registration/login/self-service demo APIs

### Design Rule

Core identity code must not import demo business modules. Business modules may depend on identity primitives, never the reverse.

## Router and Bootstrap Refactor

### Router responsibilities in `main`

`internal/api/register/router.go` should only:

- create the Huma API instance
- register health and readiness endpoints
- attach shared middleware
- register starter demo routes through a thin module hook

It should no longer know about all business modules.

### Bootstrap responsibilities in `main`

`internal/app/bootstrap/server.go` should:

- bootstrap shared runtime resources
- optionally auto-migrate starter demo models
- avoid business-specific seeding behavior by default

If starter demo data needs setup, it should live beside the demo module rather than inside the global server bootstrap path.

## Documentation Split

### `main` README

The `main` README should emphasize:

- what the starter gives you
- how to boot it locally
- where the demo module lives
- how to replace the demo module with your own domain
- where to find the richer showcase branch

### Showcase README

The showcase branch README should keep:

- the multi-surface explanation
- admin/member/business capability map
- role rules and example workflows
- richer API inventory

## Migration Strategy

1. Preserve current state in a showcase branch before deleting or simplifying modules on `main`.
2. Extract reusable identity primitives away from `internal/modules/auth`.
3. Introduce a single starter CRUD module.
4. Simplify router/bootstrap around shared infrastructure + starter demo module.
5. Rewrite `README.md` for starter positioning.
6. Add clear cross-links between `main` and the showcase branch.

## Acceptance Criteria

The redesign is successful when:

- a new reader lands on `main` and immediately understands it is a starter
- the default branch can be learned without reading role/surface business rules
- identity primitives are reusable and not coupled to the current auth module
- the repository still demonstrates a full request lifecycle through one demo module
- the current multi-surface app remains available as a first-class showcase branch

# Article Inspect Architecture Refactor Design

**Date:** 2026-04-28

## Goal

Refactor `internal/modules/articleinspect` into a maintainable, Huma-idiomatic module without breaking its current business behavior. The refactor should remove the current `handler.go` bottleneck, improve route and type ownership, centralize module wiring, and make the HTTP contract more accurately reflected in generated OpenAPI.

## Current State

The current module has one dominant structural problem: `internal/modules/articleinspect/handler.go` mixes nearly all transport concerns in one file.

Today that file contains:

- Huma request structs for every resource
- all route registration
- response envelope helpers
- status/error mapping
- path/query parsing helpers

This leads to poor locality and makes feature work awkward because any route change requires navigating a large shared file.

The module is also workflow-heavy rather than CRUD-isolated:

- task creation snapshots keywords and inspection rules
- the worker reads tasks, scans articles, and persists results
- batch actions and lifecycle flows coordinate article state and result disposition
- result detail queries also aggregate logs and field changes

Because of that coupling, a hard split into `handler/service/repository/dto` packages or per-resource packages would add package churn faster than it would reduce complexity.

## Problems To Solve

### 1. Route and transport sprawl

`handler.go` is the main navigation bottleneck. Request structs, route registration, and transport helpers are not grouped by feature.

### 2. Module assembly leakage

`internal/api/register/router.go` currently assembles every repository and service manually for this module. Tests repeat that same setup. The module should own its own wiring entrypoint.

### 3. Contract drift with Huma/OpenAPI

The current route code manually parses many IDs, booleans, and timestamps from strings instead of letting Huma type them directly. This weakens schema accuracy.

The current shared response output also documents only `200` success by default even when some routes actually return `201`.

### 4. Type ownership drift

Transport request structs, service inputs, and repository query types are currently mixed inconsistently across `handler.go`, `dto_*.go`, and repository files. Ownership is unclear.

### 5. Testing does not lock enough contract detail

The current OpenAPI test mainly checks path presence. It does not protect operation IDs, response status documentation, or parameter schema types.

## Decision

Use a two-layer refactor inside a single `articleinspect` package.

### Keep a single package

The module remains one Go package for now.

This is intentional because:

- the module is still one bounded workflow
- package-private helpers are heavily reused
- introducing subpackages now would require exporting or duplicating internal logic
- current tests and bootstrap wiring would churn more than necessary

### Split by feature files, not by package layers

Refactor transport code into route files grouped by resource/capability.

Target file layout:

- `internal/modules/articleinspect/module.go`
- `internal/modules/articleinspect/routes.go`
- `internal/modules/articleinspect/routes_common.go`
- `internal/modules/articleinspect/category_routes.go`
- `internal/modules/articleinspect/keyword_routes.go`
- `internal/modules/articleinspect/task_routes.go`
- `internal/modules/articleinspect/result_routes.go`
- `internal/modules/articleinspect/action_routes.go`
- `internal/modules/articleinspect/lifecycle_routes.go`
- `internal/modules/articleinspect/log_routes.go`
- `internal/modules/articleinspect/article_routes.go`

Existing services and repositories stay in the same package, but type files will be cleaned up over time.

### Use Huma groups for prefix ownership

Inside `RegisterRoutes`, create a base group for `/api/v1/article-inspect` and register resource-relative paths underneath it.

That keeps `huma.Register(...)` explicit while removing repeated hard-coded prefixes from every route.

Resource-level groups such as `/categories`, `/keywords`, `/tasks`, `/results`, `/actions`, `/articles`, and `/logs` are acceptable where they improve locality or shared metadata.

### Keep Huma explicit

Do not build a custom route DSL or registrar framework.

Keep explicit `huma.Register(...)` calls and explicit `OperationID`s. Thin helper constructors are acceptable only for repeated metadata or repeated output shapes.

### Correct the Huma contract where low-risk

Within this refactor, correct the most important contract drift:

- use typed Huma path/query fields where practical
- stop manually parsing IDs and booleans when Huma can do it
- document create routes with `201` success outputs instead of only `200`
- keep current explicit `OperationID`s stable

A full project-wide change to Huma-native error bodies is out of scope for this refactor. The module will keep the existing envelope runtime shape.

### Centralize module wiring

Add a module-local constructor that builds `Routes` from the database and dispatcher dependencies. `internal/api/register/router.go` should ask the module for its assembled routes instead of manually constructing every service.

A similar helper should be reusable from tests.

## Scope

### In Scope

- split `handler.go` into feature route files
- introduce a module-level assembly entrypoint
- move transport request structs next to their route registration code
- introduce a base Huma group for the module
- keep or improve existing runtime behavior and status codes
- improve OpenAPI success status and typed parameter documentation
- strengthen route registration tests and OpenAPI contract assertions
- optionally split `model.go` into more navigable files if needed during the refactor

### Out of Scope

- converting the whole project to Huma-native `problem+json` runtime errors
- rewriting unrelated modules like `post`
- introducing `handler/service/repository/dto` subpackages
- introducing per-resource subpackages like `keywords/` or `tasks/`
- deep business-logic redesign unless needed to preserve route behavior after the transport split

## Architecture Shape

### Module entrypoint

`module.go` should define the public assembly helpers for this module.

Planned responsibilities:

- public `Routes` dependency struct remains the route registration input
- add a constructor such as `NewRoutes(db, dispatcher) Routes`
- optionally add a shared test helper constructor if it reduces repeated test setup

### Route registration

`routes.go` should:

- define `RegisterRoutes`
- create the module base Huma group
- delegate to feature registrars

Each `*_routes.go` should:

- define only the request structs needed for that feature
- define only the feature registrar and its handler closures
- map transport input to service input

`routes_common.go` should own:

- shared success outputs such as `okEnvelopeOutput` and `createdEnvelopeOutput`
- shared failure helpers
- status-from-error mapping
- only the small parsing helpers that remain necessary after Huma typing cleanup

### Type ownership rules

- route-local transport bodies/queries/paths stay in `*_routes.go`
- service/use-case inputs and outputs stay in `dto_*.go` or future `*_types.go`
- repository-only filters stay in repository files until a clearer type split is done
- shared helper structs should only exist if actually reused across route files

### Contract rules

- preserve all existing route paths
- preserve all existing explicit `OperationID`s
- preserve the current JSON envelope runtime format
- update documented success statuses where currently wrong
- improve query/path field typing where low-risk

## Testing Strategy

### Backend route behavior

Keep the existing route behavior tests and update them incrementally as route files are split.

### OpenAPI contract coverage

Extend route registration tests to assert:

- required paths still exist
- key operations keep the same `operationId`
- create endpoints document `201`
- representative parameters such as `id`, `orgid`, `enabled`, and time filters use more accurate schema types

### Regression safety

Run focused `go test ./internal/modules/articleinspect` repeatedly during the refactor so the split remains behavior-preserving.

## Risks

- Switching some request fields from `string` to typed Huma fields may change when validation errors happen.
- Route file extraction can accidentally drift `OperationID` or path composition if group prefixes are applied incorrectly.
- If success output types are changed carelessly, generated OpenAPI may drift unexpectedly.
- Module-local wiring changes must not disturb worker or scheduler dependencies.

## Migration Strategy

### Phase 1: Transport and assembly refactor

- add module-local route assembly
- split route registration into feature files
- add module Huma group
- keep paths and operation IDs stable
- improve success status documentation and typed parameter declarations where practical
- strengthen OpenAPI tests

### Phase 2: Internal cleanup after route split stabilizes

- revisit type ownership in `dto_*` and repository query structs
- optionally split `model.go` for readability
- evaluate whether any workflow abstractions such as `InspectionRunner` or `AuditRecorder` should be extracted in a follow-up

## Why This Design

This design addresses the worst pain immediately without paying the cost of premature package fragmentation.

It respects Go’s package model, uses Huma more idiomatically, improves API contract clarity, and gives the module a real assembly boundary. It is a large refactor, but it is still staged and testable.

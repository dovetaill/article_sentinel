# ArticleInspect Thin Root Design

**Date:** 2026-05-08

## Goal

Finish the `articleinspect` package refactor by collapsing `internal/modules/articleinspect/` into a thin module-entry package and moving all feature ownership into the existing subpackages.

The current codebase already contains subpackages such as `actions/`, `articles/`, `audit/`, `domain/`, `lifecycle/`, `outbox/`, `results/`, `rules/`, `scan/`, `shared/`, `tasks/`, and `worker/`, but the root directory still contains 60 Go files. Many of those root files are compatibility wrappers, alias facades, or duplicated implementations that keep the module in a half-migrated state.

This design removes that half-state. The root package should express only module assembly and module registration. It should not continue to host business logic, route forwarding, duplicated repositories, or compatibility aliases.

## Why This Design Exists

The partial extraction shipped on 2026-05-08 completed the hardest part of the refactor: real subpackages now exist and compile. The remaining problem is structural, not functional.

Today the module mixes three layers in one place:

- real subpackage implementations
- root-level compatibility wrappers that simply forward to subpackages
- root-level concrete implementations that still duplicate subpackage ownership

That creates three maintenance problems:

- developers cannot tell which file is the real implementation
- future changes can accidentally land in the wrong layer
- every merge risks preserving a second copy of logic in the root package

For a new project, this is unnecessary weight. The correct cleanup is to finish the refactor instead of preserving transitional facades.

## Non-Goals

This cleanup must not change runtime behavior.

Out of scope:

- changing HTTP paths or OpenAPI semantics
- changing JSON request or response field names
- changing database table names, GORM model metadata, or migrations
- changing queue payload shape or task lifecycle semantics
- changing worker scan behavior, pagination semantics, or outbox retry behavior
- introducing a heavy DDD or layered-architecture rewrite

## Design Principles

### Thin root package

`internal/modules/articleinspect/` remains the module boundary, but only as a composition boundary.

The root package should keep only:

- `module.go`
- `routes.go`
- an additional root file only if a small module-level contract is truly required

The default target is two root non-test files, not dozens.

### Single ownership per concern

Every concern should have one package that owns it.

Examples:

- models and constants belong to `domain/`
- request scope, operator resolution, pagination, envelopes, and shared parameter types belong to `shared/`
- worker execution belongs to `worker/`
- outbox dispatch and relay belong to `outbox/`
- article list/detail and candidate article access belong to `articles/`

No concern should be implemented in both the root package and a subpackage.

### No compatibility baggage

This project is still early enough that import cleanup is cheaper than preserving transitional wrappers forever.

The refactor should therefore prefer:

- updating imports
- deleting root aliases
- deleting wrapper constructors
- deleting compatibility tests

It should not preserve a root facade just because it existed during an intermediate merge.

## Current Problem Inventory

### Root route wrappers

These files only forward route registration to subpackages and should be deleted:

- `internal/modules/articleinspect/action_routes.go`
- `internal/modules/articleinspect/article_routes.go`
- `internal/modules/articleinspect/category_routes.go`
- `internal/modules/articleinspect/keyword_routes.go`
- `internal/modules/articleinspect/lifecycle_routes.go`
- `internal/modules/articleinspect/log_routes.go`
- `internal/modules/articleinspect/result_routes.go`
- `internal/modules/articleinspect/task_routes.go`

`routes.go` should call subpackage registrars directly.

### Root service and repository wrappers

These files expose subpackage services or repositories through redundant root constructors and should be deleted after import migration:

- `internal/modules/articleinspect/action_service.go`
- `internal/modules/articleinspect/service_articles.go`
- `internal/modules/articleinspect/service_categories.go`
- `internal/modules/articleinspect/service_keywords.go`
- `internal/modules/articleinspect/service_logs.go`
- `internal/modules/articleinspect/service_results.go`
- `internal/modules/articleinspect/service_tasks.go`
- `internal/modules/articleinspect/repository_actions.go`
- `internal/modules/articleinspect/repository_categories.go`
- `internal/modules/articleinspect/repository_keywords.go`
- `internal/modules/articleinspect/repository_results.go`

### Root alias and helper wrappers

These files mostly re-export `domain`, `shared`, `scan`, `lifecycle`, `outbox`, or `worker` content and should be deleted:

- `internal/modules/articleinspect/constants.go`
- `internal/modules/articleinspect/model_article_source.go`
- `internal/modules/articleinspect/model_inspection.go`
- `internal/modules/articleinspect/operator.go`
- `internal/modules/articleinspect/operator_audit.go`
- `internal/modules/articleinspect/paging.go`
- `internal/modules/articleinspect/request_scope.go`
- `internal/modules/articleinspect/transport_envelope.go`
- `internal/modules/articleinspect/transport_errors.go`
- `internal/modules/articleinspect/transport_params.go`
- `internal/modules/articleinspect/diff.go`
- `internal/modules/articleinspect/scanner.go`
- `internal/modules/articleinspect/task_outbox.go`
- `internal/modules/articleinspect/task_outbox_relay.go`
- `internal/modules/articleinspect/task_outbox_settings.go`
- `internal/modules/articleinspect/worker_rules.go`

### Root concrete implementations that still duplicate subpackage ownership

These files should be migrated into the subpackages that already own the concern:

- `internal/modules/articleinspect/repository_articles.go`
- `internal/modules/articleinspect/repository_article_candidates.go`
- `internal/modules/articleinspect/worker.go`
- `internal/modules/articleinspect/routes_common.go`
- `internal/modules/articleinspect/dto_articles.go`
- `internal/modules/articleinspect/dto_categories.go`
- `internal/modules/articleinspect/dto_keywords.go`
- `internal/modules/articleinspect/dto_results.go`
- `internal/modules/articleinspect/dto_tasks.go`
- `internal/modules/articleinspect/util_ids.go`

### Root tests that only exist for compatibility

`internal/modules/articleinspect/subpackage_compat_test.go` exists to prove the root aliases match the extracted subpackages. Once the aliases are removed, the test should be deleted instead of maintained.

## Target Structure

### Root package responsibilities

The root package keeps only module-level composition and route registration.

`internal/modules/articleinspect/module.go`

- assembles subpackage dependencies
- constructs `Routes`
- contains no feature logic

`internal/modules/articleinspect/routes.go`

- defines the root `Routes` struct
- mounts feature routes under `/api/v1/article-inspect`
- imports feature packages directly

### Domain package

`internal/modules/articleinspect/domain/` becomes the only owner of:

- inspection models
- article source models used by the module
- status constants
- rule and action constants

External code that only needs models or constants should import `domain`, not the root package.

### Shared package

`internal/modules/articleinspect/shared/` becomes the only owner of:

- request scope helpers
- org and operator context helpers
- pagination helpers
- transport envelope helpers
- shared numeric and time parameter wrappers
- any truly cross-feature HTTP error helpers

### Feature packages

Each feature package owns its own transport, DTO, repository, service, and routes.

- `actions/`
- `articles/`
- `audit/`
- `lifecycle/`
- `results/`
- `rules/`
- `tasks/`

No feature DTO should remain in the root package after this cleanup.

### Runtime packages

`scan/`, `worker/`, and `outbox/` remain dedicated runtime packages.

Their exported constructors should become the only constructors used outside the module for those responsibilities.

## External Wiring Changes

The root package should no longer act as the public entrypoint for everything.

### HTTP router

`internal/api/register/router.go` should continue to treat `articleinspect` as the module entry package for route registration.

That file should still call:

- `articleinspect.NewRoutes(...)`
- `articleinspect.RegisterRoutes(...)`

But route assembly inside `module.go` should instantiate subpackage services directly.

### Schema registration

`internal/app/bootstrap/schema.go` should stop importing the root `articleinspect` package for models.

It should import `internal/modules/articleinspect/domain` and register domain models directly.

### Queue worker handler

`internal/queue/asynq/handlers.go` should stop importing `articleinspect.NewWorker`.

It should import `internal/modules/articleinspect/worker` and build the executor there directly.

### Scheduler and outbox wiring

`cmd/scheduler/main.go` should stop importing root outbox wrappers.

It should import `internal/modules/articleinspect/outbox` and construct:

- `outbox.NewTaskOutboxRelay(...)`
- `outbox.NewTaskOutboxSettings(...)`

### Dispatcher interface ownership

`internal/queue/asynq/articleinspect_dispatcher.go` should return the canonical outbox dispatcher interface, not a root-level mirror.

The preferred single source of truth is `internal/modules/articleinspect/outbox.TaskDispatcher`.

`tasks/routes.go` should accept that canonical interface instead of maintaining a duplicate local dispatcher interface unless the local interface is strictly needed.

## Repository Consolidation

### Articles repository ownership

The root package currently has:

- `internal/modules/articleinspect/repository_articles.go`
- `internal/modules/articleinspect/repository_article_candidates.go`

while `internal/modules/articleinspect/articles/` already contains article read-side code.

The target is a single `articles.ArticleRepository` that owns both:

- article center list/detail reads
- candidate article loading for scan and worker flows

This avoids keeping two article repositories with overlapping responsibilities in different packages.

### Results and audit separation

The extracted `results/` and `audit/` packages already have their own transport and repository entrypoints. This cleanup should preserve their current behavior but remove any remaining root-level routing or wrapper indirection.

## Test Strategy

### Keep only module-level tests in the root package

Root tests should focus on:

- module wiring
- top-level route registration
- OpenAPI aggregation
- a small number of end-to-end module contract checks, if still valuable

### Move feature and runtime tests to feature packages

Examples:

- `internal/modules/articleinspect/scanner_test.go` -> `internal/modules/articleinspect/scan/`
- `internal/modules/articleinspect/worker_test.go` -> `internal/modules/articleinspect/worker/`
- `internal/modules/articleinspect/lifecycle_test.go` -> `internal/modules/articleinspect/lifecycle/`
- `internal/modules/articleinspect/task_service_test.go` -> `internal/modules/articleinspect/tasks/`
- `internal/modules/articleinspect/task_outbox_test.go` -> `internal/modules/articleinspect/outbox/` or split between `tasks/` and `outbox/`

### Shared test support

The existing `internal/modules/articleinspect/fixtures_test.go` should be replaced by a dedicated helper package such as `internal/modules/articleinspect/testutil/` so feature tests do not need to remain in the root package only to share setup code.

### Add a structural guardrail

Add a root-structure regression test that fails if new root non-test `.go` files are introduced outside the approved whitelist.

That test turns "keep the root thin" into an enforceable rule instead of a style preference.

## Documentation Changes

Update any maintainer-facing docs that still imply the root package owns most internals.

At minimum review:

- `README.md`
- `docs/README.md`
- `docs/maintainer-development-flow.md`

The docs should describe `internal/modules/articleinspect/` as a thin module boundary over feature packages, not a large working package with optional subdirectories.

## Migration Sequence

This cleanup should happen in one coherent refactor branch, but the implementation should still proceed in controlled stages:

1. simplify the root composition layer to instantiate subpackage types directly
2. switch external imports from root wrappers to canonical subpackages
3. consolidate article repository ownership inside `articles/`
4. relocate root DTOs and helpers into their owning packages
5. move tests into feature packages and add `testutil/`
6. add the root-structure guard test
7. delete compatibility wrappers and compatibility tests
8. run full verification and update docs

The important rule is that deletion should happen as a deliberate finish, not as a long-lived "todo later" state.

## Acceptance Criteria

This cleanup is complete when all of the following are true:

- `internal/modules/articleinspect/` root contains only the approved minimal non-test Go files
- no module-external package imports root aliases for worker, outbox, models, or constants
- `internal/app/bootstrap/schema.go` imports `domain` directly
- `internal/queue/asynq/handlers.go` imports `worker` directly
- `cmd/scheduler/main.go` imports `outbox` directly
- article repository ownership is consolidated under `articles/`
- feature and runtime tests live with their owning packages
- compatibility wrappers and `subpackage_compat_test.go` are gone
- HTTP contracts, DB behavior, queue behavior, and OpenAPI routes remain unchanged

## Verification Expectations

Implementation must verify both behavior and structure.

Behavioral verification:

- `go test ./internal/modules/articleinspect/...`
- `go test ./internal/api/register`
- `go test ./internal/queue/asynq`

Structural verification:

- inspect the root package file count after the refactor
- verify the root whitelist test passes
- verify there are no remaining module-external imports relying on deleted root wrappers

## Recommendation

Proceed with the thin-root cleanup now, while the subpackage extraction is still fresh and before more code accumulates on top of the transitional root facades.

This is the lowest-cost moment to make the module elegant, obvious to navigate, and hard to regress.

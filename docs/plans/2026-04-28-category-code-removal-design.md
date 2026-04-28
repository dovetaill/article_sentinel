# Category Code Removal Design

**Date:** 2026-04-28

## Goal

Remove the unused category `code` field from the article inspection system end to end so category management only relies on `id`, `name`, `enabled`, and `sort`.

## Current State

The current implementation exposes category `code` in four places:

- admin UI category list and create/edit modal
- frontend category service request and response types
- backend category models, DTOs, handlers, validation, and repository search
- database schema and seed data for `xt_article_inspect_categories`

Within the current codebase, keyword management and inspection behavior use `category_id`, not category `code`. No active business flow appears to depend on `code`.

## Decision

Use a one-step hard removal.

The field will be removed from:

- frontend UI and service contracts
- backend request/response contracts and validation
- repository query logic
- migration SQL, seed SQL, and runtime schema definition
- automated tests and active API documentation

This is preferred over a staged compatibility approach because the requirement is to delete the field completely, including database structure.

## Scope

### Frontend

- Remove the `分类编码` column from the category table.
- Remove the `分类编码` input from create and edit dialogs.
- Remove `code` from `CategoryRecord` and `CategoryMutationInput`.
- Update category page tests so they only assert the remaining category fields.

### Backend API

- Remove `code` from category request bodies and DTO output.
- Remove `code` validation from category create and update flows.
- Update category list search to filter only by `name`.
- Keep keyword flows unchanged because they already depend on `category_id`.

Unknown JSON fields sent by older callers are expected to be ignored by Go's default decoder, so requests that still include `code` should not fail solely because the field is extra.

### Database

- Remove the `code` column from the `xt_article_inspect_categories` table definition in the base migration SQL.
- Remove the `uk_org_code` unique index.
- Remove `code` from seed inserts.
- Add a follow-up migration for existing databases that drops the index and column if they still exist.

## Migration Strategy

Two schema states must be supported:

1. New environments applying migrations from scratch
2. Existing environments that already have `xt_article_inspect_categories.code`

For new environments:

- update `migrations/20260420_01_article_inspection.sql` so fresh installs create the table without `code`

For existing environments:

- add a new migration file that checks for the `code` column and `uk_org_code` index, then drops them only when present
- keep the SQL idempotent so repeated local `migrate` runs do not fail

Because the bootstrap migration runner replays SQL files directly from the `migrations` directory, the drop migration must be safe to execute on both fresh and already-updated databases.

## Testing Strategy

### Frontend

- Update category page tests to remove code-field interactions and assertions.
- Keep keyword page tests working with category mocks that only require `id`, `name`, `enabled`, and `sort`.

### Backend

- Update category HTTP tests to stop sending or asserting `code`.
- Update schema-backed tests and seed fixtures to create categories without `code`.
- Ensure category CRUD still works and missing `orgid` validation still behaves the same.

### Verification

- run targeted admin Vitest coverage for category and keyword pages
- run targeted Go tests for `internal/modules/articleinspect`

## Documentation

Update active API docs that describe category payloads so they no longer mention `code`.

Historical design docs should remain unchanged unless they are acting as current operator guidance.

## Risks

- External callers may still read category `code` from API responses. After removal, those callers must stop depending on that field.
- Existing databases need a safe drop migration; careless DDL could break repeated migration runs.
- Search behavior narrows from `name OR code` to `name` only, which is acceptable because `code` is being removed as a managed concept.

## Out of Scope

- Renaming category fields beyond removing `code`
- Changing keyword-category relationships
- Backfilling or transforming category names

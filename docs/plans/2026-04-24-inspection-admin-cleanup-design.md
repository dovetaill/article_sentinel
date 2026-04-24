# Inspection Admin Cleanup and Rectify Editor Design

## Context

The current article inspection admin flow has four user-facing mismatches with the intended product behavior:

1. Task scope is already expressed by keyword-rule scopes, but task creation still exposes and stores a separate `include_body` concept.
2. Result list pages fetch only `xt_article_inspect_results`, so operators cannot see which field matched or what snippet triggered the hit without opening the detail drawer.
3. Rule deletion is supported by the backend but not exposed in the keyword list UI, while task deletion is missing end to end.
4. Content rectification edits the article body as raw HTML in a plain textarea, which makes normal editorial review and targeted fixes unnecessarily hard.

The cleanup should reduce duplicated task configuration, make result triage readable from the list itself, add missing destructive controls with safe policy boundaries, and replace the raw body textarea with an editor that can work in both rendered and source modes.

## Product Goals

- Make keyword-rule scopes the only business-level source of truth for scan coverage.
- Show enough hit context in result lists that operators can understand the trigger without opening a second screen.
- Allow operators to delete rules directly from the rule list.
- Allow operators to delete only `pending` and `failed` tasks, while protecting executed tasks.
- Let editors modify article body HTML either visually or in source form from the rectification page.

## Confirmed Decisions

### 1. Scope Ownership

Rule scopes remain canonical.

- Keyword rules continue to declare matching scope via `title`, `body`, `keyword`, and `rich_title`.
- The admin task form stops sending `include_body`.
- The backend treats `include_body` as a legacy compatibility field only and does not add any new behavior around it.
- Existing database columns can remain in place for now to avoid unnecessary migration churn.

This keeps scope definition in one place and removes the misleading idea that task-level body scanning overrides rule configuration.

### 2. Result Preview Behavior

Every result list row should show a compact preview of the first stored hit for that result.

Each row should expose:

- preview field name
- preview keyword text
- preview matched text
- preview snippet
- extra hit count when more than one hit exists

The first hit is defined as the earliest `xt_article_inspect_result_hits.id` for the result. This keeps ordering deterministic and mirrors how hits are currently created.

UI behavior:

- show human-readable field labels such as `标题` and `正文`
- show the snippet in a compact truncated block
- reveal the full snippet on hover
- highlight the matched text when it is available
- show `+N` additional hits when the result contains more than one stored hit

This applies to:

- global results list
- task detail embedded result list
- task result workspace list

The result detail drawer remains the place for the complete hit list.

### 3. Deletion Policy

#### Task deletion

Task deletion is added as a backend API plus admin UI action with strict status protection.

Allowed:

- `pending`
- `failed`

Rejected:

- `running`
- `success`
- `partial_success`

Deletion runs in one transaction and removes task-owned inspection data:

- `xt_article_inspect_task_keywords`
- `xt_article_inspect_result_hits`
- `xt_article_inspect_results`
- `xt_article_inspect_actions`
- `xt_article_inspect_operation_logs`
- `xt_article_inspect_field_change_logs`
- `xt_article_inspect_tasks`

If a queued worker later receives a deleted pending task, the existing `startTask` guard already fails safely because the task can no longer transition from `pending`.

#### Rule deletion

The existing keyword delete backend remains the same. The missing work is surfacing it in the keyword list UI with a confirmation action.

#### Category deletion

Category deletion already exists in the current admin and backend. This cleanup only verifies that the action remains available and behaves consistently alongside the new delete affordances.

### 4. Rectification Editor

The rectification page switches from a plain textarea to a dual-mode HTML editor.

#### Canonical data model

- The stored value remains a single HTML string.
- HTML source stays the canonical value passed to the backend.
- Visual editing updates that same string by serializing the edited DOM back to HTML.

#### Modes

1. `可视化编辑`
   - editable rendered HTML surface using `contentEditable`
   - suitable for direct text fixes, removals, and small structural cleanup
2. `HTML 源码`
   - raw HTML text editor for precise markup edits

Both modes edit the same body state, so users can switch between them without losing changes.

#### Why not convert to a schema-driven rich-text model

The article body already contains complex real-world HTML with inline styles, images, and embed-like markup. Converting it through a library-specific document model would risk dropping unsupported tags or rewriting layout unexpectedly.

A dual-mode HTML-first editor preserves fidelity better:

- visual mode supports direct editing of rendered content
- source mode supports exact markup fixes
- the backend contract remains unchanged

#### Editing affordances

The first iteration focuses on safe direct editing rather than a heavy authoring suite.

- visual mode uses a framed editable surface
- source mode uses a monospace textarea/editor pane
- the original article preview remains on the right for comparison
- save flows stay unchanged: `保存整改` and `保存并提交复核`

## Backend Design

### Result list DTOs

Introduce a list-specific DTO instead of returning raw `InspectionResult` rows directly.

Suggested fields:

- existing result fields
- `preview_field_name`
- `preview_keyword_text`
- `preview_matched_text`
- `preview_snippet`
- `extra_hit_count`

Repository flow:

1. page `xt_article_inspect_results`
2. load all `xt_article_inspect_result_hits` for the page's result IDs ordered by `result_id ASC, id ASC`
3. attach the first hit as the row preview
4. compute `extra_hit_count = max(total_hits - 1, 0)`

This avoids fragile SQL aggregation and keeps the behavior obvious in tests.

### Task delete service

Add a dedicated task delete path inside `TaskService` so policy and data cleanup live together.

Validation rules:

- `orgid` and `task_id` are required
- missing task returns `ErrTaskNotFound`
- unsupported status returns a new task delete error mapped to `400` or `409`

Deletion should be transactional and explicit rather than relying on implicit database cascade rules.

### API surface

Add:

- `DELETE /api/v1/article-inspect/tasks/{id}?orgid=<id>`

Keep existing routes for categories and keywords.

## Frontend Design

### Shared hit preview presentation

Create a shared UI helper/component for hit preview blocks so global results, task detail, and task results pages render snippets consistently.

Recommended behavior:

- field tag first
- keyword tag second
- snippet block below
- truncate long text to 2 lines in list contexts
- show full text via tooltip
- append `另有 N 条命中` when `extra_hit_count > 0`

### Task list actions

Task rows gain a delete action with confirmation text that explains the status restriction.

- pending and failed rows show enabled delete action
- other rows show disabled delete action or a protected-state hint

### Keyword list actions

Keyword rows gain a delete action next to edit.

### Task create form

Remove the `include_body` payload from the admin task creation form and related frontend typings.

### Rectification page

Replace the body textarea with tabbed editing modes:

- left: title, summary, body editor tabs
- right: original article comparison

The body editor area should visually separate editable content from the read-only comparison panel.

## Testing Strategy

### Backend

Add or extend tests for:

- result list preview fields populated from stored hits
- task delete allowed for `pending` and `failed`
- task delete rejected for `running`, `success`, and `partial_success`
- delete transaction removes task-owned dependent rows

### Frontend

Add or extend tests for:

- result list pages render field label, snippet, and extra-hit text
- task list delete action visibility and behavior
- keyword list delete action behavior
- task create no longer sends `include_body`
- rectification page supports visual/source body editing and saves HTML correctly

## Rollout Notes

- No schema migration is required for the first pass because `include_body` stays as a legacy field.
- Existing tasks and results remain readable because list preview data is derived from already stored `result_hits` rows.
- The cleanup intentionally avoids changing article lifecycle semantics or the rectification backend contract.

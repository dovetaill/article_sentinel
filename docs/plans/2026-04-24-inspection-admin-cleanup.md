# Inspection Admin Cleanup Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Remove duplicated task scan-scope configuration, show useful hit context in result lists, add safe task and rule deletion flows, and replace raw HTML rectification with a dual-mode body editor.

**Architecture:** Keep the existing inspection schema and backend contracts where possible. Add a list-specific result DTO backed by existing `result_hits`, add a guarded task deletion transaction in the task service, expose existing keyword deletion from the admin UI, and introduce an HTML-first rectification editor that supports both visual editing and source editing without converting article bodies into a library-specific document model.

**Tech Stack:** Go 1.25, GORM, Huma v2, React 18, TypeScript, Vite, Ant Design, ProComponents, Vitest, Testing Library.

---

### Task 1: Add result-list hit preview DTOs and backend tests

**Files:**
- Modify: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/repository_results.go`
- Modify: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/service_results.go`
- Create: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/dto_results.go`
- Test: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Write the failing test**

Add backend tests that seed one result with multiple hits and assert the list API returns:

- the first hit field name
- the first hit keyword text
- the first hit matched text
- the first hit snippet
- the extra-hit count

Use a narrow case where the earliest hit is `title` and the second hit is `body` so ordering is obvious.

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/modules/articleinspect -run 'Test(ResultListIncludesHitPreview|HandlerKeywordTaskAndResultsRoutes)' -v
```

Expected: FAIL because list rows currently expose only raw `InspectionResult` fields.

**Step 3: Write minimal implementation**

Implement a result-list DTO and repository helper that:

1. pages `InspectionResult` rows
2. loads all `InspectionResultHit` rows for the page's result IDs ordered by `result_id ASC, id ASC`
3. attaches the first hit as preview metadata
4. computes `extra_hit_count`

Keep the detail API unchanged.

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/modules/articleinspect -run 'Test(ResultListIncludesHitPreview|HandlerKeywordTaskAndResultsRoutes)' -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/dto_results.go internal/modules/articleinspect/repository_results.go internal/modules/articleinspect/service_results.go internal/modules/articleinspect/articleinspect_test.go
git commit -m "feat: add result hit previews to list responses"
```

### Task 2: Render hit previews consistently across result list pages

**Files:**
- Create: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/hit-preview.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/services/results.ts`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/results.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/detail.tsx`
- Test: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/index.test.tsx`
- Test: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/results.test.tsx`
- Test: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/detail.test.tsx`

**Step 1: Write the failing tests**

Update the page tests so list rows must render:

- a human-readable field label
- a visible preview snippet
- highlighted matched text when present
- `另有 N 条命中` when extra hits exist

Use test fixtures with `preview_field_name`, `preview_keyword_text`, `preview_matched_text`, `preview_snippet`, and `extra_hit_count`.

**Step 2: Run tests to verify they fail**

Run:

```bash
npm --prefix web/admin test -- --runInBand src/pages/results/index.test.tsx src/pages/tasks/results.test.tsx src/pages/tasks/detail.test.tsx
```

Expected: FAIL because the UI still expects `snippet`/`matched_keyword` only.

**Step 3: Write minimal implementation**

Create a shared hit preview component that:

- maps backend field names to Chinese labels
- truncates long text to two lines
- exposes the full snippet in a tooltip
- highlights the matched text when present

Then switch all three list contexts to that shared renderer.

**Step 4: Run tests to verify they pass**

Run:

```bash
npm --prefix web/admin test -- --runInBand src/pages/results/index.test.tsx src/pages/tasks/results.test.tsx src/pages/tasks/detail.test.tsx
```

Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin/src/components/ui/hit-preview.tsx web/admin/src/services/results.ts web/admin/src/pages/results/index.tsx web/admin/src/pages/tasks/results.tsx web/admin/src/pages/tasks/detail.tsx web/admin/src/pages/results/index.test.tsx web/admin/src/pages/tasks/results.test.tsx web/admin/src/pages/tasks/detail.test.tsx
git commit -m "feat: show hit previews in result lists"
```

### Task 3: Add guarded task deletion and keep task payload scope-clean

**Files:**
- Modify: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/service_tasks.go`
- Modify: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/handler.go`
- Modify: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/articleinspect_test.go`
- Modify: `/home/wwwroot/article_sentinel/internal/modules/articleinspect/dto_tasks.go`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/services/tasks.ts`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/new.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.tsx`
- Test: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/new.test.tsx`
- Test: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.test.tsx`

**Step 1: Write the failing tests**

Backend tests:

- delete pending task succeeds and removes dependent rows
- delete failed task succeeds and removes dependent rows
- delete running/success/partial-success task fails

Frontend tests:

- new task submission no longer sends `include_body`
- task list shows delete action for pending/failed tasks and blocks executed tasks

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./internal/modules/articleinspect -run 'Test(TaskDelete|HandlerKeywordTaskAndResultsRoutes)' -v
npm --prefix web/admin test -- --runInBand src/pages/tasks/new.test.tsx src/pages/tasks/index.test.tsx
```

Expected: FAIL because task deletion route/service do not exist and the admin still submits `include_body`.

**Step 3: Write minimal implementation**

Implement:

- task service delete method with status validation
- transactional deletion of task-owned rows
- `DELETE /api/v1/article-inspect/tasks/{id}` route
- frontend `deleteTask` service helper
- delete button and confirmation in task list
- task-create payload cleanup so the admin no longer sends `include_body`

Keep backend compatibility by accepting legacy `include_body` if older clients still send it.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./internal/modules/articleinspect -run 'Test(TaskDelete|HandlerKeywordTaskAndResultsRoutes)' -v
npm --prefix web/admin test -- --runInBand src/pages/tasks/new.test.tsx src/pages/tasks/index.test.tsx
```

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/service_tasks.go internal/modules/articleinspect/handler.go internal/modules/articleinspect/articleinspect_test.go internal/modules/articleinspect/dto_tasks.go web/admin/src/services/tasks.ts web/admin/src/pages/tasks/new.tsx web/admin/src/pages/tasks/index.tsx web/admin/src/pages/tasks/new.test.tsx web/admin/src/pages/tasks/index.test.tsx
git commit -m "feat: add guarded task deletion"
```

### Task 4: Expose keyword deletion from the rule list

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/keywords/index.tsx`
- Test: `/home/wwwroot/article_sentinel/web/admin/src/pages/keywords/index.test.tsx`

**Step 1: Write the failing test**

Add a UI test that loads one keyword row, clicks `删除规则`, confirms the dialog, and asserts the delete service is called and the table reloads.

**Step 2: Run test to verify it fails**

Run:

```bash
npm --prefix web/admin test -- --runInBand src/pages/keywords/index.test.tsx
```

Expected: FAIL because the rule list does not expose a delete action.

**Step 3: Write minimal implementation**

Add a delete action with `Popconfirm` to the keyword list page, reusing the existing backend `DELETE /keywords/{id}` endpoint through the frontend service.

**Step 4: Run test to verify it passes**

Run:

```bash
npm --prefix web/admin test -- --runInBand src/pages/keywords/index.test.tsx
```

Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin/src/pages/keywords/index.tsx web/admin/src/pages/keywords/index.test.tsx
git commit -m "feat: add keyword deletion to rule list"
```

### Task 5: Replace rectification body textarea with a dual-mode HTML editor

**Files:**
- Create: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/html-body-editor.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/rectify.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles.css`
- Test: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/rectify.test.tsx`

**Step 1: Write the failing tests**

Add tests that verify:

- the rectification page shows `可视化编辑` and `HTML 源码` tabs
- editing body text in visual mode updates the saved HTML payload
- editing raw HTML in source mode updates the saved HTML payload
- existing title/summary save behavior still works

Use a body fixture that contains visible text wrapped in markup so both modes can assert against the same HTML string.

**Step 2: Run test to verify it fails**

Run:

```bash
npm --prefix web/admin test -- --runInBand src/pages/articles/rectify.test.tsx
```

Expected: FAIL because the page still renders a plain textarea.

**Step 3: Write minimal implementation**

Implement a shared HTML body editor component with:

- `Tabs` for visual and source modes
- a `contentEditable` visual surface bound to the current HTML string
- a source textarea bound to the same string
- change handlers that serialize visual edits back to HTML

Then replace the body `ProFormTextArea` in the rectification page with that component and keep the existing save/review flows unchanged.

**Step 4: Run test to verify it passes**

Run:

```bash
npm --prefix web/admin test -- --runInBand src/pages/articles/rectify.test.tsx
```

Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin/src/components/ui/html-body-editor.tsx web/admin/src/pages/articles/rectify.tsx web/admin/src/styles.css web/admin/src/pages/articles/rectify.test.tsx
git commit -m "feat: add dual-mode html rectification editor"
```

### Task 6: Final verification sweep

**Files:**
- Modify: `/home/wwwroot/article_sentinel/docs/plans/2026-04-24-inspection-admin-cleanup-design.md`
- Modify: `/home/wwwroot/article_sentinel/README.md` if API/UI behavior docs changed during implementation

**Step 1: Run focused backend verification**

Run:

```bash
go test ./internal/modules/articleinspect -v
```

Expected: PASS.

**Step 2: Run focused frontend verification**

Run:

```bash
npm --prefix web/admin test -- --runInBand src/pages/results/index.test.tsx src/pages/tasks/results.test.tsx src/pages/tasks/detail.test.tsx src/pages/tasks/new.test.tsx src/pages/tasks/index.test.tsx src/pages/keywords/index.test.tsx src/pages/articles/rectify.test.tsx
npm --prefix web/admin run build
```

Expected: PASS.

**Step 3: Update any user-facing docs that changed**

If the implementation adjusted task creation semantics or task deletion behavior in visible docs, update `README.md` or the design doc accordingly.

**Step 4: Commit**

```bash
git add docs/plans/2026-04-24-inspection-admin-cleanup-design.md README.md
if ! git diff --cached --quiet; then git commit -m "docs: finalize inspection admin cleanup notes"; fi
```

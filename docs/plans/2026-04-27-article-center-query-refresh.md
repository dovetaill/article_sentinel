# Article Center Query Refresh Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make article-center searches re-fetch on repeated clicks, block duplicate submissions while loading, and default the article-center list to published plus offline articles.

**Architecture:** Lock the new behavior with frontend and backend tests first. Then update the React page to drive fetches with an explicit refresh trigger and loading guard, and update the Go article-list repository to default unspecified state filters to `IN (8, 9)` while leaving inspection-task candidate scanning unchanged.

**Tech Stack:** React, Ant Design, TypeScript, Vitest, Go, GORM

---

### Task 1: Lock frontend query behavior

**Files:**
- Modify: `web/admin/src/pages/articles/index.test.tsx`
- Modify: `web/admin/src/services/articles.test.ts`

**Step 1:** Write failing tests for repeated same-condition queries, loading-time duplicate-click protection, and the article-center request omitting `state=9`.

**Step 2:** Run `npm --prefix web/admin test -- src/pages/articles/index.test.tsx src/services/articles.test.ts` and confirm the new expectations fail for the current implementation.

### Task 2: Lock backend visible-state behavior

**Files:**
- Modify: `internal/modules/articleinspect/articleinspect_test.go`

**Step 1:** Add a pending-state fixture plus failing expectations that the default article list only returns offline/online rows.

**Step 2:** Run `go test ./internal/modules/articleinspect -run TestHandler` and confirm the new expectations fail for the current implementation.

### Task 3: Implement frontend query guards

**Files:**
- Modify: `web/admin/src/pages/articles/index.tsx`

**Step 1:** Add an explicit refresh trigger so clicking query/reset re-fetches even when submitted filters are unchanged.

**Step 2:** Enter loading state immediately on query/reset/pagination, disable repeated interactions while the request is active, and ignore stale responses.

**Step 3:** Stop passing `state: 9` from the article-center page.

**Step 4:** Re-run `npm --prefix web/admin test -- src/pages/articles/index.test.tsx src/services/articles.test.ts` and confirm green.

### Task 4: Implement backend visible-state defaults

**Files:**
- Modify: `internal/modules/articleinspect/repository_articles.go`

**Step 1:** When the article list endpoint receives no explicit `state`, default the query to `state IN (8, 9)`.

**Step 2:** Preserve exact-state behavior when a `state` query parameter is provided.

**Step 3:** Re-run `go test ./internal/modules/articleinspect -run TestHandler` and confirm green.

### Task 5: Verify end-to-end scope

**Files:**
- Verify only: `internal/modules/articleinspect/service_tasks.go`
- Verify only: `internal/modules/articleinspect/repository_articles.go`

**Step 1:** Re-check that inspection-task creation still defaults to `ArticleStateOnline`.

**Step 2:** Run focused frontend and backend verification again, then summarize the unchanged inspection-task scope alongside the article-center fix.

# Article Center Filter Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Restrict article-center filters to title fuzzy match and exact article ID, and show plain article IDs in task-result pages.

**Architecture:** Update the admin page and article service to send explicit filter params, then update the Go article list endpoint/repository to consume the same explicit params. Finally expose article ID in both result tables and verify with frontend/backend tests.

**Tech Stack:** React, Ant Design, TypeScript, Vitest, Go, GORM

---

### Task 1: Lock expected frontend filter behavior with tests
- Modify: `web/admin/src/pages/articles/index.test.tsx`
- Modify: `web/admin/src/services/articles.test.ts`
- Modify: `web/admin/src/pages/tasks/results.test.tsx`
- Step 1: Add failing assertions for separate title/article ID filters and plain numeric article IDs.
- Step 2: Run targeted Vitest commands to confirm failures.

### Task 2: Implement frontend filter and display changes
- Modify: `web/admin/src/pages/articles/index.tsx`
- Modify: `web/admin/src/services/articles.ts`
- Modify: `web/admin/src/pages/tasks/results.tsx`
- Modify: `web/admin/src/pages/results/index.tsx`
- Step 1: Replace generic search with explicit `title` and `article_id` params.
- Step 2: Add article ID column to both result tables and remove `#` prefix from article-center IDs.
- Step 3: Re-run targeted Vitest commands to confirm green.

### Task 3: Lock expected backend query behavior with tests
- Modify: `internal/modules/articleinspect/articleinspect_test.go`
- Step 1: Add failing endpoint tests for `title` fuzzy query and `article_id` exact query.
- Step 2: Run targeted Go tests to confirm failures.

### Task 4: Implement backend article list query changes
- Modify: `internal/modules/articleinspect/handler.go`
- Modify: `internal/modules/articleinspect/dto_articles.go`
- Modify: `internal/modules/articleinspect/service_articles.go`
- Modify: `internal/modules/articleinspect/repository_articles.go`
- Step 1: Parse explicit `title` and `article_id` filters.
- Step 2: Apply exact article ID matching and title-only fuzzy filtering.
- Step 3: Re-run targeted Go tests and then combined verification.

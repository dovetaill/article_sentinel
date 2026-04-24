# Task Detail Snapshot Cleanup Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Hide legacy `include_body` fields from the task detail page's request/rule snapshot display without changing stored backend data.

**Architecture:** Keep the API payload unchanged and sanitize snapshot JSON only in the task detail page before formatting it for display. Add a focused UI regression test first so the page is locked to never render `include_body` again.

**Tech Stack:** React, TypeScript, Vitest, Testing Library

---

### Task 1: Lock the UI behavior with a failing test

**Files:**
- Modify: `web/admin/src/pages/tasks/detail.test.tsx`
- Test: `web/admin/src/pages/tasks/detail.test.tsx`

**Step 1: Write the failing test**
- Extend the existing task detail page test fixture so `request_snapshot` includes `include_body`.
- Assert the rendered request snapshot still shows useful fields like `title_like`.
- Assert the page does not render the string `include_body` anywhere.

**Step 2: Run test to verify it fails**
- Run: `npm --prefix web/admin test -- --runInBand src/pages/tasks/detail.test.tsx`
- Expected: FAIL because the current page renders raw formatted JSON including `include_body`.

### Task 2: Implement minimal snapshot sanitization in the page

**Files:**
- Modify: `web/admin/src/pages/tasks/detail.tsx`

**Step 3: Write minimal implementation**
- Add a small helper that parses snapshot JSON, recursively removes `include_body` keys, and re-formats the sanitized value.
- Reuse the helper for both request snapshot and rule snapshot tabs so any historical residue is hidden consistently.
- Preserve current fallback behavior for empty or invalid snapshots.

**Step 4: Run test to verify it passes**
- Run: `npm --prefix web/admin test -- --runInBand src/pages/tasks/detail.test.tsx`
- Expected: PASS.

### Task 3: Run focused verification

**Files:**
- Verify only

**Step 5: Run related tests**
- Run: `npm --prefix web/admin test -- --runInBand src/pages/tasks/detail.test.tsx src/pages/tasks/results.test.tsx`
- Expected: PASS.

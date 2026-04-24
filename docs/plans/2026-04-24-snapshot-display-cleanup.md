# Snapshot Display Cleanup Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Hide legacy `include_body` keys anywhere the admin UI renders stored inspection request snapshots.

**Architecture:** Keep backend payloads untouched and sanitize snapshot text only in the frontend display layer. Extract one shared snapshot formatter so task detail and logs modal stay consistent and future snapshot views can reuse the same filtering behavior.

**Tech Stack:** React, TypeScript, Ant Design, Vitest, Testing Library

---

### Task 1: Lock logs page behavior with a failing test

**Files:**
- Modify: `web/admin/src/pages/logs/index.test.tsx`
- Test: `web/admin/src/pages/logs/index.test.tsx`

**Step 1: Write the failing test**
- Update the mocked `request_snapshot` to include `include_body` plus another useful field.
- Open the snapshot modal from the logs table.
- Assert the useful field is still shown and `include_body` is not rendered.

**Step 2: Run test to verify it fails**
Run: `npm --prefix web/admin test -- --runInBand src/pages/logs/index.test.tsx`
Expected: FAIL because the modal currently renders raw snapshot text.

### Task 2: Extract shared snapshot formatting helper

**Files:**
- Create: `web/admin/src/lib/inspection-snapshot.ts`
- Modify: `web/admin/src/pages/tasks/detail.tsx`
- Modify: `web/admin/src/pages/logs/index.tsx`

**Step 3: Write minimal implementation**
- Move the existing task detail snapshot sanitization into a shared helper.
- Helper behavior: parse JSON when possible, recursively remove `include_body`, pretty-print the result, and fall back to filtering raw lines if the snapshot is malformed.
- Use the helper in task detail and in the logs modal.

**Step 4: Run targeted tests to verify it passes**
Run: `npm --prefix web/admin test -- --runInBand src/pages/logs/index.test.tsx src/pages/tasks/detail.test.tsx`
Expected: PASS.

### Task 3: Run focused regression coverage

**Files:**
- Verify only

**Step 5: Run related tests**
Run: `npm --prefix web/admin test -- --runInBand src/pages/logs/index.test.tsx src/pages/tasks/detail.test.tsx src/pages/tasks/results.test.tsx`
Expected: PASS.

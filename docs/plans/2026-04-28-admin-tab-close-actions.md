# Admin Tab Close Actions Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Simplify the admin workbench tab action menu to only `关闭其他` and `关闭全部`, and make `关闭其他` keep only the active tab.

**Architecture:** Keep the existing workbench provider/session architecture intact and change only the tab action menu plus the reducer branch that handles bulk-close behavior. Lock the new semantics with store-level and UI-level regression tests before changing implementation.

**Tech Stack:** React 18, React Router 7, Ant Design 5, Vitest, Testing Library, TypeScript

---

### Task 1: Lock the new close-action semantics in tests

**Files:**
- Modify: `web/admin/src/workbench/store.test.ts`
- Modify: `web/admin/src/workbench/tabs.test.tsx`

**Step 1: Write the failing store test**

Add a test that opens `/articles` and `/tasks/new`, activates `/tasks/new`, runs `closeOtherTabs('/tasks/new')`, and expects only `/tasks/new` to remain.

**Step 2: Run test to verify it fails**

Run: `npm --prefix web/admin test -- src/workbench/store.test.ts`
Expected: FAIL because `/tasks` is still preserved alongside `/tasks/new`.

**Step 3: Write the failing tabs test**

Add a test that opens the tag action menu and expects only `关闭其他` and `关闭全部` to be present, while `关闭当前` / `关闭左侧` / `关闭右侧` are absent.

**Step 4: Run test to verify it fails**

Run: `npm --prefix web/admin test -- src/workbench/tabs.test.tsx`
Expected: FAIL because the extra menu items are still rendered.

**Step 5: Commit**

```bash
git add web/admin/src/workbench/store.test.ts web/admin/src/workbench/tabs.test.tsx
git commit -m "test: lock admin tab close action semantics"
```

### Task 2: Implement the minimal reducer and menu changes

**Files:**
- Modify: `web/admin/src/workbench/store.tsx`
- Modify: `web/admin/src/workbench/tabs.tsx`

**Step 1: Write the minimal reducer change**

Update `closeTabsByAction(..., 'others', ...)` so that it keeps only the target tab instead of keeping the target tab plus the base tab.

**Step 2: Write the minimal menu change**

Remove `关闭当前` / `关闭左侧` / `关闭右侧` from the dropdown items and from the click switch so only `others` and `all` remain.

**Step 3: Run focused tests to verify they pass**

Run: `npm --prefix web/admin test -- src/workbench/store.test.ts src/workbench/tabs.test.tsx`
Expected: PASS.

**Step 4: Run broader workbench regression coverage**

Run: `npm --prefix web/admin test -- src/workbench/provider.test.tsx src/workbench/tabs.test.tsx src/workbench/store.test.ts`
Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin/src/workbench/store.tsx web/admin/src/workbench/tabs.tsx
git commit -m "fix: simplify admin tab close actions"
```

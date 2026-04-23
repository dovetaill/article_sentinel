# Final Admin Rhythm Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Deliver one final visual pass that compresses shell rhythm, summary cards, tables, tabs, and form controls so the admin UI lands closer to the default `shadcn-admin` workbench feel.

**Architecture:** Keep the current routes and page structures intact, then do the final refinement primarily through shared shell JSX and CSS tokens/layout rules. Lock the desired shell and control density with CSS regression tests first, then make minimal JSX changes where the sidebar brand block needs structural simplification.

**Tech Stack:** React 18, TypeScript, React Router 7, Ant Design, custom CSS tokens/layouts, Vitest.

---

### Task 1: Lock the final shell rhythm in a failing style regression test

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/visual-tokens.test.ts`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/App.test.tsx`

**Step 1: Write the failing test**

Extend the CSS regression test to assert the final compressed shell values and neutral component selectors exist.

```tsx
expect(layoutCss).toContain('width: 36px;');
expect(layoutCss).toContain('padding: 14px 14px;');
expect(layoutCss).toContain('.section-card .ant-tabs-ink-bar');
```

Also keep `App.test.tsx` free of old oversized shell copy.

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/styles/visual-tokens.test.ts src/App.test.tsx`
Expected: FAIL because the current shell still uses looser spacing and larger brand block dimensions.

**Step 3: Write minimal implementation**

Only add the smallest set of target assertions that represent the desired final rhythm.

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/styles/visual-tokens.test.ts src/App.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add /home/wwwroot/article_sentinel/web/admin/src/styles/visual-tokens.test.ts /home/wwwroot/article_sentinel/web/admin/src/App.test.tsx
git commit -m "test(admin): lock final shell rhythm"
```

### Task 2: Compress the sidebar brand block and topbar structure

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/layout/sidebar-nav.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing test**

Use the CSS regression test from Task 1 as the red test, then adjust `App.test.tsx` only if the sidebar brand text structure changes while keeping the product title present.

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/styles/visual-tokens.test.ts src/App.test.tsx`
Expected: FAIL until the brand block is thinner and the topbar shell is tighter.

**Step 3: Write minimal implementation**

- Simplify the sidebar brand to a single compact title group
- Reduce badge size and visual weight
- Shrink topbar padding, title scale, and breadcrumb emphasis

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/styles/visual-tokens.test.ts src/App.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add /home/wwwroot/article_sentinel/web/admin/src/components/layout/sidebar-nav.tsx /home/wwwroot/article_sentinel/web/admin/src/styles/layout.css
git commit -m "style(admin): compress sidebar and topbar rhythm"
```

### Task 3: Flatten summary cards and tighten table density

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/keywords/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.test.tsx`

**Step 1: Write the failing test**

Add assertions that the task/log toolbar actions remain present after the tighter workbench spacing.

```tsx
expect(screen.getByRole('button', { name: '查询任务' })).toBeInTheDocument();
expect(screen.getByRole('button', { name: '查询日志' })).toBeInTheDocument();
```

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/tasks/index.test.tsx src/pages/logs/index.test.tsx src/styles/visual-tokens.test.ts`
Expected: FAIL until the new final density values are applied.

**Step 3: Write minimal implementation**

- Lower summary-card padding and number scale
- Reduce table header and row padding one more step
- Make toolbars and action links lighter and thinner

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/tasks/index.test.tsx src/pages/logs/index.test.tsx src/styles/visual-tokens.test.ts`
Expected: PASS.

**Step 5: Commit**

```bash
git add /home/wwwroot/article_sentinel/web/admin/src/styles/layout.css /home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/results/index.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/keywords/index.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.test.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.test.tsx
git commit -m "style(admin): tighten summary cards and tables"
```

### Task 4: Tighten tabs and form controls across list/detail workspaces

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/detail.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/detail.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/rectify.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/detail.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/detail.test.tsx`

**Step 1: Write the failing test**

Keep the existing detail-page tests and extend the CSS regression test with the final tab/control selectors.

```tsx
expect(layoutCss).toContain('.section-card .ant-tabs-tab');
expect(layoutCss).toContain('.ant-select-item-option-selected');
expect(layoutCss).toContain('.ant-input-outlined:focus');
```

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/styles/visual-tokens.test.ts src/pages/articles/detail.test.tsx src/pages/tasks/detail.test.tsx`
Expected: FAIL until the final tab and form-control rhythm is applied.

**Step 3: Write minimal implementation**

- Make tabs smaller and quieter
- Reduce tab ink bar and spacing
- Tighten input/select/picker rhythm and neutral focus state
- Keep article/task detail tabs readable while matching list-page controls

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/styles/visual-tokens.test.ts src/pages/articles/detail.test.tsx src/pages/tasks/detail.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add /home/wwwroot/article_sentinel/web/admin/src/styles/layout.css /home/wwwroot/article_sentinel/web/admin/src/pages/articles/detail.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/tasks/detail.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/articles/rectify.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/articles/detail.test.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/tasks/detail.test.tsx
git commit -m "style(admin): unify tabs and form controls"
```

### Task 5: Run full verification and create the final polish commit

**Files:**
- Reference: `/home/wwwroot/article_sentinel/web/admin/src/**`

**Step 1: Run the full test suite**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test`
Expected: PASS with zero failing tests.

**Step 2: Run the production build**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm run build`
Expected: PASS.

**Step 3: Review the diff**

Run: `cd /home/wwwroot/article_sentinel && git diff --stat`
Expected: only the final rhythm-alignment files plus the plan file are modified.

**Step 4: Commit**

```bash
git add /home/wwwroot/article_sentinel/web/admin/src /home/wwwroot/article_sentinel/docs/plans/2026-04-23-admin-ui-final-rhythm.md
git commit -m "style(admin): finish the workbench rhythm polish"
```

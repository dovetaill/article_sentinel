# Neutral Admin Alignment Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Push the admin UI from the current polished state to a colder, more neutral `shadcn-admin` style by removing blue emphasis and tightening spacing, borders, and component rhythm.

**Architecture:** Keep the current route structure, task/article full-page details, and page composition intact. Perform the work in two layers: first lock the desired behavior with tests, then shift theme tokens and layout primitives so pages inherit the new visual system with only light page-level adjustments.

**Tech Stack:** React 18, TypeScript, React Router 7, Ant Design, custom CSS tokens/layouts, Vitest, Testing Library.

---

### Task 1: Lock the neutral shell and shared primitives in tests

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/App.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/components.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/layout/sidebar-nav.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/layout/topbar.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/page-header.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/section-card.tsx`

**Step 1: Write the failing test**

Add assertions that the shell stays minimal and that shared components keep the current compact, utility-first structure.

```tsx
expect(screen.queryByText('政务融媒')).not.toBeInTheDocument();
expect(screen.queryByText('值守中')).not.toBeInTheDocument();
expect(screen.queryByText('栏目说明')).not.toBeInTheDocument();
```

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/App.test.tsx src/components/ui/components.test.tsx`
Expected: FAIL if the structure regresses while visual tightening is in progress.

**Step 3: Write minimal implementation**

Keep the shell copy minimal and avoid adding presentation-only metadata back into shared components.

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/App.test.tsx src/components/ui/components.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add /home/wwwroot/article_sentinel/web/admin/src/App.test.tsx /home/wwwroot/article_sentinel/web/admin/src/components/ui/components.test.tsx /home/wwwroot/article_sentinel/web/admin/src/components/layout/sidebar-nav.tsx /home/wwwroot/article_sentinel/web/admin/src/components/layout/topbar.tsx /home/wwwroot/article_sentinel/web/admin/src/components/ui/page-header.tsx /home/wwwroot/article_sentinel/web/admin/src/components/ui/section-card.tsx
git commit -m "test(admin): lock neutral shell primitives"
```

### Task 2: Replace blue-centric tokens with a neutral shadcn-like palette

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/theme.css`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles.css`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing test**

Add a CSS/token-oriented regression assertion by checking for classes or structural expectations already consumed by tests, then run the focused visual tests.

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/App.test.tsx src/components/ui/components.test.tsx`
Expected: FAIL if any supporting structure changed incorrectly.

**Step 3: Write minimal implementation**

Update tokens to a neutral palette and lighter surface system.

```css
--admin-bg-base: #f8f8f9;
--admin-accent: #18181b;
--admin-surface-1: #ffffff;
--admin-line-soft: rgba(24, 24, 27, 0.08);
```

Then remove remaining blue-heavy gradients, active states, and glossy effects from layout styles.

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/App.test.tsx src/components/ui/components.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add /home/wwwroot/article_sentinel/web/admin/src/styles/theme.css /home/wwwroot/article_sentinel/web/admin/src/styles.css /home/wwwroot/article_sentinel/web/admin/src/styles/layout.css
git commit -m "style(admin): switch UI tokens to neutral palette"
```

### Task 3: Tighten sidebar, topbar, cards, and buttons toward shadcn-admin rhythm

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/layout/sidebar-nav.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/layout/topbar.tsx`

**Step 1: Write the failing test**

Use existing shell/component tests to protect structure while changing density and style.

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/App.test.tsx src/components/ui/components.test.tsx`
Expected: FAIL if structure accidentally shifts.

**Step 3: Write minimal implementation**

- Make active sidebar items gray instead of blue
- Lower card radii and shadows
- Reduce topbar/banner feeling
- Make primary buttons charcoal instead of blue

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/App.test.tsx src/components/ui/components.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add /home/wwwroot/article_sentinel/web/admin/src/styles/layout.css /home/wwwroot/article_sentinel/web/admin/src/components/layout/sidebar-nav.tsx /home/wwwroot/article_sentinel/web/admin/src/components/layout/topbar.tsx
git commit -m "style(admin): tighten shell spacing and neutral actions"
```

### Task 4: Tighten data tables and filter bars on keyword, task, result, and log pages

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/keywords/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.test.tsx`

**Step 1: Write the failing test**

Keep route and filter behavior assertions intact and add any needed checks that the task/log toolbar controls still exist after layout tightening.

```tsx
expect(screen.getByRole('button', { name: '查询任务' })).toBeInTheDocument();
expect(screen.getByRole('button', { name: '查询日志' })).toBeInTheDocument();
```

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/tasks/index.test.tsx src/pages/logs/index.test.tsx`
Expected: FAIL if toolbar structure changes unexpectedly.

**Step 3: Write minimal implementation**

- Reduce toolbar padding and input height
- Lower table header emphasis
- Keep actions as restrained text links
- Ensure task and log routes remain linked

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/tasks/index.test.tsx src/pages/logs/index.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add /home/wwwroot/article_sentinel/web/admin/src/pages/keywords/index.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/results/index.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.test.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.test.tsx
git commit -m "style(admin): refine table density and filter bars"
```

### Task 5: Neutralize the article and task detail workspaces

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/detail.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/detail.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/detail.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/detail.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing test**

Keep the existing full-page detail expectations and add any assertions needed for stable entry points like action buttons and tabs.

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/articles/detail.test.tsx src/pages/tasks/detail.test.tsx`
Expected: FAIL if page structure drifts while reducing visual noise.

**Step 3: Write minimal implementation**

- Soften summary cards
- Make side panels and code blocks more neutral
- Keep tabs and detail lists consistent with the same gray workbench language

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/articles/detail.test.tsx src/pages/tasks/detail.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add /home/wwwroot/article_sentinel/web/admin/src/pages/articles/detail.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/tasks/detail.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/articles/detail.test.tsx /home/wwwroot/article_sentinel/web/admin/src/pages/tasks/detail.test.tsx /home/wwwroot/article_sentinel/web/admin/src/styles/layout.css
git commit -m "style(admin): neutralize detail workspaces"
```

### Task 6: Run complete verification and create the final polish commit

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
Expected: only the intended neutral visual alignment files are modified.

**Step 4: Commit**

```bash
git add /home/wwwroot/article_sentinel/web/admin/src /home/wwwroot/article_sentinel/docs/plans/2026-04-23-admin-ui-neutral-alignment.md
git commit -m "style(admin): align workbench with neutral shadcn palette"
```

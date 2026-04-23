# Admin UI Polish And Task Detail Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refine the admin UI so keyword/task/log pages feel closer to `shadcn-admin`, remove non-functional shell copy, and replace task-detail drawers with a full `/tasks/:taskId` detail page.

**Architecture:** Keep the first-pass dashboard shell and article-detail model, then iterate on component density, filter-bar composition, and action styling. Add a task-detail page that composes existing task, result, and log APIs into the same full-page detail model already used for articles.

**Tech Stack:** React 18, TypeScript, React Router 7, Ant Design, custom CSS tokens/layouts, Vitest, Testing Library.

---

### Task 1: Remove non-functional shell copy and lock the cleaner chrome in tests

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/App.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/layout/sidebar-nav.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/layout/topbar.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing shell assertions**

Update `web/admin/src/App.test.tsx` so it explicitly rejects the remaining chrome copy:

```tsx
expect(screen.queryByText('当前环境')).not.toBeInTheDocument();
expect(screen.queryByText('值守模式')).not.toBeInTheDocument();
expect(screen.queryByText('政务融媒')).not.toBeInTheDocument();
expect(screen.queryByText('值守中')).not.toBeInTheDocument();
```

**Step 2: Run the test and verify RED**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/App.test.tsx`
Expected: FAIL because the sidebar footer and topbar tags still render this copy.

**Step 3: Implement the cleaner shell**

Remove the footer block and topbar tags so the shell becomes:

```tsx
<aside className="admin-sidebar">
  <Brand />
  <SidebarNav />
</aside>
<header className="admin-topbar">
  <Breadcrumbs />
  <PageTitle />
</header>
```

**Step 4: Re-run the shell test and verify GREEN**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/App.test.tsx`
Expected: PASS.

### Task 2: Tighten shared density for page headers, cards, buttons, and filter rows

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/components.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/page-header.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/section-card.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/toolbar-strip.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/theme.css`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing shared-component tests**

Add assertions for a denser card/filter presentation, for example:

```tsx
expect(screen.getByText('按筛选条件查看当前命中记录。')).toBeInTheDocument();
expect(screen.queryByText('栏目说明')).not.toBeInTheDocument();
```

If needed, add a new filter-row test case that checks grouped controls and action alignment.

**Step 2: Run the component test and verify RED**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/components/ui/components.test.tsx`
Expected: FAIL until the shared components expose the cleaner structure.

**Step 3: Implement denser shared primitives**

Refine these patterns:

```tsx
<PageHeader title="检测任务" description="统一发起巡检任务。" extra={<Button />} />
<SectionCard title="任务列表" description="查看任务状态和扫描规模。">...</SectionCard>
<ToolbarStrip>filter controls + compact actions</ToolbarStrip>
```

Reduce padding, tighten gaps, and standardize light text-action buttons.

**Step 4: Re-run the component test and verify GREEN**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/components/ui/components.test.tsx`
Expected: PASS.

### Task 3: Refine the keyword rules page toward a cleaner data-table layout

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/keywords/index.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/keywords/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing keyword-page test**

Extend `index.test.tsx` to assert the refined page framing, for example:

```tsx
expect(screen.getByRole('button', { name: '新增规则' })).toBeInTheDocument();
expect(screen.getByText('统一查看当前生效规则、风险等级和适用范围。')).toBeInTheDocument();
```

If the test file does not cover it yet, add an assertion for denser action wording and the absence of oversized spacing cues.

**Step 2: Run the keyword test and verify RED**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/keywords/index.test.tsx`
Expected: FAIL if the new card/filter/header structure is not yet reflected.

**Step 3: Implement the keyword page polish**

- Add a compact filter row above the table
- Tighten the ProTable density and action column spacing
- Refine modal layout spacing and button alignment

**Step 4: Re-run the keyword test and verify GREEN**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/keywords/index.test.tsx`
Expected: PASS.

### Task 4: Replace task drawer behavior with a full task detail route

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/routes.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/detail.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/detail.test.tsx`

**Step 1: Write the failing task-route tests**

Update `src/pages/tasks/index.test.tsx` so task detail becomes a link target, not a drawer result:

```tsx
expect(screen.getByRole('link', { name: '查看详情' })).toHaveAttribute('href', '/tasks/77');
expect(screen.queryByText(/spam keyword/i)).not.toBeInTheDocument();
```

Create `src/pages/tasks/detail.test.tsx` with expectations such as:

```tsx
expect(await screen.findByText('inspect-20260420-01')).toBeInTheDocument();
expect(screen.getByRole('tab', { name: '命中结果' })).toBeInTheDocument();
expect(screen.getByText('规则快照')).toBeInTheDocument();
```

**Step 2: Run the task tests and verify RED**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/tasks/index.test.tsx src/pages/tasks/detail.test.tsx`
Expected: FAIL because the index page still opens a drawer and the detail page does not exist.

**Step 3: Implement the full task detail page**

Add `/tasks/:taskId` and build the page with this load sequence:

```ts
const task = await getTaskDetail(taskId, 100);
const results = await listResults({ orgid: 100, task_id: Number(taskId), page: 1, pageSize: 20 });
const logs = await listOperationLogs({ orgid: 100, task_id: Number(taskId), page: 1, pageSize: 20 });
```

Render:

- top navigation + status
- summary metrics
- left/right detail sections
- tabs for hit results, rule snapshot, request snapshot, and logs

Then change the task list action to a simple link.

**Step 4: Re-run the task tests and verify GREEN**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/tasks/index.test.tsx src/pages/tasks/detail.test.tsx`
Expected: PASS.

### Task 5: Polish the task list and new-task page to match the new detail model

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/new.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/new.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing task-page polish assertions**

Add assertions for the cleaner header actions and stronger task workbench framing.

**Step 2: Run the task-page tests and verify RED**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/tasks/new.test.tsx src/pages/tasks/index.test.tsx`
Expected: FAIL where the revised structure is not yet in place.

**Step 3: Implement the task-page polish**

- Tighten summary cards and list spacing
- Align the create page header, side notes, and action bar with the article-detail family
- Ensure list rows visually support “browse then enter full detail”

**Step 4: Re-run the task-page tests and verify GREEN**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/tasks/new.test.tsx src/pages/tasks/index.test.tsx`
Expected: PASS.

### Task 6: Refine the logs page filter bar and add task-detail links

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing logs-page assertions**

Extend the test to require both article and task navigation links:

```tsx
expect(screen.getByRole('link', { name: '#501' })).toHaveAttribute('href', '/articles/501');
expect(screen.getByRole('link', { name: '#77' })).toHaveAttribute('href', '/tasks/77');
```

Also add any assertions needed for the refined filter-row grouping.

**Step 2: Run the logs test and verify RED**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/logs/index.test.tsx`
Expected: FAIL until the task link and new filter layout exist.

**Step 3: Implement the logs page polish**

- Group filter controls more cleanly
- Add task-detail links in the task ID column
- Keep the snapshot modal, but make the trigger styling align with the lighter action system

**Step 4: Re-run the logs test and verify GREEN**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/logs/index.test.tsx`
Expected: PASS.

### Task 7: Full verification for the second-pass polish

**Files:**
- Reference: `/home/wwwroot/article_sentinel/web/admin/src/**`

**Step 1: Run the full admin test suite**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test`
Expected: PASS with zero failing tests.

**Step 2: Run the production build**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm run build`
Expected: exit code 0.

**Step 3: Review the diff**

Run: `cd /home/wwwroot/article_sentinel && git diff -- web/admin/src docs/plans/2026-04-23-admin-ui-polish-task-detail*`
Expected: only shell cleanup, task-detail routing, page polish, and plan/doc changes.

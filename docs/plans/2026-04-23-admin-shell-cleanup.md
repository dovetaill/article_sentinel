# Admin Shell Cleanup Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Remove the duplicated shell/page header model, promote `文稿列表` into the primary sidebar, and keep list pages visually compact while detail pages remain full standalone workspaces.

**Architecture:** Delete the shell-level `topbar` so the app returns to a plain `sidebar + content` frame, then move `文稿列表` into `appRoutes` so the sidebar renders it as a first-class destination. Rework list pages to start directly from summary cards, toolbars, and section cards, while trimming verbose descriptions from detail-page `PageHeader` usage instead of introducing a new header system.

**Tech Stack:** React 18, React Router 7, Ant Design 5, Pro Components tables, Vitest, Testing Library, Vite

---

### Task 1: Lock the shell regression expectations

**Files:**
- Modify: `web/admin/src/App.test.tsx`
- Reference: `web/admin/src/routes.tsx`
- Reference: `web/admin/src/components/layout/admin-shell.tsx`

**Step 1: Write the failing test**

Add assertions that capture the new shell contract:

```tsx
expect(appRoutes.map((route) => route.label)).toEqual([
  '关键词规则',
  '检测任务',
  '风险结果',
  '文稿列表',
  '操作日志',
]);
expect(screen.getByRole('link', { name: /文稿列表/i })).toBeInTheDocument();
expect(screen.queryByText('控制台')).not.toBeInTheDocument();
expect(screen.queryByText('留存任务执行与稿件处置过程的关键记录。')).not.toBeInTheDocument();
```

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/App.test.tsx`
Expected: FAIL because `appRoutes` still excludes `文稿列表` and the shell still renders the `topbar` breadcrumb/title copy.

**Step 3: Write minimal implementation**

Do not implement yet; note the shell changes needed next:

```tsx
// target shape after Task 2
export function App() {
  return (
    <AdminShell>
      <AppRouteOutlet />
    </AdminShell>
  );
}
```

**Step 4: Run test to verify it still fails for the intended reason**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/App.test.tsx`
Expected: FAIL with the same route/topbar mismatch, confirming the regression is real before changing code.

**Step 5: Commit**

```bash
git add web/admin/src/App.test.tsx
git commit -m "test(admin): capture shell cleanup regressions"
```

### Task 2: Remove the shell topbar and promote `文稿列表` into primary navigation

**Files:**
- Modify: `web/admin/src/App.tsx`
- Modify: `web/admin/src/routes.tsx`
- Modify: `web/admin/src/components/layout/admin-shell.tsx`
- Modify: `web/admin/src/components/layout/sidebar-nav.tsx`
- Modify: `web/admin/src/styles/layout.css`
- Delete: `web/admin/src/components/layout/topbar.tsx`
- Test: `web/admin/src/App.test.tsx`

**Step 1: Write the minimal implementation**

Apply the shell/navigation changes:

```ts
export const appRoutes: AppRoute[] = [
  keywordsRoute,
  tasksRoute,
  resultsRoute,
  articlesRoute,
  logsRoute,
];
```

```tsx
export function AdminShell({ children }: PropsWithChildren) {
  return (
    <div className="admin-shell">
      <SidebarNav routes={appRoutes} />
      <div className="admin-shell__main">
        <main className="admin-shell__content">{children}</main>
      </div>
    </div>
  );
}
```

Implementation notes:
- Remove `useLocation` / `routeForPath` / `activeRoute` plumbing if they become unused.
- Add a sidebar icon entry for the `articles` route.
- Delete `.admin-topbar*` styles and reset `.admin-shell__content` min-height so it no longer assumes a topbar exists.
- Keep `article-detail`, `task-detail`, `article-rectify`, and `task-new` in hidden route matching only.

**Step 2: Run the focused test**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/App.test.tsx`
Expected: PASS.

**Step 3: Run a route smoke test**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/articles/index.test.tsx src/pages/tasks/detail.test.tsx`
Expected: PASS, proving the sidebar change did not break the article list route or detail routing.

**Step 4: Review the diff for dead imports and stale exports**

Run: `cd /home/wwwroot/article_sentinel && git diff -- web/admin/src/App.tsx web/admin/src/routes.tsx web/admin/src/components/layout/admin-shell.tsx web/admin/src/components/layout/sidebar-nav.tsx web/admin/src/styles/layout.css`
Expected: only shell/nav cleanup; no unrelated changes in `README.md` or `web/admin/vite.config.ts`.

**Step 5: Commit**

```bash
git add web/admin/src/App.tsx web/admin/src/routes.tsx web/admin/src/components/layout/admin-shell.tsx web/admin/src/components/layout/sidebar-nav.tsx web/admin/src/styles/layout.css web/admin/src/App.test.tsx
git rm web/admin/src/components/layout/topbar.tsx
git commit -m "refactor(admin): simplify shell and expose article list nav"
```

### Task 3: Lock compact list-page expectations before removing page headers

**Files:**
- Modify: `web/admin/src/pages/keywords/index.test.tsx`
- Modify: `web/admin/src/pages/tasks/index.test.tsx`
- Modify: `web/admin/src/pages/results/index.test.tsx`
- Modify: `web/admin/src/pages/logs/index.test.tsx`
- Modify: `web/admin/src/pages/articles/index.test.tsx`
- Reference: `web/admin/src/components/ui/page-header.tsx`

**Step 1: Write the failing tests**

Update each list-page test to assert that the old page-header descriptions are gone while core actions/content remain.

Example assertions:

```tsx
expect(screen.queryByText('统一维护规则词库、匹配方式、风险等级与建议处置。')).not.toBeInTheDocument();
expect(screen.getByRole('button', { name: '新增规则' })).toBeInTheDocument();
```

```tsx
expect(screen.queryByText('统一发起巡检任务，查看执行状态、扫描规模与命中情况。')).not.toBeInTheDocument();
expect(screen.getByRole('link', { name: '新建任务' })).toHaveAttribute('href', '/tasks/new');
```

```tsx
expect(screen.queryByText('集中查看命中文稿、风险等级与处置状态，按批次完成研判、下线与整改。')).not.toBeInTheDocument();
expect(screen.getByRole('button', { name: '批量下线处置' })).toBeInTheDocument();
```

```tsx
expect(screen.queryByText('查询任务执行、结果处置与请求快照。')).not.toBeInTheDocument();
expect(screen.getByRole('button', { name: '查询日志' })).toBeInTheDocument();
```

```tsx
expect(screen.queryByText('按稿件维度查看巡检命中、处置状态与最近任务，便于持续跟进单篇稿件。')).not.toBeInTheDocument();
expect(screen.getByRole('link', { name: '查看详情' })).toHaveAttribute('href', '/articles/501');
```

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/keywords/index.test.tsx src/pages/tasks/index.test.tsx src/pages/results/index.test.tsx src/pages/logs/index.test.tsx src/pages/articles/index.test.tsx`
Expected: FAIL because each page still renders `PageHeader` with the old descriptive copy.

**Step 3: Write down the intended component move**

Use the following target shape for the two pages that still need a top-level action:

```tsx
<SectionCard
  title="任务列表"
  description="按任务编号与执行状态浏览当前批次。"
  extra={<Button type="primary" href="/tasks/new">新建任务</Button>}
>
```

```tsx
<SectionCard
  title="规则列表"
  description="统一查看当前生效规则、风险等级和适用范围。"
  extra={<Button type="primary">新增规则</Button>}
>
```

**Step 4: Re-run the focused tests without implementation**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/tasks/index.test.tsx src/pages/keywords/index.test.tsx`
Expected: FAIL, still driven by the old `PageHeader` copy.

**Step 5: Commit**

```bash
git add web/admin/src/pages/keywords/index.test.tsx web/admin/src/pages/tasks/index.test.tsx web/admin/src/pages/results/index.test.tsx web/admin/src/pages/logs/index.test.tsx web/admin/src/pages/articles/index.test.tsx
git commit -m "test(admin): capture compact list page layout"
```

### Task 4: Remove large list-page headers and fold actions into the working surface

**Files:**
- Modify: `web/admin/src/pages/keywords/index.tsx`
- Modify: `web/admin/src/pages/tasks/index.tsx`
- Modify: `web/admin/src/pages/results/index.tsx`
- Modify: `web/admin/src/pages/logs/index.tsx`
- Modify: `web/admin/src/pages/articles/index.tsx`
- Optionally modify: `web/admin/src/styles/layout.css`
- Test: `web/admin/src/pages/keywords/index.test.tsx`
- Test: `web/admin/src/pages/tasks/index.test.tsx`
- Test: `web/admin/src/pages/results/index.test.tsx`
- Test: `web/admin/src/pages/logs/index.test.tsx`
- Test: `web/admin/src/pages/articles/index.test.tsx`

**Step 1: Write the minimal implementation**

For all five list pages:
- Remove the `PageHeader` import and JSX.
- Keep the page body starting directly with summary cards, toolbars, or section cards.
- Move list-level actions into `SectionCard.extra` where needed.
- Keep existing filtering and table behavior unchanged.

Representative target pattern:

```tsx
return (
  <>
    <div className="summary-card-grid">...</div>
    <SectionCard title="任务列表" extra={<Button href="/tasks/new">新建任务</Button>}>
      <ToolbarStrip>...</ToolbarStrip>
      <ProTable ... />
    </SectionCard>
  </>
);
```

**Step 2: Run the focused list-page tests**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/keywords/index.test.tsx src/pages/tasks/index.test.tsx src/pages/results/index.test.tsx src/pages/logs/index.test.tsx src/pages/articles/index.test.tsx`
Expected: PASS.

**Step 3: Run a broader list-page regression sweep**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/App.test.tsx src/pages/keywords/index.test.tsx src/pages/tasks/index.test.tsx src/pages/results/index.test.tsx src/pages/logs/index.test.tsx src/pages/articles/index.test.tsx`
Expected: PASS.

**Step 4: Review spacing and copy changes in the diff**

Run: `cd /home/wwwroot/article_sentinel && git diff -- web/admin/src/pages/keywords/index.tsx web/admin/src/pages/tasks/index.tsx web/admin/src/pages/results/index.tsx web/admin/src/pages/logs/index.tsx web/admin/src/pages/articles/index.tsx web/admin/src/styles/layout.css`
Expected: only header removals, lighter copy, and any minimal spacing updates required by the deleted hero area.

**Step 5: Commit**

```bash
git add web/admin/src/pages/keywords/index.tsx web/admin/src/pages/tasks/index.tsx web/admin/src/pages/results/index.tsx web/admin/src/pages/logs/index.tsx web/admin/src/pages/articles/index.tsx web/admin/src/pages/keywords/index.test.tsx web/admin/src/pages/tasks/index.test.tsx web/admin/src/pages/results/index.test.tsx web/admin/src/pages/logs/index.test.tsx web/admin/src/pages/articles/index.test.tsx web/admin/src/styles/layout.css
git commit -m "refactor(admin): compact list page headers"
```

### Task 5: Lock the streamlined detail-header behavior

**Files:**
- Modify: `web/admin/src/pages/tasks/detail.test.tsx`
- Modify: `web/admin/src/pages/articles/detail.test.tsx`
- Modify: `web/admin/src/pages/articles/rectify.test.tsx`
- Reference: `web/admin/src/components/ui/page-header.tsx`

**Step 1: Write the failing test**

Capture that detail pages still keep a single standalone header but no longer show verbose helper copy.

```tsx
expect(screen.getByRole('link', { name: '返回任务列表' })).toHaveAttribute('href', '/tasks');
expect(screen.queryByText('集中查看任务配置、命中结果与关联日志。')).not.toBeInTheDocument();
```

```tsx
expect(screen.getByRole('link', { name: '进入整改' })).toHaveAttribute('href', '/articles/501/rectify');
expect(screen.queryByText('集中查看单篇文稿的命中情况、正文快照、处置记录与整改入口。')).not.toBeInTheDocument();
```

```tsx
expect(screen.getByRole('heading', { name: '内容整改' })).toBeInTheDocument();
expect(screen.queryByText(/围绕当前稿件进行标题、摘要与正文修订。/)).not.toBeInTheDocument();
```

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/tasks/detail.test.tsx src/pages/articles/detail.test.tsx src/pages/articles/rectify.test.tsx`
Expected: FAIL because the detail headers still render the old descriptions.

**Step 3: Note the component API change needed**

Make `PageHeader.description` optional instead of required:

```tsx
export interface PageHeaderProps {
  title: string;
  description?: string;
  extra?: ReactNode;
}
```

**Step 4: Re-run the focused tests before implementation**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/tasks/detail.test.tsx`
Expected: FAIL for the same description assertion.

**Step 5: Commit**

```bash
git add web/admin/src/pages/tasks/detail.test.tsx web/admin/src/pages/articles/detail.test.tsx web/admin/src/pages/articles/rectify.test.tsx
git commit -m "test(admin): capture streamlined detail headers"
```

### Task 6: Trim detail-page descriptions, run the full suite, and build

**Files:**
- Modify: `web/admin/src/components/ui/page-header.tsx`
- Modify: `web/admin/src/pages/tasks/detail.tsx`
- Modify: `web/admin/src/pages/articles/detail.tsx`
- Modify: `web/admin/src/pages/articles/rectify.tsx`
- Optionally modify: `web/admin/src/pages/tasks/new.tsx`
- Test: `web/admin/src/pages/tasks/detail.test.tsx`
- Test: `web/admin/src/pages/articles/detail.test.tsx`
- Test: `web/admin/src/pages/articles/rectify.test.tsx`

**Step 1: Write the minimal implementation**

Apply the smallest header cleanup that keeps detail pages complete but quieter:

```tsx
<PageHeader
  title="任务详情"
  extra={
    <Space wrap>
      <Button href="/tasks">返回任务列表</Button>
      <Button type="primary" href="/results">查看风险结果</Button>
    </Space>
  }
/>
```

```tsx
export function PageHeader({ title, description, extra }: PageHeaderProps) {
  return (
    <div className="page-header">
      <div>
        <h2>{title}</h2>
        {description ? <p className="page-header__description">{description}</p> : null}
      </div>
      {extra ? <div className="page-header__extra">{extra}</div> : null}
    </div>
  );
}
```

Implementation notes:
- Remove the `description` prop from `tasks/detail.tsx`, `articles/detail.tsx`, and `articles/rectify.tsx`.
- Leave `tasks/new.tsx` unchanged unless the resulting UI looks inconsistent during review.

**Step 2: Run the focused detail tests**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/tasks/detail.test.tsx src/pages/articles/detail.test.tsx src/pages/articles/rectify.test.tsx`
Expected: PASS.

**Step 3: Run the complete frontend test suite**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test`
Expected: PASS for the entire Vitest suite.

**Step 4: Run the production build**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm run build`
Expected: PASS with a successful Vite production build.

**Step 5: Commit**

```bash
git add web/admin/src/components/ui/page-header.tsx web/admin/src/pages/tasks/detail.tsx web/admin/src/pages/articles/detail.tsx web/admin/src/pages/articles/rectify.tsx web/admin/src/pages/tasks/detail.test.tsx web/admin/src/pages/articles/detail.test.tsx web/admin/src/pages/articles/rectify.test.tsx
git commit -m "refactor(admin): streamline detail page headers"
```

# Admin UI Shadcn Rewrite Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rebuild `web/admin` so the existing inspection console keeps the four current primary routes, adds article list/detail pages, and fully adopts a `shadcn-admin`-style dashboard shell without changing backend APIs.

**Architecture:** Replace the current hero-like shell with a compact dashboard frame, create a shared set of list/detail UI primitives, and recompose each page around those primitives. Use existing inspection result and log APIs to derive article-centric list/detail screens, so all new behavior stays in the presentation layer.

**Tech Stack:** React 18, TypeScript, React Router 7, Ant Design, custom CSS tokens/layouts, Vitest, Testing Library.

---

### Task 1: Rewrite the shell foundation to remove the oversized page header

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/App.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/routes.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/layout/admin-shell.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/layout/sidebar-nav.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/layout/topbar.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/theme.css`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing shell test**

Update `web/admin/src/App.test.tsx` to assert:

```tsx
expect(screen.getByRole('heading', { name: '关键词规则' })).toBeInTheDocument();
expect(screen.queryByText('适用机构')).not.toBeInTheDocument();
expect(screen.queryByText('巡检时段')).not.toBeInTheDocument();
expect(screen.queryByText('提示状态')).not.toBeInTheDocument();
```

**Step 2: Run the test and verify RED**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/App.test.tsx`
Expected: FAIL because the current topbar still renders the large hero metadata blocks.

**Step 3: Implement the compact dashboard shell**

Refactor the shell toward this structure:

```tsx
<div className="admin-shell">
  <SidebarNav />
  <div className="admin-shell__main">
    <Topbar breadcrumb={...} actions={...} />
    <main className="admin-shell__content">{children}</main>
  </div>
</div>
```

And replace the current large-card topbar with a compact title/action row.

**Step 4: Run the same test and verify GREEN**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/App.test.tsx`
Expected: PASS.

**Step 5: Refactor CSS tokens**

Normalize tokens around light gray surfaces, subtle borders, compact spacing, and muted blue accents. Keep class names stable where possible.

### Task 2: Add article routes and article-centric route metadata

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/routes.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/App.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/index.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/detail.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/index.test.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/detail.test.tsx`

**Step 1: Write the failing article route tests**

Add tests that expect:

```tsx
render(<MemoryRouter initialEntries={['/articles']}><App /></MemoryRouter>);
expect(screen.getByRole('heading', { name: '文稿列表' })).toBeInTheDocument();

render(<MemoryRouter initialEntries={['/articles/501']}><App /></MemoryRouter>);
expect(screen.getByRole('heading', { name: '文稿详情' })).toBeInTheDocument();
```

**Step 2: Run article route tests and verify RED**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/articles/index.test.tsx src/pages/articles/detail.test.tsx`
Expected: FAIL because the routes and pages do not exist.

**Step 3: Add the routes and metadata**

Register:

```tsx
<Route path="/articles" element={<ArticlesPage />} />
<Route path="/articles/:articleId" element={<ArticleDetailPage />} />
<Route path="/articles/:articleId/rectify" element={<RectifyPage />} />
```

Keep these routes outside the primary sidebar nav while still resolving correct breadcrumb/title text.

**Step 4: Re-run the route tests and verify GREEN**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/articles/index.test.tsx src/pages/articles/detail.test.tsx`
Expected: PASS.

### Task 3: Build shared list/detail primitives for shadcn-like admin pages

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/components.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/page-header.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/section-card.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/toolbar-strip.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/filter-bar.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/detail-metric.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/detail-section.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write failing primitive tests**

Extend `components.test.tsx` with assertions for a compact filter bar and detail metric rendering:

```tsx
expect(screen.getByText('筛选条件')).toBeInTheDocument();
expect(screen.getByText('文稿编号')).toBeInTheDocument();
expect(screen.getByText('命中次数')).toBeInTheDocument();
```

**Step 2: Run the primitive test and verify RED**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/components/ui/components.test.tsx`
Expected: FAIL because the new primitives do not exist.

**Step 3: Implement compact primitives**

Create shared components following this pattern:

```tsx
<SectionCard title="结果列表" description="按筛选条件查看当前命中记录">
  <FilterBar>{filters}</FilterBar>
  <ToolbarStrip>{actions}</ToolbarStrip>
</SectionCard>
```

And for detail summaries:

```tsx
<div className="detail-metrics-grid">
  <DetailMetric label="文稿编号" value="#501" />
  <DetailMetric label="命中次数" value={3} />
</div>
```

**Step 4: Re-run the primitive test and verify GREEN**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/components/ui/components.test.tsx`
Expected: PASS.

### Task 4: Rework the results page into a compact workbench and route to full article detail

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/index.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/detail.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/detail.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write failing results-page expectations**

Update the list test to assert article-title navigation instead of drawer usage:

```tsx
expect(screen.getByRole('link', { name: 'Spam alert' })).toHaveAttribute('href', '/articles/501');
expect(screen.queryByText('结果详情')).not.toBeInTheDocument();
```

Update the detail test to cover the extracted detail content renderer instead of a `Drawer` shell.

**Step 2: Run result tests and verify RED**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/results/index.test.tsx src/pages/results/detail.test.tsx`
Expected: FAIL because the page still opens a drawer.

**Step 3: Implement the new results workbench**

- Replace the page header with a compact header + filter row + action row.
- Make article titles and detail actions navigate to `/articles/:articleId`.
- Keep batch offline behavior unchanged.
- Convert `results/detail.tsx` into a reusable content section or loader that can be embedded in `ArticleDetailPage`.

**Step 4: Re-run result tests and verify GREEN**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/results/index.test.tsx src/pages/results/detail.test.tsx`
Expected: PASS.

### Task 5: Implement the article list page from existing result data

**Files:**
- Create: `/home/wwwroot/article_sentinel/web/admin/src/services/articles.ts`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/services/articles.test.ts`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/index.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing article-list tests**

Create tests that expect article-centric rows derived from repeated result records:

```tsx
expect(await screen.findByText('Spam alert')).toBeInTheDocument();
expect(screen.getByText('#501')).toBeInTheDocument();
expect(screen.getByRole('link', { name: '查看详情' })).toHaveAttribute('href', '/articles/501');
```

Also add a service test that collapses result rows into one article summary.

**Step 2: Run tests and verify RED**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/services/articles.test.ts src/pages/articles/index.test.tsx`
Expected: FAIL because there is no article aggregation layer.

**Step 3: Implement the article aggregation layer**

Create a small adapter around `listResults`:

```ts
export function summarizeArticles(items: ResultRecord[]): ArticleListItem[] {
  // group by article_id and keep highest risk / latest task / total hits
}
```

Render the new `文稿列表` page using that adapter plus existing result filters.

**Step 4: Re-run tests and verify GREEN**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/services/articles.test.ts src/pages/articles/index.test.tsx`
Expected: PASS.

### Task 6: Implement the full article detail page

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/services/results.ts`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/services/logs.ts`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/detail.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/detail.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write failing article-detail tests**

Add tests that expect:

```tsx
expect(await screen.findByText('Spam alert')).toBeInTheDocument();
expect(screen.getByText('文稿编号')).toBeInTheDocument();
expect(screen.getByRole('tab', { name: '命中记录' })).toBeInTheDocument();
expect(screen.getByRole('link', { name: '进入整改' })).toHaveAttribute('href', '/articles/501/rectify');
```

**Step 2: Run the detail test and verify RED**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/articles/detail.test.tsx`
Expected: FAIL because the detail page does not exist.

**Step 3: Build the full detail page**

Use this loading flow:

```ts
const resultList = await listResults({ orgid: 100, article_id });
const primary = resultList.items[0];
const [detail, logs, changes] = await Promise.all([
  getResultDetail(primary.id, 100),
  listArticleOperationLogs(article_id, 100),
  listArticleFieldChanges(article_id, 100),
]);
```

Then render:

- top action bar
- detail metrics grid
- left/right summary sections
- tabbed content for hits, operation logs, field changes, and body snapshot

**Step 4: Re-run the detail test and verify GREEN**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/articles/detail.test.tsx`
Expected: PASS.

### Task 7: Align the remaining pages with the new shell and article workflow

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/keywords/index.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/keywords/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/new.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/new.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/rectify.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/rectify.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write failing page-level assertions**

Add assertions that each page uses compact headers and article-detail navigation where relevant.

**Step 2: Run the page tests and verify RED**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/keywords/index.test.tsx src/pages/tasks/index.test.tsx src/pages/tasks/new.test.tsx src/pages/logs/index.test.tsx src/pages/articles/rectify.test.tsx`
Expected: FAIL where pages still rely on the old spacing or lack the new links.

**Step 3: Restyle and reconnect the pages**

- Keywords: compact list page + consistent modal spacing
- Tasks: compact metrics + cleaner create layout
- Logs: filter bar + article detail links
- Rectify: return-to-detail action + new two-column layout polish

**Step 4: Re-run the page tests and verify GREEN**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/keywords/index.test.tsx src/pages/tasks/index.test.tsx src/pages/tasks/new.test.tsx src/pages/logs/index.test.tsx src/pages/articles/rectify.test.tsx`
Expected: PASS.

### Task 8: Run full verification

**Files:**
- Reference: `/home/wwwroot/article_sentinel/web/admin/src/**`

**Step 1: Run the full admin test suite**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test`
Expected: PASS with zero failing tests.

**Step 2: Run the production build**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm run build`
Expected: exit code 0.

**Step 3: Review the diff for unintended shell regressions**

Run: `cd /home/wwwroot/article_sentinel && git diff -- web/admin/src docs/plans/2026-04-23-admin-ui-shadcn-rewrite*`
Expected: only the intended shell, page, test, and plan changes.

# Admin Shell Cleanup and Task-Centric Inspection Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Remove noisy admin shell controls, make every workspace tab closable, and move inspection result operations back into task detail so operators work from `检测任务 -> 任务详情 -> 命中结果`.

**Architecture:** Keep `BasicLayout` as the root shell, but simplify its header, remove standalone `/inspection/results` from the active workspace model, and drive tab/menu behavior from `route-meta.ts`. Upgrade `TaskDetailPage` so its `命中结果` tab reuses the result-table workflow directly inside the detail route, while compatibility redirects funnel old result URLs into `任务详情` with query-driven tab state.

**Tech Stack:** Umi Max, React 18, Ant Design 5, ProLayout, ProComponents, Vitest, Testing Library, Less

---

### Task 1: Lock the new shell and tab contracts with failing tests

**Files:**
- Modify: `web/admin/src/layouts/BasicLayout.test.tsx`
- Modify: `web/admin/src/components/PageTabs/route-meta.test.ts`
- Modify: `web/admin/src/components/PageTabs/store.test.ts`
- Modify: `web/admin/src/components/PageTabs/index.test.tsx`

**Step 1: Write the failing test**

Update `web/admin/src/layouts/BasicLayout.test.tsx` so it asserts the cleaned shell behavior:

```tsx
expect(screen.queryByRole('button', { name: '收缩侧边栏' })).not.toBeInTheDocument();
expect(screen.queryByLabelText('搜索入口')).not.toBeInTheDocument();
expect(screen.queryByRole('button', { name: '切换全屏' })).not.toBeInTheDocument();
expect(screen.queryByRole('button', { name: '通知中心' })).not.toBeInTheDocument();
await user.click(screen.getByRole('button', { name: '用户菜单' }));
expect(screen.queryByText('个人中心')).not.toBeInTheDocument();
expect(screen.getByText('退出登录')).toBeInTheDocument();
```

Update `web/admin/src/components/PageTabs/route-meta.test.ts` to assert:

```ts
expect(resolveTabDescriptor('/inspection/tasks').closable).toBe(true);
expect(resolveRouteMeta('/inspection/results').hiddenInMenu).toBe(true);
expect(resolveRouteMeta('/inspection/results').opensTab).toBe(false);
```

Update `web/admin/src/components/PageTabs/store.test.ts` so the close-path uses `/inspection/tasks` as a normal closable tab and verifies that closing it is allowed.

Update `web/admin/src/components/PageTabs/index.test.tsx` so the fixture renders a close button for `检测任务` and asserts `onClose('/inspection/tasks')` is fired.

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/layouts/BasicLayout.test.tsx src/components/PageTabs/route-meta.test.ts src/components/PageTabs/store.test.ts src/components/PageTabs/index.test.tsx`
Expected: FAIL because the shell still renders the old controls and `/inspection/tasks` is still non-closable in route metadata.

**Step 3: Write minimal implementation**

Do not change production code yet beyond the minimum needed to satisfy the tests conceptually:
- mark `/inspection/tasks` closable in `route-meta.ts`
- mark `/inspection/results` as `hiddenInMenu: true` and `opensTab: false`
- keep the tab rail test fixtures aligned with the new all-closable expectation

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/layouts/BasicLayout.test.tsx src/components/PageTabs/route-meta.test.ts src/components/PageTabs/store.test.ts src/components/PageTabs/index.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add web/admin/src/layouts/BasicLayout.test.tsx \
  web/admin/src/components/PageTabs/route-meta.test.ts \
  web/admin/src/components/PageTabs/store.test.ts \
  web/admin/src/components/PageTabs/index.test.tsx \
  web/admin/src/components/PageTabs/route-meta.ts
git commit -m "test(admin): lock cleaned shell and closable tabs"
```

### Task 2: Implement the cleaned shell header and remove unused controls

**Files:**
- Modify: `web/admin/src/layouts/BasicLayout.tsx`
- Modify: `web/admin/src/global.less`

**Step 1: Write the failing test**

Use the shell test from Task 1 as the red test. It must still be failing before production edits if not already green for the right reason.

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/layouts/BasicLayout.test.tsx`
Expected: FAIL because the old buttons and menu items are still rendered.

**Step 3: Write minimal implementation**

Refactor `web/admin/src/layouts/BasicLayout.tsx` to remove these imports/usages:

- `Badge`
- `Breadcrumb`
- `Input`
- `BellOutlined`
- `FullscreenOutlined`
- `FullscreenExitOutlined`
- `MenuFoldOutlined`
- `MenuUnfoldOutlined`
- `SearchOutlined`

Keep the custom shell wrapper and tab handling, but simplify the header to:

```tsx
<header className="admin-header admin-light-surface" data-testid="admin-header">
  <div className="admin-header__left">
    <Typography.Title level={5} className="admin-header__title">
      {resolveRouteMeta(location.pathname).name}
    </Typography.Title>
  </div>
  <div className="admin-header__right">
    <Dropdown ...>
      <Button type="text" className="admin-user-menu" aria-label="用户菜单">...</Button>
    </Dropdown>
  </div>
</header>
```

Update the dropdown items to only contain:

```ts
[{ key: 'logout', label: '退出登录' }]
```

Set `collapsedButtonRender={false}` on `ProLayout` so the built-in sider toggle disappears.

Clean up `web/admin/src/global.less` so removed selectors no longer reserve space for search/action controls, and add a small `.admin-header__title` style for the simplified left area.

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/layouts/BasicLayout.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add web/admin/src/layouts/BasicLayout.tsx \
  web/admin/src/global.less
git commit -m "feat(admin): simplify workspace header"
```

### Task 3: Remove standalone result workspace routing and redirect old URLs into task detail

**Files:**
- Modify: `web/admin/src/pages/LegacyRedirect/index.test.tsx`
- Modify: `web/admin/src/pages/LegacyRedirect/index.tsx`
- Modify: `web/admin/src/components/PageTabs/route-meta.ts`
- Modify: `web/admin/config/routes.ts`

**Step 1: Write the failing test**

Update `web/admin/src/pages/LegacyRedirect/index.test.tsx` expectations to:

```ts
expect(resolveLegacyPath('/tasks/77/results')).toBe('/inspection/tasks/77?tab=results');
expect(resolveLegacyPath('/results?task_id=77&page=2')).toBe('/inspection/tasks/77?tab=results&page=2');
expect(resolveLegacyPath('/results')).toBe('/inspection/tasks');
```

Also add a route-meta expectation that `/inspection/results` is hidden and does not open a tab.

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/LegacyRedirect/index.test.tsx src/components/PageTabs/route-meta.test.ts`
Expected: FAIL because old result routes still resolve to `/inspection/results`.

**Step 3: Write minimal implementation**

In `web/admin/src/pages/LegacyRedirect/index.tsx`:
- map `/results` with `task_id` to `/inspection/tasks/:taskId?tab=results`
- map `/results` without task context to `/inspection/tasks`
- map `/tasks/:taskId/results` to `/inspection/tasks/:taskId?tab=results`

In `web/admin/src/components/PageTabs/route-meta.ts`:
- keep `/inspection/results` as a compatibility route only
- set `hiddenInMenu: true`
- set `opensTab: false`
- set `menuKey: '/inspection/tasks'`

Leave `config/routes.ts` route registration in place only so existing URLs still hit `LegacyRedirect` or the hidden compatibility route cleanly.

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/LegacyRedirect/index.test.tsx src/components/PageTabs/route-meta.test.ts`
Expected: PASS

**Step 5: Commit**

```bash
git add web/admin/src/pages/LegacyRedirect/index.test.tsx \
  web/admin/src/pages/LegacyRedirect/index.tsx \
  web/admin/src/components/PageTabs/route-meta.ts \
  web/admin/config/routes.ts
git commit -m "refactor(admin): route result links into task detail"
```

### Task 4: Restore task-detail entry from the task list

**Files:**
- Modify: `web/admin/src/pages/Inspection/TaskList/index.test.tsx`
- Modify: `web/admin/src/pages/Inspection/TaskList/index.tsx`

**Step 1: Write the failing test**

Extend `web/admin/src/pages/Inspection/TaskList/index.test.tsx` with a populated-row scenario:

```tsx
mockedListTasks.mockResolvedValue({
  page: 1,
  pageSize: 20,
  total: 1,
  items: [{
    id: 77,
    orgid: 29,
    task_no: 'inspect-20260420-01',
    status: 'success',
    total_scanned: 42,
    hit_count: 8,
    creator_name: 'operator',
    created_at: '2026-04-20 12:00:00'
  }]
});

expect(await screen.findByRole('button', { name: 'inspect-20260420-01' })).toBeInTheDocument();
expect(screen.getByRole('button', { name: '查看任务' })).toBeInTheDocument();
```

Click each action and assert navigation to `/inspection/tasks/77`.

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Inspection/TaskList/index.test.tsx`
Expected: FAIL because the task number is plain text and there is no `查看任务` action.

**Step 3: Write minimal implementation**

In `web/admin/src/pages/Inspection/TaskList/index.tsx`:
- render `task_no` as a link-style button to `/inspection/tasks/${record.id}`
- add `查看任务` to the operation column for every row
- keep delete logic for `pending` and `failed`
- preserve existing summary/filter/table layout

Example shape:

```tsx
{
  title: '任务编号',
  dataIndex: 'task_no',
  render: (_, record) => (
    <Button type="link" onClick={() => navigate(`/inspection/tasks/${record.id}`)}>
      {record.task_no}
    </Button>
  )
}
```

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Inspection/TaskList/index.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add web/admin/src/pages/Inspection/TaskList/index.test.tsx \
  web/admin/src/pages/Inspection/TaskList/index.tsx
git commit -m "feat(admin): restore task detail entry points"
```

### Task 5: Convert task detail results from preview list to full result workspace

**Files:**
- Modify: `web/admin/src/pages/Inspection/TaskDetail/index.test.tsx`
- Modify: `web/admin/src/pages/Inspection/TaskDetail/index.tsx`
- Modify: `web/admin/src/global.less`

**Step 1: Write the failing test**

Rewrite `web/admin/src/pages/Inspection/TaskDetail/index.test.tsx` so it expects the `命中结果` tab to render the full result-workspace controls:

```tsx
expect(await screen.findByRole('button', { name: '批量下线处置' })).toBeInTheDocument();
expect(screen.getByText('已选 0 项')).toBeInTheDocument();
expect(screen.getByRole('button', { name: '查看详情' })).toBeInTheDocument();
expect(screen.getByRole('button', { name: '进入整改' })).toBeInTheDocument();
```

Add a case starting from `/inspection/tasks/77?tab=results&page=2` and assert:

```ts
expect(mockedListResults).toHaveBeenLastCalledWith(expect.objectContaining({
  task_id: 77,
  page: 2,
  pageSize: 20
}));
```

Add a navigation assertion for `查看详情`:

```ts
/content/articles/501?return_to=%2Finspection%2Ftasks%2F77%3Ftab%3Dresults
```

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Inspection/TaskDetail/index.test.tsx`
Expected: FAIL because the page still renders a simple `List` in the results tab and ignores `tab/page` query state.

**Step 3: Write minimal implementation**

In `web/admin/src/pages/Inspection/TaskDetail/index.tsx`:
- replace the `List`-based `results` tab body with a table-first panel modeled on the current `ResultListPage`
- use `useSearchParams()` to read and write:
  - `tab`
  - `page`
- default to `results` tab
- call `listResults({ task_id: numericTaskId, page, pageSize: 20 })`
- keep local row selection and offline confirmation modal
- keep existing summary/snapshot/log sections intact
- use `return_to=${encodeURIComponent(currentHref)}` for article-detail and rectify navigation

If a small internal helper is needed, keep it inside this page for now rather than over-abstracting.

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Inspection/TaskDetail/index.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add web/admin/src/pages/Inspection/TaskDetail/index.test.tsx \
  web/admin/src/pages/Inspection/TaskDetail/index.tsx \
  web/admin/src/global.less
git commit -m "feat(admin): embed task results in task detail"
```

### Task 6: Repoint audit and article return paths to the task-centric workspace

**Files:**
- Modify: `web/admin/src/pages/Audit/OperationLogList/index.test.tsx`
- Modify: `web/admin/src/pages/Audit/OperationLogList/index.tsx`
- Modify: `web/admin/src/pages/Content/ArticleDetail/index.test.tsx`

**Step 1: Write the failing test**

Update `web/admin/src/pages/Audit/OperationLogList/index.test.tsx` so task links route to:

```ts
/inspection/tasks/77?tab=results
```

Update `web/admin/src/pages/Content/ArticleDetail/index.test.tsx` return-path case to use:

```ts
/content/articles/501?return_to=%2Finspection%2Ftasks%2F77%3Ftab%3Dresults%26page%3D2
```

and assert the return button navigates back to `/inspection/tasks/77?tab=results&page=2`.

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Audit/OperationLogList/index.test.tsx src/pages/Content/ArticleDetail/index.test.tsx`
Expected: FAIL because links and return targets still point to `/inspection/results`.

**Step 3: Write minimal implementation**

In `web/admin/src/pages/Audit/OperationLogList/index.tsx`:

```tsx
<Button type="link" onClick={() => navigate(`/inspection/tasks/${record.task_id}?tab=results`)}>
  #{record.task_id}
</Button>
```

Do not change article link behavior other than preserving the correct `return_to` based on current URL.

Adjust any article-detail tests only for the updated `return_to` semantics; the page implementation may already be generic enough once callers pass the new URL.

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Audit/OperationLogList/index.test.tsx src/pages/Content/ArticleDetail/index.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add web/admin/src/pages/Audit/OperationLogList/index.test.tsx \
  web/admin/src/pages/Audit/OperationLogList/index.tsx \
  web/admin/src/pages/Content/ArticleDetail/index.test.tsx
git commit -m "refactor(admin): keep audit navigation in task workspaces"
```

### Task 7: Remove residual Pro page-header chrome from business pages

**Files:**
- Modify: `web/admin/src/pages/Inspection/TaskList/index.tsx`
- Modify: `web/admin/src/pages/Inspection/TaskDetail/index.tsx`
- Modify: `web/admin/src/pages/Inspection/TaskCreate/index.tsx`
- Modify: `web/admin/src/pages/Inspection/ResultList/index.tsx`
- Modify: `web/admin/src/pages/Rules/CategoryList/index.tsx`
- Modify: `web/admin/src/pages/Rules/KeywordList/index.tsx`
- Modify: `web/admin/src/pages/Content/ArticleList/index.tsx`
- Modify: `web/admin/src/pages/Content/ArticleDetail/index.tsx`
- Modify: `web/admin/src/pages/Content/ArticleRectify/index.tsx`
- Modify: `web/admin/src/pages/Audit/OperationLogList/index.tsx`
- Modify: `web/admin/src/global.less`

**Step 1: Write the failing test**

Add or extend an existing shell/page contract test so it asserts the page-header wrapper is not rendered, for example:

```tsx
expect(document.querySelector('.ant-pro-page-container-warp-page-header')).not.toBeInTheDocument();
expect(document.querySelector('.ant-page-header')).not.toBeInTheDocument();
```

Use `BasicLayout.test.tsx` or a focused page test that renders a full page under `PageContainer`.

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/layouts/BasicLayout.test.tsx src/pages/Inspection/TaskList/index.test.tsx`
Expected: FAIL because `PageContainer` still contributes Pro page-header markup.

**Step 3: Write minimal implementation**

Add `pageHeaderRender={false}` to each business `PageContainer` call:

```tsx
<PageContainer title={false} pageHeaderRender={false}>
```

Add a light defensive CSS fallback in `web/admin/src/global.less` only if needed for remaining wrapper spacing.

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/layouts/BasicLayout.test.tsx src/pages/Inspection/TaskList/index.test.tsx`
Expected: PASS

**Step 5: Commit**

```bash
git add web/admin/src/pages/Inspection/TaskList/index.tsx \
  web/admin/src/pages/Inspection/TaskDetail/index.tsx \
  web/admin/src/pages/Inspection/TaskCreate/index.tsx \
  web/admin/src/pages/Inspection/ResultList/index.tsx \
  web/admin/src/pages/Rules/CategoryList/index.tsx \
  web/admin/src/pages/Rules/KeywordList/index.tsx \
  web/admin/src/pages/Content/ArticleList/index.tsx \
  web/admin/src/pages/Content/ArticleDetail/index.tsx \
  web/admin/src/pages/Content/ArticleRectify/index.tsx \
  web/admin/src/pages/Audit/OperationLogList/index.tsx \
  web/admin/src/global.less
git commit -m "style(admin): remove residual page header chrome"
```

### Task 8: Run focused verification and update docs only if behavior changed during implementation

**Files:**
- Modify if needed: `docs/article-inspection-pages.md`
- Modify if needed: `docs/plans/2026-05-11-admin-header-tabs-design.md`
- Modify if needed: `docs/plans/2026-05-11-admin-header-tabs-implementation-plan.md`

**Step 1: Write the verification checklist**

No new test file is required. Prepare the final command list:

```bash
cd /home/wwwroot/article_sentinel/web/admin && npm test -- \
  src/layouts/BasicLayout.test.tsx \
  src/components/PageTabs/route-meta.test.ts \
  src/components/PageTabs/store.test.ts \
  src/components/PageTabs/index.test.tsx \
  src/pages/LegacyRedirect/index.test.tsx \
  src/pages/Inspection/TaskList/index.test.tsx \
  src/pages/Inspection/TaskDetail/index.test.tsx \
  src/pages/Audit/OperationLogList/index.test.tsx \
  src/pages/Content/ArticleDetail/index.test.tsx
```

**Step 2: Run verification**

Run the command above.
Expected: PASS for the focused shell/workflow suite.

**Step 3: Fix anything still failing**

If any failures remain:
- update the minimal production code
- re-run only the failing tests
- re-run the full focused suite

Keep changes YAGNI and scoped to the verified behavior.

**Step 4: Update docs if needed**

If the final behavior differs from `docs/article-inspection-pages.md`, update the route and workflow description so it matches:
- task list enters task detail
- task detail contains result operations
- `/inspection/results` is compatibility-only rather than a standalone operator page

**Step 5: Commit**

```bash
git add docs/article-inspection-pages.md \
  docs/plans/2026-05-11-admin-header-tabs-design.md \
  docs/plans/2026-05-11-admin-header-tabs-implementation-plan.md

git commit -m "docs(admin): align inspection workspace docs"
```

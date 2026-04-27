# Admin Tabbed Workbench Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add an org-scoped tabbed workbench to the admin frontend so sidebar and in-page navigation open pages without full-page reloads while preserving key page state.

**Architecture:** Add a React workbench layer above the route outlet that owns tab descriptors, activation, closing, session persistence, and org isolation. Keep React Router as the URL source of truth, define a route-to-tab registry for single-instance list pages and multi-instance detail pages, then migrate page navigation to a shared workbench navigation API and move critical list state into query strings.

**Tech Stack:** React 18, React Router DOM 7, TypeScript, Ant Design, Vitest, Testing Library

---

### Task 1: Lock workbench route and reducer behavior with tests

**Files:**
- Create: `web/admin/src/workbench/registry.test.ts`
- Create: `web/admin/src/workbench/store.test.ts`
- Modify: `web/admin/src/routes.tsx`

**Step 1:** Write failing registry tests for these cases:

```ts
expect(resolveWorkbenchRoute('/articles')).toMatchObject({
  kind: 'list',
  key: '/articles',
  reusable: true,
  title: '文稿中心'
});

expect(resolveWorkbenchRoute('/articles/123')).toMatchObject({
  kind: 'detail',
  key: 'article:123',
  reusable: true,
  title: '文稿#123'
});

expect(resolveWorkbenchRoute('/tasks/88/results')).toMatchObject({
  kind: 'detail',
  key: 'task:88:results',
  title: '任务结果#88'
});
```

**Step 2:** Write failing reducer/store tests for these behaviors:

```ts
openTab('/articles');
openTab('/articles');
expect(state.tabs).toHaveLength(1);

openTab('/articles/123');
openTab('/articles/456');
expect(state.tabs.map((tab) => tab.key)).toEqual(['/tasks', 'article:123', 'article:456']);

closeAllTabs();
expect(state.activeKey).toBe('/tasks');
```

**Step 3:** Run `npm --prefix web/admin test -- src/workbench/registry.test.ts src/workbench/store.test.ts` and confirm the new tests fail because the workbench layer does not exist yet.

**Step 4:** Commit only the test scaffolding.

```bash
git add web/admin/src/workbench/registry.test.ts web/admin/src/workbench/store.test.ts web/admin/src/routes.tsx
git commit -m "test: lock admin workbench tab rules"
```

### Task 2: Implement the workbench registry and state store

**Files:**
- Create: `web/admin/src/workbench/types.ts`
- Create: `web/admin/src/workbench/registry.ts`
- Create: `web/admin/src/workbench/store.tsx`
- Create: `web/admin/src/workbench/session.ts`
- Modify: `web/admin/src/routes.tsx`

**Step 1:** Add exact workbench types for route descriptors, tab records, persisted session data, and close actions.

```ts
export type WorkbenchTab = {
  key: string;
  pathname: string;
  search: string;
  title: string;
  closable: boolean;
  keepAlive: boolean;
  orgId: number;
};
```

**Step 2:** Implement a route registry that resolves every supported pathname into a descriptor containing:
- single-instance vs multi-instance policy
- initial title
- fallback route
- cache eligibility
- async title update support

**Step 3:** Implement a reducer-driven store that supports:
- `openTab`
- `activateTab`
- `closeTab`
- `closeOtherTabs`
- `closeTabsToLeft`
- `closeTabsToRight`
- `closeAllTabs`
- `replaceTabTitle`
- `restoreSession`

**Step 4:** Persist one session snapshot per org in `sessionStorage`, keyed like `admin-workbench:<orgId>`.

**Step 5:** Re-run `npm --prefix web/admin test -- src/workbench/registry.test.ts src/workbench/store.test.ts` and confirm green.

**Step 6:** Commit the registry and store.

```bash
git add web/admin/src/workbench/types.ts web/admin/src/workbench/registry.ts web/admin/src/workbench/store.tsx web/admin/src/workbench/session.ts web/admin/src/routes.tsx
git commit -m "feat: add admin workbench store"
```

### Task 3: Lock shell, tabs UI, and org restore behavior with tests

**Files:**
- Modify: `web/admin/src/App.test.tsx`
- Create: `web/admin/src/workbench/tabs.test.tsx`
- Create: `web/admin/src/workbench/provider.test.tsx`

**Step 1:** Extend `web/admin/src/App.test.tsx` with failing expectations for:
- a visible tab strip under the header
- a default `检测任务` base tab
- no top-level full-page reload behavior for sidebar page switches

**Step 2:** Add `tabs.test.tsx` coverage for:
- rendering a tab per open descriptor
- activating an existing tab without duplication
- right-side action menu items: `关闭当前` / `关闭其他` / `关闭左侧` / `关闭右侧` / `关闭全部`

**Step 3:** Add `provider.test.tsx` coverage for:
- restoring tabs from `sessionStorage`
- isolating session snapshots by org
- falling back to the base tab when an org has no snapshot

**Step 4:** Run `npm --prefix web/admin test -- src/App.test.tsx src/workbench/tabs.test.tsx src/workbench/provider.test.tsx` and confirm failure.

**Step 5:** Commit the failing UI tests.

```bash
git add web/admin/src/App.test.tsx web/admin/src/workbench/tabs.test.tsx web/admin/src/workbench/provider.test.tsx
git commit -m "test: cover admin workbench shell"
```

### Task 4: Implement the shell workbench provider and tabs UI

**Files:**
- Create: `web/admin/src/workbench/provider.tsx`
- Create: `web/admin/src/workbench/tabs.tsx`
- Create: `web/admin/src/workbench/use-workbench.ts`
- Modify: `web/admin/src/App.tsx`
- Modify: `web/admin/src/components/layout/admin-shell.tsx`
- Modify: `web/admin/src/components/layout/header-bar.tsx`
- Modify: `web/admin/src/styles/layout.css`
- Modify: `web/admin/src/styles.css`

**Step 1:** Wrap the admin app with a `WorkbenchProvider` inside `OrgProvider` so the workbench can react to active-org changes.

**Step 2:** Render a tab strip between `HeaderBar` and the content frame, using Ant Design tabs or a custom horizontal tab row with a right-side dropdown menu.

**Step 3:** Ensure the base `检测任务` tab is always present and non-closable.

**Step 4:** On org switch, load the matching org snapshot and replace the current tab set without tearing down the global shell.

**Step 5:** Style the tab strip so it fits the current neutral admin shell, supports horizontal overflow, and keeps the action menu pinned on the right.

**Step 6:** Re-run `npm --prefix web/admin test -- src/App.test.tsx src/workbench/tabs.test.tsx src/workbench/provider.test.tsx` and confirm green.

**Step 7:** Commit the shell integration.

```bash
git add web/admin/src/workbench/provider.tsx web/admin/src/workbench/tabs.tsx web/admin/src/workbench/use-workbench.ts web/admin/src/App.tsx web/admin/src/components/layout/admin-shell.tsx web/admin/src/components/layout/header-bar.tsx web/admin/src/styles/layout.css web/admin/src/styles.css
git commit -m "feat: add admin workbench shell tabs"
```

### Task 5: Lock navigation integration before replacing direct links

**Files:**
- Modify: `web/admin/src/pages/articles/index.test.tsx`
- Modify: `web/admin/src/pages/articles/detail.test.tsx`
- Modify: `web/admin/src/pages/articles/rectify.test.tsx`
- Modify: `web/admin/src/pages/tasks/index.test.tsx`
- Modify: `web/admin/src/pages/tasks/detail.test.tsx`
- Modify: `web/admin/src/pages/tasks/results.test.tsx`
- Modify: `web/admin/src/pages/logs/index.test.tsx`
- Modify: `web/admin/src/pages/categories/index.test.tsx`

**Step 1:** Replace expectations that currently assert raw `href` navigation with workbench-aware expectations, for example:

```ts
await user.click(screen.getByRole('button', { name: '查看详情' }));
expect(mockedOpenWorkbenchTab).toHaveBeenCalledWith('/articles/501', {
  returnTo: '/articles'
});
```

**Step 2:** Add failing coverage for:
- sidebar navigation reusing list tabs
- detail pages opening in new tabs
- result/rectify actions opening the correct resource routes
- history-safe return targets preserved in query strings where required

**Step 3:** Run `npm --prefix web/admin test -- src/pages/articles/index.test.tsx src/pages/articles/detail.test.tsx src/pages/articles/rectify.test.tsx src/pages/tasks/index.test.tsx src/pages/tasks/detail.test.tsx src/pages/tasks/results.test.tsx src/pages/logs/index.test.tsx src/pages/categories/index.test.tsx` and confirm failure.

**Step 4:** Commit the navigation-focused test updates.

```bash
git add web/admin/src/pages/articles/index.test.tsx web/admin/src/pages/articles/detail.test.tsx web/admin/src/pages/articles/rectify.test.tsx web/admin/src/pages/tasks/index.test.tsx web/admin/src/pages/tasks/detail.test.tsx web/admin/src/pages/tasks/results.test.tsx web/admin/src/pages/logs/index.test.tsx web/admin/src/pages/categories/index.test.tsx
git commit -m "test: lock admin workbench navigation"
```

### Task 6: Implement the shared workbench navigation API and migrate hotspots

**Files:**
- Create: `web/admin/src/workbench/navigation.ts`
- Create: `web/admin/src/workbench/link.tsx`
- Modify: `web/admin/src/components/layout/sidebar-nav.tsx`
- Modify: `web/admin/src/pages/articles/index.tsx`
- Modify: `web/admin/src/pages/articles/detail.tsx`
- Modify: `web/admin/src/pages/articles/rectify.tsx`
- Modify: `web/admin/src/pages/tasks/index.tsx`
- Modify: `web/admin/src/pages/tasks/detail.tsx`
- Modify: `web/admin/src/pages/tasks/results.tsx`
- Modify: `web/admin/src/pages/logs/index.tsx`
- Modify: `web/admin/src/pages/results/index.tsx`
- Modify: `web/admin/src/pages/categories/index.tsx`

**Step 1:** Add a shared navigation helper and optional `WorkbenchLink` wrapper that routes every internal jump through the workbench store plus `useNavigate`.

**Step 2:** Update `SidebarNav` so clicking a top-level item activates or opens a tab instead of relying on plain `NavLink` behavior.

**Step 3:** Replace in-page `href` and raw `<a>` usage with workbench navigation in the article, task, log, category, and result pages.

**Step 4:** Preserve deep-link access: if a page is loaded directly by URL, the provider should still create or activate the matching tab during mount.

**Step 5:** Re-run the focused navigation suite from Task 5 and confirm green.

**Step 6:** Commit the migrated navigation.

```bash
git add web/admin/src/workbench/navigation.ts web/admin/src/workbench/link.tsx web/admin/src/components/layout/sidebar-nav.tsx web/admin/src/pages/articles/index.tsx web/admin/src/pages/articles/detail.tsx web/admin/src/pages/articles/rectify.tsx web/admin/src/pages/tasks/index.tsx web/admin/src/pages/tasks/detail.tsx web/admin/src/pages/tasks/results.tsx web/admin/src/pages/logs/index.tsx web/admin/src/pages/results/index.tsx web/admin/src/pages/categories/index.tsx
git commit -m "feat: route admin navigation through workbench tabs"
```

### Task 7: Lock query-string state restoration for list pages

**Files:**
- Modify: `web/admin/src/pages/articles/index.test.tsx`
- Modify: `web/admin/src/pages/tasks/index.test.tsx`
- Modify: `web/admin/src/pages/logs/index.test.tsx`
- Modify: `web/admin/src/pages/categories/index.test.tsx`
- Modify: `web/admin/src/pages/keywords/index.test.tsx`

**Step 1:** Add failing tests that mount each list page with a query string and expect the UI to hydrate from URL state.

```ts
renderWithRoute('/articles?title=命中&page=2&article_id=901');
expect(screen.getByRole('textbox', { name: '标题模糊查询' })).toHaveValue('命中');
expect(mockedListArticles).toHaveBeenCalledWith(expect.objectContaining({
  title: '命中',
  article_id: 901,
  page: 2,
}));
```

**Step 2:** Add expectations that submitting filters updates the URL and that reopening the existing list tab restores the same filters.

**Step 3:** Run `npm --prefix web/admin test -- src/pages/articles/index.test.tsx src/pages/tasks/index.test.tsx src/pages/logs/index.test.tsx src/pages/categories/index.test.tsx src/pages/keywords/index.test.tsx` and confirm failure.

**Step 4:** Commit the list-state tests.

```bash
git add web/admin/src/pages/articles/index.test.tsx web/admin/src/pages/tasks/index.test.tsx web/admin/src/pages/logs/index.test.tsx web/admin/src/pages/categories/index.test.tsx web/admin/src/pages/keywords/index.test.tsx
git commit -m "test: lock admin list query restoration"
```

### Task 8: Implement URL-backed list state and tab-session page state

**Files:**
- Create: `web/admin/src/workbench/page-session.ts`
- Modify: `web/admin/src/pages/articles/index.tsx`
- Modify: `web/admin/src/pages/tasks/index.tsx`
- Modify: `web/admin/src/pages/logs/index.tsx`
- Modify: `web/admin/src/pages/categories/index.tsx`
- Modify: `web/admin/src/pages/keywords/index.tsx`
- Modify: `web/admin/src/pages/articles/detail.tsx`
- Modify: `web/admin/src/pages/articles/rectify.tsx`
- Modify: `web/admin/src/pages/tasks/detail.tsx`
- Modify: `web/admin/src/pages/tasks/results.tsx`

**Step 1:** Add a small page-session helper keyed by tab key for non-URL state such as scroll position, nested tabs, and unsaved form drafts.

**Step 2:** Refactor each list page to read initial filter/pagination state from `useSearchParams`, write updates back on submit and pagination changes, and keep network requests driven by the URL-backed submitted state.

**Step 3:** Refactor detail, result, and rectify pages to store their local active-tab / draft / scroll state through the page-session helper rather than losing it when the active tab changes.

**Step 4:** Ensure closing a tab clears its in-memory page session payload.

**Step 5:** Re-run the focused list-page suite from Task 7 plus detail-page tests from Task 5 and confirm green.

**Step 6:** Commit the state restoration work.

```bash
git add web/admin/src/workbench/page-session.ts web/admin/src/pages/articles/index.tsx web/admin/src/pages/tasks/index.tsx web/admin/src/pages/logs/index.tsx web/admin/src/pages/categories/index.tsx web/admin/src/pages/keywords/index.tsx web/admin/src/pages/articles/detail.tsx web/admin/src/pages/articles/rectify.tsx web/admin/src/pages/tasks/detail.tsx web/admin/src/pages/tasks/results.tsx
git commit -m "feat: preserve admin workbench page state"
```

### Task 9: Verify direct-route recovery, build stability, and full frontend coverage

**Files:**
- Verify only: `web/admin/src/App.tsx`
- Verify only: `web/admin/src/workbench/provider.tsx`
- Verify only: `web/admin/src/workbench/navigation.ts`

**Step 1:** Run the full admin frontend test suite.

Run: `npm --prefix web/admin test`
Expected: PASS with the new workbench tests and updated page suites green.

**Step 2:** Run the production build.

Run: `npm --prefix web/admin build`
Expected: PASS with no TypeScript errors and a successful Vite build.

**Step 3:** Smoke-check these direct-entry URLs in tests or a manual local run:
- `/tasks`
- `/articles`
- `/articles/501`
- `/articles/501/rectify?task_id=77&result_id=11`
- `/tasks/77`
- `/tasks/77/results`

**Step 4:** Summarize any remaining gaps, especially around session size limits or future keep-alive eviction, before merging.

**Step 5:** Commit the final verified state.

```bash
git add web/admin
git commit -m "feat: add admin tabbed workbench"
```

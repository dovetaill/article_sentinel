# Admin Ant Design Pro Replatform Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace `web/admin` with an `ant-design-pro`-based admin app that preserves all existing business capabilities, adds an `iview-admin`-style tabbed workspace, and keeps old admin history only as a Git tag.

**Architecture:** Rebuild the frontend on `Umi Max + ant-design-pro + Ant Design/ProComponents`, keep the existing backend API contract, and reorganize the information architecture into inspection, rules, content, and audit domains. Use `ProLayout` as the shell, add a custom `PageTabs` layer for multi-tab navigation, and preserve session/org semantics through runtime initialization rather than local login UI.

**Tech Stack:** React 19, Umi Max 4, ant-design-pro, Ant Design, ProComponents, TypeScript, Less, Vitest, React Testing Library

---

## Preflight

Before Task 1, do these git-only steps in order:

1. Verify the design doc commit exists:
   - `git log --oneline -n 3`
   - Confirm commit `58f530f` (`docs(plans): add admin pro replatform design`) is present.
2. Create the legacy admin tag from the current tree before deleting `web/admin`:
   - `git tag -a archive/admin-react-legacy-2026-05-09 -m "Archive legacy React admin before ant-design-pro rewrite"`
   - Expected: no output, then `git tag --list | rg 'archive/admin-react-legacy-2026-05-09'` prints the tag name.
3. Create a dedicated worktree for implementation:
   - `git worktree add ../article_sentinel-admin-pro -b feat/admin-ant-design-pro-replatform HEAD`
   - `cd ../article_sentinel-admin-pro`
4. Execute the remaining tasks only inside that worktree.

### Task 1: Bootstrap the ant-design-pro / Umi Max app shell

**Files:**
- Delete: `web/admin/index.html`
- Delete: `web/admin/vite.config.ts`
- Delete: `web/admin/src/main.tsx`
- Delete: `web/admin/src/App.tsx`
- Delete: `web/admin/src/routes.tsx`
- Delete: `web/admin/src/styles.css`
- Delete: `web/admin/src/components/`
- Delete: `web/admin/src/context/`
- Delete: `web/admin/src/lib/`
- Delete: `web/admin/src/pages/`
- Delete: `web/admin/src/services/`
- Delete: `web/admin/src/styles/`
- Delete: `web/admin/src/workbench/`
- Modify: `web/admin/package.json`
- Modify: `web/admin/tsconfig.json`
- Create: `web/admin/config/config.ts`
- Create: `web/admin/config/routes.ts`
- Create: `web/admin/config/defaultSettings.ts`
- Create: `web/admin/config/proxy.ts`
- Create: `web/admin/src/app.tsx`
- Create: `web/admin/src/access.ts`
- Create: `web/admin/src/global.less`
- Create: `web/admin/src/layouts/BasicLayout.tsx`
- Create: `web/admin/src/pages/User/LoginRedirect/index.tsx`
- Create: `web/admin/vitest.config.ts`
- Create: `web/admin/src/test/setup.ts`

**Step 1: Replace the package/toolchain skeleton**

Use `@superpowers:test-driven-development` where practical, but treat this task as framework bootstrap rather than feature TDD.

Replace `web/admin/package.json` with scripts and dependencies shaped like:

```json
{
  "scripts": {
    "dev": "umi dev --port 5173 --host 0.0.0.0",
    "build": "umi build",
    "test": "vitest run"
  },
  "dependencies": {
    "@ant-design/pro-components": "...",
    "@ant-design/pro-layout": "...",
    "@umijs/max": "...",
    "antd": "...",
    "react": "...",
    "react-dom": "..."
  }
}
```

**Step 2: Add the minimal Umi config and placeholder routes**

Create `web/admin/config/routes.ts` with a minimal placeholder tree:

```ts
export default [
  { path: '/user/login', layout: false, component: './User/LoginRedirect' },
  { path: '/', redirect: '/inspection/tasks' },
  {
    path: '/',
    component: '@/layouts/BasicLayout',
    routes: [
      { path: '/inspection/tasks', name: '检测任务', component: './Inspection/TaskList' },
    ],
  },
  { path: '/*', component: '404', layout: false },
];
```

**Step 3: Install dependencies and verify the skeleton builds**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm install`
Expected: install completes without peer dependency errors.

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm run build`
Expected: PASS and emits a new `web/admin/dist` build from Umi.

**Step 4: Commit the bootstrap**

```bash
git add web/admin/package.json web/admin/package-lock.json web/admin/tsconfig.json web/admin/config web/admin/src web/admin/vitest.config.ts
git commit -m "build(admin): bootstrap umi pro shell"
```

### Task 2: Rebuild the request/runtime auth layer

**Files:**
- Create: `web/admin/src/utils/auth.ts`
- Create: `web/admin/src/utils/auth.test.ts`
- Create: `web/admin/src/services/auth.ts`
- Create: `web/admin/src/services/request.ts`
- Modify: `web/admin/src/app.tsx`
- Modify: `web/admin/src/access.ts`
- Modify: `web/admin/src/pages/User/LoginRedirect/index.tsx`
- Test: `web/admin/src/utils/auth.test.ts`

**Step 1: Write the failing auth helper tests**

Create `web/admin/src/utils/auth.test.ts` with cases like:

```ts
import { describe, expect, it } from 'vitest';
import { LOGIN_ENTRY_PATH, normalizeAdminPath } from './auth';

describe('normalizeAdminPath', () => {
  it('maps the root path to /inspection/tasks', () => {
    expect(normalizeAdminPath('/')).toBe('/inspection/tasks');
  });

  it('keeps known business paths unchanged', () => {
    expect(normalizeAdminPath('/content/articles')).toBe('/content/articles');
  });
});

describe('LOGIN_ENTRY_PATH', () => {
  it('points to the fixed backend login entry', () => {
    expect(LOGIN_ENTRY_PATH).toBe('/auth/login');
  });
});
```

**Step 2: Run the tests to verify they fail**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/utils/auth.test.ts`
Expected: FAIL because `src/utils/auth.ts` does not exist yet.

**Step 3: Implement the auth/request helpers**

Create `web/admin/src/utils/auth.ts` and `web/admin/src/services/request.ts` around this shape:

```ts
export const LOGIN_ENTRY_PATH = '/auth/login';

export function redirectToFixedLogin() {
  window.location.assign(LOGIN_ENTRY_PATH);
}

export function normalizeAdminPath(pathname: string) {
  return pathname === '/' ? '/inspection/tasks' : pathname;
}
```

In `web/admin/src/app.tsx`, export Umi runtime hooks that:

- call `/api/v1/auth/session` during `getInitialState`
- store `currentUser`, `currentOrgId`, `currentOrgName`
- redirect to `/auth/login` on 401
- set `credentials: 'same-origin'`

**Step 4: Re-run the focused tests**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/utils/auth.test.ts`
Expected: PASS.

**Step 5: Smoke test the runtime build**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm run build`
Expected: PASS.

**Step 6: Commit**

```bash
git add web/admin/src/utils/auth.ts web/admin/src/utils/auth.test.ts web/admin/src/services/auth.ts web/admin/src/services/request.ts web/admin/src/app.tsx web/admin/src/access.ts web/admin/src/pages/User/LoginRedirect/index.tsx
git commit -m "feat(admin): restore session runtime and login redirect"
```

### Task 3: Add Umi proxy, layout settings, and browser-compat configuration

**Files:**
- Modify: `web/admin/config/config.ts`
- Modify: `web/admin/config/defaultSettings.ts`
- Modify: `web/admin/config/proxy.ts`
- Modify: `web/admin/src/global.less`
- Test: `web/admin/package.json`

**Step 1: Configure the runtime targets and antd style compatibility**

Update `web/admin/config/config.ts` with Umi/antd settings like:

```ts
export default defineConfig({
  routes,
  antd: {
    styleProvider: {
      hashPriority: 'high',
      legacyTransformer: true,
    },
  },
  targets: {
    chrome: 88,
    edge: 88,
  },
  proxy: proxy[process.env.NODE_ENV || 'dev'],
});
```

**Step 2: Wire the API proxy to match the old Vite behavior**

Set `web/admin/config/proxy.ts` so `/api` and `/auth` proxy to `process.env.ADMIN_API_BASE_URL || 'http://127.0.0.1:8080'`.

**Step 3: Apply the dark sider defaults**

In `web/admin/config/defaultSettings.ts`, set the shell baseline:

```ts
const settings = {
  navTheme: 'realDark',
  layout: 'mix',
  colorPrimary: '#2d8cf0',
  fixSiderbar: true,
  fixedHeader: true,
  siderWidth: 232,
};
```

And in `web/admin/src/global.less`, add a root token block that includes `@admin-sider-bg: #191a23;` and the light content background palette.

**Step 4: Verify dev/build startup**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm run build`
Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin/config/config.ts web/admin/config/defaultSettings.ts web/admin/config/proxy.ts web/admin/src/global.less
git commit -m "feat(admin): configure proxy theme and browser targets"
```

### Task 4: Implement the tabbed workspace shell on top of ProLayout

**Files:**
- Create: `web/admin/src/components/PageTabs/route-meta.ts`
- Create: `web/admin/src/components/PageTabs/store.ts`
- Create: `web/admin/src/components/PageTabs/store.test.ts`
- Create: `web/admin/src/components/PageTabs/index.tsx`
- Modify: `web/admin/src/layouts/BasicLayout.tsx`
- Modify: `web/admin/config/routes.ts`
- Modify: `web/admin/src/global.less`
- Test: `web/admin/src/components/PageTabs/store.test.ts`

**Step 1: Write the failing tab store tests**

Create `web/admin/src/components/PageTabs/store.test.ts` with reducer cases like:

```ts
import { describe, expect, it } from 'vitest';
import { reduceTabs, restoreDefaultTabs } from './store';

describe('reduceTabs', () => {
  it('opens /inspection/tasks as the fixed base tab', () => {
    const state = restoreDefaultTabs(29);
    expect(state.tabs[0].pathname).toBe('/inspection/tasks');
    expect(state.tabs[0].closable).toBe(false);
  });

  it('deduplicates detail routes by key', () => {
    const next = reduceTabs(restoreDefaultTabs(29), {
      type: 'open',
      href: '/content/articles/42',
      orgId: 29,
    });
    expect(next.tabs).toHaveLength(2);
  });
});
```

**Step 2: Run the tests to verify they fail**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/components/PageTabs/store.test.ts`
Expected: FAIL because the store files do not exist yet.

**Step 3: Implement route metadata and tab persistence**

Create the reducer around a structure like:

```ts
export type TabState = {
  orgId: number;
  activeKey: string;
  tabs: Array<{ key: string; pathname: string; search: string; title: string; closable: boolean }>;
};

export function restoreDefaultTabs(orgId: number): TabState {
  return {
    orgId,
    activeKey: '/inspection/tasks',
    tabs: [
      { key: '/inspection/tasks', pathname: '/inspection/tasks', search: '', title: '检测任务', closable: false },
    ],
  };
}
```

Then render it in `web/admin/src/layouts/BasicLayout.tsx` with `ProLayout` + `PageTabs` + `Outlet`.

**Step 4: Re-run the focused tests**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/components/PageTabs/store.test.ts`
Expected: PASS.

**Step 5: Manually verify tab behavior in dev**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm run dev`
Expected: app starts on `http://0.0.0.0:5173` and shows the dark sider plus a visible tab bar.

**Step 6: Commit**

```bash
git add web/admin/src/components/PageTabs web/admin/src/layouts/BasicLayout.tsx web/admin/config/routes.ts web/admin/src/global.less
git commit -m "feat(admin): add pro layout and tabbed workspace"
```

### Task 5: Recreate shared UI primitives and common service modules

**Files:**
- Create: `web/admin/src/components/StatusTag/index.tsx`
- Create: `web/admin/src/components/StatusTag/index.test.tsx`
- Create: `web/admin/src/components/SnapshotViewer/index.tsx`
- Create: `web/admin/src/components/SnapshotViewer/index.test.tsx`
- Create: `web/admin/src/components/HitPreview/index.tsx`
- Create: `web/admin/src/components/HtmlArticleEditor/index.tsx`
- Create: `web/admin/src/services/categories.ts`
- Create: `web/admin/src/services/keywords.ts`
- Create: `web/admin/src/services/tasks.ts`
- Create: `web/admin/src/services/results.ts`
- Create: `web/admin/src/services/articles.ts`
- Create: `web/admin/src/services/logs.ts`
- Test: `web/admin/src/components/StatusTag/index.test.tsx`
- Test: `web/admin/src/components/SnapshotViewer/index.test.tsx`

**Step 1: Write failing tests for the shared UI pieces**

Create `StatusTag` and `SnapshotViewer` tests, for example:

```tsx
it('renders task status labels in Chinese', () => {
  render(<StatusTag kind="task" value="running" />);
  expect(screen.getByText('执行中')).toBeInTheDocument();
});

it('shows a fallback when the snapshot is empty', () => {
  render(<SnapshotViewer value={null} emptyText="暂无请求快照。" />);
  expect(screen.getByText('暂无请求快照。')).toBeInTheDocument();
});
```

**Step 2: Run the tests to verify they fail**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/components/StatusTag/index.test.tsx src/components/SnapshotViewer/index.test.tsx`
Expected: FAIL.

**Step 3: Port the shared presentation logic and service contracts**

Reuse the current business semantics from the legacy tag, but reshape the UI components around the new names and folders. Keep each service contract stable:

```ts
export async function listTasks(params: ListTaskParams) {
  return request<ApiPage<TaskRecord>>('/api/v1/tasks', { params });
}
```

**Step 4: Re-run the focused tests**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/components/StatusTag/index.test.tsx src/components/SnapshotViewer/index.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin/src/components web/admin/src/services
git commit -m "feat(admin): restore shared components and service modules"
```

### Task 6: Migrate the rules center pages

**Files:**
- Create: `web/admin/src/pages/Rules/CategoryList/index.tsx`
- Create: `web/admin/src/pages/Rules/CategoryList/index.test.tsx`
- Create: `web/admin/src/pages/Rules/KeywordList/index.tsx`
- Create: `web/admin/src/pages/Rules/KeywordList/index.test.tsx`
- Modify: `web/admin/config/routes.ts`
- Modify: `web/admin/src/global.less`
- Test: `web/admin/src/pages/Rules/CategoryList/index.test.tsx`
- Test: `web/admin/src/pages/Rules/KeywordList/index.test.tsx`

**Step 1: Write the failing page tests**

Create assertions for the two list pages, for example:

```tsx
it('renders the category page heading and query controls', async () => {
  render(<CategoryListPage />);
  expect(screen.getByRole('heading', { name: '规则分类' })).toBeInTheDocument();
  expect(screen.getByLabelText('分类名称')).toBeInTheDocument();
});

it('renders the keyword page heading and add button', async () => {
  render(<KeywordListPage />);
  expect(screen.getByRole('heading', { name: '规则管理' })).toBeInTheDocument();
  expect(screen.getByRole('button', { name: '新增规则' })).toBeInTheDocument();
});
```

**Step 2: Run the tests to verify they fail**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Rules/CategoryList/index.test.tsx src/pages/Rules/KeywordList/index.test.tsx`
Expected: FAIL.

**Step 3: Implement the category and keyword pages**

Port the existing list/search/form behavior into `PageContainer + ProTable + ModalForm`, keeping search params in the URL and preserving current CRUD semantics.

**Step 4: Re-run the focused tests**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Rules/CategoryList/index.test.tsx src/pages/Rules/KeywordList/index.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin/src/pages/Rules web/admin/config/routes.ts web/admin/src/global.less
git commit -m "feat(admin): migrate rules center pages"
```

### Task 7: Migrate the task list and task creation pages

**Files:**
- Create: `web/admin/src/pages/Inspection/TaskList/index.tsx`
- Create: `web/admin/src/pages/Inspection/TaskList/index.test.tsx`
- Create: `web/admin/src/pages/Inspection/TaskCreate/index.tsx`
- Create: `web/admin/src/pages/Inspection/TaskCreate/index.test.tsx`
- Modify: `web/admin/config/routes.ts`
- Modify: `web/admin/src/components/PageTabs/route-meta.ts`
- Test: `web/admin/src/pages/Inspection/TaskList/index.test.tsx`
- Test: `web/admin/src/pages/Inspection/TaskCreate/index.test.tsx`

**Step 1: Write the failing tests**

Create tests that lock in the key behaviors:

```tsx
it('shows the task list filters and create button', async () => {
  render(<TaskListPage />);
  expect(screen.getByLabelText('任务编号')).toBeInTheDocument();
  expect(screen.getByRole('button', { name: '新建任务' })).toBeInTheDocument();
});

it('requires at least one rule before submitting a task', async () => {
  render(<TaskCreatePage />);
  await user.click(screen.getByRole('button', { name: '提交任务' }));
  expect(await screen.findByText('请先选择至少一条规则')).toBeInTheDocument();
});
```

**Step 2: Run the tests to verify they fail**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Inspection/TaskList/index.test.tsx src/pages/Inspection/TaskCreate/index.test.tsx`
Expected: FAIL.

**Step 3: Port the task list/create flows**

Rebuild them with `PageContainer`, `ProTable`, and `ProForm`, while keeping:

- URL-based filters and pagination
- keyword selection from the org scope
- task delete restrictions
- redirect to `/inspection/tasks` after successful create

**Step 4: Re-run the focused tests**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Inspection/TaskList/index.test.tsx src/pages/Inspection/TaskCreate/index.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin/src/pages/Inspection/TaskList web/admin/src/pages/Inspection/TaskCreate web/admin/config/routes.ts web/admin/src/components/PageTabs/route-meta.ts
git commit -m "feat(admin): migrate task list and creation flow"
```

### Task 8: Migrate task detail and the global results list

**Files:**
- Create: `web/admin/src/pages/Inspection/TaskDetail/index.tsx`
- Create: `web/admin/src/pages/Inspection/TaskDetail/index.test.tsx`
- Create: `web/admin/src/pages/Inspection/ResultList/index.tsx`
- Create: `web/admin/src/pages/Inspection/ResultList/index.test.tsx`
- Modify: `web/admin/config/routes.ts`
- Modify: `web/admin/src/components/PageTabs/route-meta.ts`
- Test: `web/admin/src/pages/Inspection/TaskDetail/index.test.tsx`
- Test: `web/admin/src/pages/Inspection/ResultList/index.test.tsx`

**Step 1: Write the failing tests**

Use cases to lock in:

```tsx
it('renders task detail tabs for hit results and snapshots', async () => {
  render(<TaskDetailPage />);
  expect(screen.getByRole('tab', { name: '命中结果' })).toBeInTheDocument();
  expect(screen.getByRole('tab', { name: '规则快照' })).toBeInTheDocument();
});

it('renders the global result list batch action', async () => {
  render(<ResultListPage />);
  expect(screen.getByRole('button', { name: '批量下线处置' })).toBeInTheDocument();
});
```

**Step 2: Run the tests to verify they fail**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Inspection/TaskDetail/index.test.tsx src/pages/Inspection/ResultList/index.test.tsx`
Expected: FAIL.

**Step 3: Implement the detail/results views**

Port the current behavior into the new route structure, preserving:

- result previews
- snapshot rendering
- log summaries
- batch offline action
- links from results to content detail / rectify pages

**Step 4: Re-run the focused tests**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Inspection/TaskDetail/index.test.tsx src/pages/Inspection/ResultList/index.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin/src/pages/Inspection/TaskDetail web/admin/src/pages/Inspection/ResultList web/admin/config/routes.ts web/admin/src/components/PageTabs/route-meta.ts
git commit -m "feat(admin): migrate task detail and result list"
```

### Task 9: Migrate the article list and article detail pages

**Files:**
- Create: `web/admin/src/pages/Content/ArticleList/index.tsx`
- Create: `web/admin/src/pages/Content/ArticleList/index.test.tsx`
- Create: `web/admin/src/pages/Content/ArticleDetail/index.tsx`
- Create: `web/admin/src/pages/Content/ArticleDetail/index.test.tsx`
- Modify: `web/admin/config/routes.ts`
- Modify: `web/admin/src/components/PageTabs/route-meta.ts`
- Test: `web/admin/src/pages/Content/ArticleList/index.test.tsx`
- Test: `web/admin/src/pages/Content/ArticleDetail/index.test.tsx`

**Step 1: Write the failing tests**

```tsx
it('renders article list filters for title and article id', () => {
  render(<ArticleListPage />);
  expect(screen.getByLabelText('标题模糊查询')).toBeInTheDocument();
  expect(screen.getByLabelText('按文稿ID查询')).toBeInTheDocument();
});

it('renders article detail actions and tabs', async () => {
  render(<ArticleDetailPage />);
  expect(screen.getByRole('button', { name: '进入整改' })).toBeInTheDocument();
  expect(screen.getByRole('tab', { name: '命中记录' })).toBeInTheDocument();
});
```

**Step 2: Run the tests to verify they fail**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Content/ArticleList/index.test.tsx src/pages/Content/ArticleDetail/index.test.tsx`
Expected: FAIL.

**Step 3: Implement the article list/detail views**

Keep:

- article title/id URL filters
- return-to semantics from result/task pages
- operation logs and field change tabs
- latest task/result summary cards

**Step 4: Re-run the focused tests**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Content/ArticleList/index.test.tsx src/pages/Content/ArticleDetail/index.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin/src/pages/Content/ArticleList web/admin/src/pages/Content/ArticleDetail web/admin/config/routes.ts web/admin/src/components/PageTabs/route-meta.ts
git commit -m "feat(admin): migrate article list and detail"
```

### Task 10: Migrate the content rectify page and draft persistence

**Files:**
- Create: `web/admin/src/pages/Content/ArticleRectify/index.tsx`
- Create: `web/admin/src/pages/Content/ArticleRectify/index.test.tsx`
- Modify: `web/admin/src/components/HtmlArticleEditor/index.tsx`
- Modify: `web/admin/src/components/PageTabs/route-meta.ts`
- Modify: `web/admin/src/global.less`
- Test: `web/admin/src/pages/Content/ArticleRectify/index.test.tsx`

**Step 1: Write the failing rectify tests**

```tsx
it('renders the rectify form and original article panel', async () => {
  render(<ArticleRectifyPage />);
  expect(screen.getByLabelText('整改标题')).toBeInTheDocument();
  expect(screen.getByText('原稿对照')).toBeInTheDocument();
});

it('keeps a local draft while editing', async () => {
  render(<ArticleRectifyPage />);
  await user.type(screen.getByLabelText('整改标题'), '新标题');
  expect(saveDraft).toHaveBeenCalled();
});
```

**Step 2: Run the tests to verify they fail**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Content/ArticleRectify/index.test.tsx`
Expected: FAIL.

**Step 3: Implement the rectify flow**

Carry forward:

- title/desc/body editing
- original article comparison panel
- save rectification
- save and submit for review
- draft persistence keyed by org + route key

**Step 4: Re-run the focused tests**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Content/ArticleRectify/index.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin/src/pages/Content/ArticleRectify web/admin/src/components/HtmlArticleEditor/index.tsx web/admin/src/components/PageTabs/route-meta.ts web/admin/src/global.less
git commit -m "feat(admin): migrate article rectify workspace"
```

### Task 11: Migrate the audit log page and legacy route redirects

**Files:**
- Create: `web/admin/src/pages/Audit/OperationLogList/index.tsx`
- Create: `web/admin/src/pages/Audit/OperationLogList/index.test.tsx`
- Create: `web/admin/src/pages/LegacyRedirect/index.tsx`
- Create: `web/admin/src/pages/LegacyRedirect/index.test.tsx`
- Modify: `web/admin/config/routes.ts`
- Test: `web/admin/src/pages/Audit/OperationLogList/index.test.tsx`
- Test: `web/admin/src/pages/LegacyRedirect/index.test.tsx`

**Step 1: Write the failing tests**

```tsx
it('renders log filters and the snapshot action', async () => {
  render(<OperationLogListPage />);
  expect(screen.getByLabelText('文章编号')).toBeInTheDocument();
  expect(screen.getByRole('button', { name: '查询日志' })).toBeInTheDocument();
});

it('maps /tasks to /inspection/tasks', () => {
  expect(resolveLegacyPath('/tasks')).toBe('/inspection/tasks');
});
```

**Step 2: Run the tests to verify they fail**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Audit/OperationLogList/index.test.tsx src/pages/LegacyRedirect/index.test.tsx`
Expected: FAIL.

**Step 3: Implement the audit page and redirect page**

Use a redirect helper that maps:

```ts
const legacyMap = {
  '/tasks': '/inspection/tasks',
  '/tasks/new': '/inspection/tasks/create',
  '/articles': '/content/articles',
  '/logs': '/audit/logs',
};
```

For dynamic routes, render a tiny page component that reads params and `navigate()`s to the new path.

**Step 4: Re-run the focused tests**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Audit/OperationLogList/index.test.tsx src/pages/LegacyRedirect/index.test.tsx`
Expected: PASS.

**Step 5: Commit**

```bash
git add web/admin/src/pages/Audit/OperationLogList web/admin/src/pages/LegacyRedirect web/admin/config/routes.ts
git commit -m "feat(admin): migrate audit logs and legacy redirects"
```

### Task 12: Update repo tooling/docs and run the full verification sweep

**Files:**
- Modify: `README.md`
- Modify: `scripts/dev_test.sh`
- Modify: `web/admin/package.json`
- Modify: `web/admin/config/config.ts`
- Modify: `web/admin/config/proxy.ts`
- Modify: `web/admin/src/global.less`
- Test: `docs/plans/2026-05-09-admin-ant-design-pro-replatform-design.md`
- Test: `docs/plans/2026-05-09-admin-ant-design-pro-replatform.md`

**Step 1: Update docs and dev expectations**

Adjust `README.md` and `scripts/dev_test.sh` so they describe and assert:

- `web/admin` is now `React + Umi Max + ant-design-pro`
- dev server remains `0.0.0.0:5173`
- proxy behavior now comes from `web/admin/config/proxy.ts`
- the old Vite-specific wording is removed

**Step 2: Run the focused tooling checks**

Run: `bash scripts/dev_test.sh`
Expected: PASS, including the updated admin proxy checks.

**Step 3: Run the frontend test suite**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test`
Expected: PASS for all page/component/runtime tests.

**Step 4: Run the production build**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm run build`
Expected: PASS and emits a deployable `dist` bundle.

**Step 5: Run final repo-level verification**

Use `@superpowers:verification-before-completion` and, if anything fails unexpectedly, switch to `@superpowers:systematic-debugging`.

Run: `git diff --stat`
Expected: only the intended admin rewrite, docs, and tooling files are changed.

**Step 6: Commit**

```bash
git add README.md scripts/dev_test.sh web/admin
git commit -m "feat(admin): replace legacy admin with ant design pro"
```

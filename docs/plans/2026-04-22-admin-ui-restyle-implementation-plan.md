# Admin UI Restyle Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rebuild `web/admin` so it preserves the current article-inspection information architecture and backend integrations while visually matching the `shadcn-admin` dashboard style in a formal, natural Chinese presentation for a government media organization.

**Architecture:** Keep the existing React Router routes and `src/services/*` API layer intact, but replace the current ProLayout-driven shell with a custom dashboard shell, shared presentational primitives, and page-specific layouts. Rework each business page around one unified set of design tokens, formal Chinese copy, and lightweight wrappers so the UI changes stay in the presentation layer without changing business behavior.

**Tech Stack:** React 18, TypeScript, Vite, Ant Design base components, React Router 7, custom CSS tokens/layouts, Vitest, Testing Library.

---

### Task 1: Establish the Chinese theme foundation and design tokens

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/main.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles.css`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/styles/theme.css`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`
- Test: `/home/wwwroot/article_sentinel/web/admin/src/App.test.tsx`

**Step 1: Write the failing shell copy test**

Update `App.test.tsx` to expect:

- Chinese platform name `融媒内容安全巡检平台`
- Chinese navigation labels `关键词规则 / 检测任务 / 风险结果 / 操作日志`
- no English `Article Sentinel` heading in the shell

**Step 2: Run test to verify it fails**

Run:

```bash
npm test -- src/App.test.tsx
```

Expected: FAIL because the shell still renders the English title and route labels.

**Step 3: Add formal Chinese theme tokens**

Implement:

- `zh_CN` locale in `main.tsx`
- formal color tokens centered on slate blue / ink gray / muted gold
- shared CSS variables for typography, border, shadow, spacing, and status colors
- stylesheet imports for `theme.css` and `layout.css`

**Step 4: Re-run the shell test**

Run:

```bash
npm test -- src/App.test.tsx
```

Expected: PASS once the root shell copy and tokens are switched to the Chinese design system.

**Step 5: Commit**

```bash
git add web/admin/src/main.tsx web/admin/src/styles.css web/admin/src/styles/theme.css web/admin/src/styles/layout.css web/admin/src/App.test.tsx
git commit -m "feat: establish formal admin theme tokens"
```

### Task 2: Replace the app shell with a shadcn-admin-like dashboard frame

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/App.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/routes.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/components/layout/admin-shell.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/components/layout/sidebar-nav.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/components/layout/topbar.tsx`
- Test: `/home/wwwroot/article_sentinel/web/admin/src/App.test.tsx`

**Step 1: Write the failing dashboard-shell assertions**

Extend `App.test.tsx` to check:

- left navigation renders the four approved Chinese modules
- top bar renders the current section title and the system title
- the old hero banner copy is gone

**Step 2: Run test to verify it fails**

Run:

```bash
npm test -- src/App.test.tsx
```

Expected: FAIL because `ProLayout` still renders the previous shell.

**Step 3: Implement the new shell**

Build:

- `AdminShell` with persistent sidebar, top bar, and content container
- route metadata in Chinese, including formal titles and descriptions
- a shell structure visually aligned with the `shadcn-admin` layout proportions
- active navigation styling, breadcrumb/title presentation, and top-right status area

**Step 4: Re-run the app shell test**

Run:

```bash
npm test -- src/App.test.tsx
```

Expected: PASS once the custom shell replaces `ProLayout`.

**Step 5: Commit**

```bash
git add web/admin/src/App.tsx web/admin/src/routes.tsx web/admin/src/components/layout web/admin/src/App.test.tsx
git commit -m "feat: rebuild admin dashboard shell"
```

### Task 3: Add shared presentation primitives for headers, summary cards, and status badges

**Files:**
- Create: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/page-header.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/summary-card.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/status-badge.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/section-card.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/toolbar-strip.tsx`
- Create: `/home/wwwroot/article_sentinel/web/admin/src/components/ui/components.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing shared-component test**

Add `components.test.tsx` coverage for:

- `PageHeader` rendering a Chinese title and description
- `SummaryCard` rendering label, value, and helper text
- `StatusBadge` mapping low/medium/high and task states to formal badge copy/classes

**Step 2: Run test to verify it fails**

Run:

```bash
npm test -- src/components/ui/components.test.tsx
```

Expected: FAIL because the shared UI primitives do not exist yet.

**Step 3: Implement the shared primitives**

Create reusable presentational blocks for:

- page titles and action areas
- summary metrics
- consistent content sections
- toolbars for table actions and filters
- unified formal badge styling across risk and workflow states

**Step 4: Re-run the component test**

Run:

```bash
npm test -- src/components/ui/components.test.tsx
```

Expected: PASS after the shared primitives render the new shell language correctly.

**Step 5: Commit**

```bash
git add web/admin/src/components/ui web/admin/src/styles/layout.css
git commit -m "feat: add shared admin presentation components"
```

### Task 4: Restyle the keyword rules page without changing keyword behavior

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/keywords/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/keywords/index.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing keyword-page expectations**

Update `index.test.tsx` to expect:

- page title `关键词规则`
- action button `新增规则`
- modal titles `新增规则` and `编辑规则`
- formal Chinese table headers and status copy

**Step 2: Run test to verify it fails**

Run:

```bash
npm test -- src/pages/keywords/index.test.tsx
```

Expected: FAIL because the page still renders English copy and the old ProTable framing.

**Step 3: Rebuild the keyword page presentation**

Implement:

- `PageHeader` plus `SectionCard` wrapper
- formal Chinese toolbar, table headers, modal copy, and field labels
- unified risk/status badges and cleaner table cell spacing
- dialog layout that feels like a rule-maintenance panel rather than default form chrome

**Step 4: Re-run the keyword-page test**

Run:

```bash
npm test -- src/pages/keywords/index.test.tsx
```

Expected: PASS while keyword CRUD behavior remains unchanged.

**Step 5: Commit**

```bash
git add web/admin/src/pages/keywords/index.tsx web/admin/src/pages/keywords/index.test.tsx web/admin/src/styles/layout.css
git commit -m "feat: restyle keyword rules page"
```

### Task 5: Restyle the task list and task creation pages with formal dashboard patterns

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/new.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/new.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing task-page tests**

Update the task tests to expect:

- page title `检测任务`
- button copy `新建任务`
- summary cards for task overview
- formal Chinese labels on the new-task form, including `所属机构`, `关键词范围`, `发布时间起`, `发布时间止`, `是否检索正文`, `文章编号`, `标题检索`

**Step 2: Run tests to verify they fail**

Run:

```bash
npm test -- src/pages/tasks/index.test.tsx src/pages/tasks/new.test.tsx
```

Expected: FAIL because both task pages still use English framing and ad-hoc layout.

**Step 3: Implement the task-page redesign**

Build:

- overview summary cards above the task list
- formal Chinese table headers and detail drawer copy
- a two-column new-task page with the form on the left and process guidance on the right
- polished switch/select/input presentation without changing the submission payload

**Step 4: Re-run the task-page tests**

Run:

```bash
npm test -- src/pages/tasks/index.test.tsx src/pages/tasks/new.test.tsx
```

Expected: PASS and the API payload assertions remain intact.

**Step 5: Commit**

```bash
git add web/admin/src/pages/tasks/index.tsx web/admin/src/pages/tasks/index.test.tsx web/admin/src/pages/tasks/new.tsx web/admin/src/pages/tasks/new.test.tsx web/admin/src/styles/layout.css
git commit -m "feat: restyle inspection task pages"
```

### Task 6: Restyle the risk results page and result detail drawer

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/index.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/detail.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/detail.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing results-page tests**

Update the tests to expect:

- page title `风险结果`
- toolbar copy `本页全选` and `批量下线处置`
- action labels `查看详情`, `下线处置`, `进入整改`
- result detail tabs with Chinese labels `命中详情 / 正文快照 / 操作记录 / 字段变更`

**Step 2: Run tests to verify they fail**

Run:

```bash
npm test -- src/pages/results/index.test.tsx src/pages/results/detail.test.tsx
```

Expected: FAIL because the page still renders English actions and the old detail framing.

**Step 3: Implement the risk-result redesign**

Implement:

- summary cards and a results-focused header
- formal Chinese status/risk/disposition badges
- a dedicated batch-action toolbar styled like a control strip
- a more orderly result detail drawer with sectioned content and improved spacing

**Step 4: Re-run the results tests**

Run:

```bash
npm test -- src/pages/results/index.test.tsx src/pages/results/detail.test.tsx
```

Expected: PASS while selection, batch offline, and detail loading behavior keep working.

**Step 5: Commit**

```bash
git add web/admin/src/pages/results/index.tsx web/admin/src/pages/results/index.test.tsx web/admin/src/pages/results/detail.tsx web/admin/src/pages/results/detail.test.tsx web/admin/src/styles/layout.css
git commit -m "feat: restyle risk result workflows"
```

### Task 7: Restyle the operation logs and content rectification pages

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/rectify.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/rectify.test.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`

**Step 1: Write the failing logs/rectify tests**

Update the tests to expect:

- page titles `操作日志` and `内容整改`
- Chinese filter labels `文章编号`, `任务编号`, `操作人`
- Chinese action/button copy `查询日志`, `查看快照`, `保存整改`, `保存并提交复核`

**Step 2: Run tests to verify they fail**

Run:

```bash
npm test -- src/pages/logs/index.test.tsx src/pages/articles/rectify.test.tsx
```

Expected: FAIL because the current pages still use English labels and default presentation.

**Step 3: Implement the logs and rectify redesign**

Build:

- a page header and formal filter strip for logs
- audit-style table headers and a snapshot dialog that reads like a request archive
- a split rectify page with original content cards and a formal editing form
- submission buttons and helper copy that match the approved Chinese tone

**Step 4: Re-run the logs/rectify tests**

Run:

```bash
npm test -- src/pages/logs/index.test.tsx src/pages/articles/rectify.test.tsx
```

Expected: PASS and the existing data-handling behavior remains unchanged.

**Step 5: Commit**

```bash
git add web/admin/src/pages/logs/index.tsx web/admin/src/pages/logs/index.test.tsx web/admin/src/pages/articles/rectify.tsx web/admin/src/pages/articles/rectify.test.tsx web/admin/src/styles/layout.css
git commit -m "feat: restyle audit and rectification pages"
```

### Task 8: Run the full admin verification sweep and polish residual copy

**Files:**
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/routes.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/App.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/keywords/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/tasks/new.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/results/detail.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/logs/index.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/pages/articles/rectify.tsx`
- Modify: `/home/wwwroot/article_sentinel/web/admin/src/styles/layout.css`
- Test: `/home/wwwroot/article_sentinel/web/admin/src/App.test.tsx`
- Test: `/home/wwwroot/article_sentinel/web/admin/src/pages/**/*.test.tsx`

**Step 1: Run the focused admin test suite**

Run:

```bash
npm test -- src/App.test.tsx src/pages/keywords/index.test.tsx src/pages/tasks/index.test.tsx src/pages/tasks/new.test.tsx src/pages/results/index.test.tsx src/pages/results/detail.test.tsx src/pages/logs/index.test.tsx src/pages/articles/rectify.test.tsx
```

Expected: all page tests pass under the new shell.

**Step 2: Fix any remaining copy or accessibility regressions**

Adjust:

- stale English strings
- outdated aria labels used by tests
- spacing or badge inconsistencies discovered during the verification run

**Step 3: Run the full admin test suite**

Run:

```bash
npm test
```

Expected: PASS for the full `web/admin` test suite.

**Step 4: Run the production build**

Run:

```bash
npm run build
```

Expected: PASS and generate a production-ready bundle for the redesigned admin UI.

**Step 5: Commit**

```bash
git add web/admin/src web/admin/package.json web/admin/package-lock.json
git commit -m "feat: finalize formal admin dashboard restyle"
```

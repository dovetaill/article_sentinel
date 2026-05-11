# Admin Light Shell Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Repair the admin shell so only the sidebar remains dark while the header, tabs, content area, cards, filters, and tables are all light surfaces.

**Architecture:** Keep the existing Umi + ProLayout shell, but make the visual contract explicit in three layers: ProLayout settings, runtime Ant Design light theme configuration, and shared CSS surface classes in `src/global.less`. Lock the behavior with focused Vitest contract tests before production changes.

**Tech Stack:** Umi Max, React 18, Ant Design 5, ProLayout, ProComponents, Vitest, Testing Library, Less

---

### Task 1: Add shell contract tests

**Files:**
- Create: `web/admin/src/layouts/BasicLayout.test.tsx`
- Test: `web/admin/src/layouts/BasicLayout.test.tsx`

**Step 1: Write the failing test**

Write a test that renders `BasicLayout` with mocked Umi hooks and a mocked `@ant-design/pro-layout` wrapper. Assert:
- the shell uses side layout
- the sidebar is configured with the dark nav theme
- the workspace body uses a light-surface class
- the page tabs container uses a light-surface class

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/layouts/BasicLayout.test.tsx`
Expected: FAIL because the current layout still uses mixed layout and does not expose the planned light shell classes.

**Step 3: Write minimal implementation**

Update `web/admin/config/defaultSettings.ts` and `web/admin/src/layouts/BasicLayout.tsx` so the rendered shell satisfies the contract.

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/layouts/BasicLayout.test.tsx`
Expected: PASS

### Task 2: Add visual token contract tests

**Files:**
- Create: `web/admin/src/styles/admin-theme.ts`
- Create: `web/admin/src/styles/visual-tokens.test.ts`
- Modify: `web/admin/src/app.tsx`
- Modify: `web/admin/src/global.less`

**Step 1: Write the failing test**

Write a test that asserts:
- sidebar token is dark
- page/content tokens are light
- surface/card/table tokens are white or near-white
- runtime theme uses `defaultAlgorithm`
- `global.less` does not define dark content/surface tokens

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/styles/visual-tokens.test.ts`
Expected: FAIL because the token module does not exist yet and the current contract is not explicit.

**Step 3: Write minimal implementation**

Add a light runtime Ant Design theme module, wrap the app root with `ConfigProvider`, and align `global.less` CSS variables with the same shell contract.

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/styles/visual-tokens.test.ts`
Expected: PASS

### Task 3: Add `/inspection/tasks` light-surface page contract test

**Files:**
- Modify: `web/admin/src/pages/Inspection/TaskList/index.test.tsx`
- Modify: `web/admin/src/pages/Inspection/TaskList/index.tsx`
- Modify: `web/admin/src/global.less`

**Step 1: Write the failing test**

Extend `TaskList` tests so they assert:
- summary cards use a shared light-surface class
- filter toolbar card uses a shared light-surface class
- table wrapper uses a shared light-surface class

**Step 2: Run test to verify it fails**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Inspection/TaskList/index.test.tsx`
Expected: FAIL because the wrappers/classes do not exist yet.

**Step 3: Write minimal implementation**

Add shared surface classes in `global.less` and apply them to the task list page markup.

**Step 4: Run test to verify it passes**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/pages/Inspection/TaskList/index.test.tsx`
Expected: PASS

### Task 4: Propagate shared light-surface classes to sibling list pages

**Files:**
- Modify: `web/admin/src/pages/Inspection/ResultList/index.tsx`
- Modify: `web/admin/src/pages/Rules/KeywordList/index.tsx`
- Modify: `web/admin/src/pages/Rules/CategoryList/index.tsx`
- Modify: `web/admin/src/pages/Audit/OperationLogList/index.tsx`
- Modify: `web/admin/src/pages/Content/ArticleList/index.tsx`
- Modify: `web/admin/src/pages/Inspection/TaskCreate/index.tsx`
- Optionally modify detail pages if their shared surface classes benefit from the same cleanup

**Step 1: Write/confirm the test boundary**

Use the shell and task-list contract tests as the regression boundary. No extra page-specific tests are required unless a page needs unique surface behavior.

**Step 2: Apply minimal implementation**

Reuse the same shared classes for summary strips, filter cards, and table shells so sibling list pages inherit the corrected aesthetic without page-specific one-off styles.

**Step 3: Run focused tests**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/layouts/BasicLayout.test.tsx src/styles/visual-tokens.test.ts src/pages/Inspection/TaskList/index.test.tsx`
Expected: PASS

### Task 5: Verify full requested commands and build

**Files:**
- No code changes required unless verification uncovers regressions

**Step 1: Run targeted verification**

Run the closest valid replacements for the requested commands in the current codebase:
- `cd /home/wwwroot/article_sentinel/web/admin && npm test -- src/layouts/BasicLayout.test.tsx src/styles/visual-tokens.test.ts src/pages/Inspection/TaskList/index.test.tsx`

**Step 2: Run build**

Run: `cd /home/wwwroot/article_sentinel/web/admin && npm run build`
Expected: exit 0

**Step 3: Report results honestly**

If a requested path like `src/App.test.tsx` does not exist in the current Umi codebase, report the exact replacement command that was used and why.

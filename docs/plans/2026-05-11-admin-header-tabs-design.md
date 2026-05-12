# Admin Shell Cleanup and Task-Centric Inspection Design

**Date:** 2026-05-11

## Goal

Bring the current `web/admin` workspace back to a focused operator flow:

- remove low-value shell chrome from the top area
- allow every tab, including `检测任务`, to be closed
- restore inspection review to a task-centric flow
- make `任务详情` the place where users review and operate on task results

The stack stays on Umi + React + ProLayout + Ant Design. This is a shell and workflow correction, not a backend rewrite.

## Current Code Reality

The active admin frontend already has the right technical foundation, but the information architecture and shell details drifted away from the intended operator workflow.

Relevant files:

- `web/admin/src/layouts/BasicLayout.tsx`
- `web/admin/src/components/PageTabs/index.tsx`
- `web/admin/src/components/PageTabs/store.ts`
- `web/admin/src/components/PageTabs/route-meta.ts`
- `web/admin/config/routes.ts`
- `web/admin/src/global.less`
- `web/admin/src/pages/Inspection/TaskList/index.tsx`
- `web/admin/src/pages/Inspection/TaskDetail/index.tsx`
- `web/admin/src/pages/Inspection/ResultList/index.tsx`
- `web/admin/src/pages/Audit/OperationLogList/index.tsx`
- `web/admin/src/pages/LegacyRedirect/index.tsx`

Observed problems:

1. The shell still renders too many non-essential controls: search, fullscreen, notification badge, user-center entry, and collapsers that the current workflow does not need.
2. `PageContainer` still renders a Pro breadcrumb/page-header layer that adds visual noise and surfaces broken breadcrumb separators.
3. `检测任务` is treated as a pinned tab because the base route resolves to `closable: false`.
4. `/inspection/tasks` no longer gives users a clear way to enter task detail and continue the workflow from there.
5. `/inspection/results` exists as a standalone workspace even though the actual business flow is task output review, not a stable top-level domain.
6. `任务详情` already has a `命中结果` tab, but it only shows a lightweight preview list instead of the full result-operation workspace.

## Accepted Direction

### Shell Cleanup

Keep the dark sidebar and light workspace, but simplify the right-side shell:

- remove the ProLayout sider collapse button
- remove the custom header collapse button
- remove the top search field
- remove the fullscreen button
- remove the notification badge/action
- remove the `个人中心` entry
- keep only essential user/session UI, such as current user text and `退出登录`
- remove the extra `PageContainer` breadcrumb/page-header strip from business pages

The header should become a calm light bar that supports the workspace rather than competing with it.

### Tabs

All tabs are closable, including `检测任务`.

Rules:

- clicking a route opens or activates a tab
- closing the active tab falls back left, then right, then the hidden empty workspace route
- closing all tabs is allowed
- when every tab is closed, the app navigates to the hidden `/workspace` empty state

No business tab is pinned by default.

### Inspection Workflow

Restore a task-centric flow:

- `检测任务列表` remains the entry page
- task rows must expose a clear path into `任务详情`
- `任务详情` becomes the main inspection workbench
- the `命中结果` tab inside `任务详情` is upgraded to a full result table with pagination, batch offline action, and article/rectify jumps
- `/inspection/results` is no longer a standalone menu page and should only exist as a hidden compatibility redirect

Canonical operator flow:

`检测任务列表 -> 任务详情 -> 命中结果 -> 文稿详情 / 内容整改`

## Route and Metadata Decisions

### `/inspection/tasks`

- stays as the default admin landing page
- remains in the menu
- becomes closable like every other business tab

### `/inspection/tasks/:taskId`

- remains the main detail route
- owns the task summary, snapshots, logs, and embedded result workspace
- supports `?tab=results` for direct entry into the result view
- may also carry result pagination state in query parameters

### `/inspection/results`

- stays as a route only for compatibility
- becomes `hiddenInMenu: true`
- becomes `opensTab: false`
- redirects to `/inspection/tasks/:taskId?tab=results` when `task_id` is present
- redirects to `/inspection/tasks` when no task context is available

### Legacy routes

- `/tasks/:taskId/results` should now resolve to `/inspection/tasks/:taskId?tab=results`
- `/results?task_id=77` should resolve to `/inspection/tasks/77?tab=results`

## Task Detail Design

`任务详情` should keep the current top-level summary and side information cards, but the first tab becomes a true workbench.

### Results Tab

The `命中结果` tab should provide:

- result table with the current task bound as context
- selectable rows
- batch offline action
- per-row actions:
  - `查看详情`
  - `下线处置`
  - `进入整改`
- URL-backed return targets so article detail and rectify pages can navigate back into the task result context

This tab replaces the current lightweight list preview.

### Other Tabs

Keep the existing supporting tabs:

- `规则快照`
- `请求快照`
- `关联日志`

These remain inside the same detail route so the operator stays in one task workspace.

## Task List Design

`/inspection/tasks` should feel like the control tower for inspection batches.

Required corrections:

- task number should be clickable into `任务详情`
- action column should expose a clear `查看任务` action
- delete remains limited to pending/failed tasks
- the layout should stay light and compact, but the main behavioral fix is restoring the missing detail entry

## Audit Link Design

Audit records should point to the task-centric workspace:

- article links still go to `文稿详情`
- task links now go to `/inspection/tasks/:taskId?tab=results`

This keeps investigation and remediation anchored to the task that produced the hit.

## Page Header Removal Strategy

Every business page already uses `PageContainer`, but `title={false}` alone still leaves Pro page-header structure in play.

Accepted fix:

- add `pageHeaderRender={false}` to business `PageContainer` usage
- use the custom shell header and page-level cards instead of Pro page header chrome
- keep this explicit in code rather than relying on CSS-only hiding

A small defensive CSS rule may still hide residual layout artifacts, but the primary fix should be component-level.

## Visual Direction

The new shell should feel more like a workbench and less like a demo admin template.

### Header

- white background
- thin bottom border
- compact vertical rhythm
- no extra controls beyond route title and session/logout area

### Tabs

- white rail
- compact rectangular tabs
- all tabs closeable
- active tab uses blue emphasis without heavy fill blocks

### Task Detail Results Area

- white panel with tight toolbar
- clear selection state
- table-first layout, not card pile-up
- action buttons remain text-light and enterprise-like

## Out of Scope

This change does not:

- add new backend APIs
- redesign the login flow
- change task creation request semantics
- reintroduce the old drawer-based task detail
- migrate the app away from Umi/ProLayout

## Test Strategy

Use focused contract tests instead of screenshots.

1. shell test for the cleaned header and absence of removed controls
2. route-meta and tab-store tests for closable tabs and hidden redirect routes
3. task list test for task-detail entry
4. task detail test for the embedded full result workspace
5. redirect tests for `/inspection/results` and legacy `/tasks/:id/results`
6. audit/article return-path tests so operator navigation stays intact

## Result

After this change:

- the header is quieter and more purposeful
- the broken breadcrumb/page-header layer is gone
- every business tab can be closed
- task review returns to the intended `任务 -> 结果 -> 文稿/整改` flow
- `/inspection/results` stops acting like a standalone business page

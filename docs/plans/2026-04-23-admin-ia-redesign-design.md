# Admin Information Architecture and Shell Redesign

## Context

The current admin app has three structural mismatches with the intended product flow:

1. `/articles` currently behaves like an inspection-result aggregation view instead of a real article center.
2. `风险结果` exists as a first-class navigation item even though it is operationally a task output, not a stable domain object.
3. The shell and page hierarchy still diverge from the intended `shadcn-admin` style workbench with a collapsible sidebar, fixed header, organization switcher, and tighter content frame.

The current implementation also hardcodes `orgid=100` through the frontend and test fixtures, uses free-form keyword categories instead of managed category data, and leaves local dev requests broken because the Vite dev server has no `/api` proxy.

## Product Goals

This redesign aligns the admin app around stable business objects and organization-scoped workflows.

- Make every dataset organization-scoped.
- Restore `/articles` as the real article center backed by the article tables.
- Move risk review into a task-centric results workspace at `/tasks/:id/results`.
- Introduce managed keyword categories as first-class data.
- Replace raw organization IDs in the UI with organization names.
- Bring the shell closer to `shadcn-admin` interaction patterns without copying irrelevant demo semantics.
- Make local development work end to end by routing frontend API calls to the backend and starting queue services together.

## Scope Decisions

### Confirmed Information Architecture

Top-level navigation becomes:

- `规则中心`
- `检测任务`
- `文稿中心`
- `操作日志`

Secondary structure becomes:

- `规则中心`
  - `规则分类`
  - `关键词规则`
- `检测任务`
  - `任务列表`
  - `新建任务`
  - `任务结果页` via `/tasks/:id/results`
- `文稿中心`
  - `真实文稿列表`
  - `真实文稿详情`
  - `内容整改`
- `操作日志`
  - global log search

### Confirmed Workflow

Canonical business flow:

`机构 -> 规则分类 -> 关键词规则 -> 检测任务 -> 任务结果 -> 真实文稿详情/整改`

### Confirmed Task Creation Behavior

Task creation is simplified to the main use case only.

- Current organization is selected by the active organization context.
- User selects keyword rules with a searchable multi-select control.
- User does not input article ID or title filters.
- New tasks always scan all articles in the current organization where `state=9`.

### Confirmed Results Workspace Behavior

`风险结果` is removed as a top-level navigation item.

Task rows instead expose a `运行结果` action that navigates to `/tasks/:id/results`, where operators can:

- view article detail
- enter rectify flow
- offline article
- ignore or mark processed
- batch dispose
- inspect task-related logs

### Confirmed Organization Model

All business data is organization-scoped, including:

- categories
- keywords
- tasks
- results
- real articles
- operation logs
- field change logs

The header includes an organization switcher. For now the organization list is backed by a real table but initially contains one seeded option:

- `29 / 一县一端`

The switcher is still required now so the product shape is correct before multi-organization rollout.

### Confirmed Article Center Scope

The real article center only needs:

- article list
- article detail
- links into inspect and rectify workflows

This redesign does not add full publishing CRUD in the article center.

## Data Model Design

## 1. Organization Master Data

Introduce a formal organization table instead of raw numeric entry in forms.

### Table: `xt_chuangqi_org`

Suggested fields:

- `id`
- `name`
- `cateid`
- `enabled`
- `sort`
- `create_at`
- `update_at`

Rationale:

- `enabled` is useful for hiding inactive organizations from the switcher.
- `sort` gives predictable ordering in dropdowns.
- `cateid` is preserved because the user explicitly requested it.
- The UI always renders `name`; the backend stores `id`.

## 2. Keyword Category Master Data

Keyword categories become first-class managed data instead of free-form strings.

### Table: `xt_article_inspect_categories`

Suggested fields:

- `id`
- `orgid`
- `name`
- `code`
- `enabled`
- `sort`
- `creator_id`
- `creator_name`
- `updater_id`
- `updater_name`
- `create_at`
- `update_at`

Rationale:

- Categories are organization-specific.
- `code` prevents future ambiguity when category names change.
- audit columns align with existing inspect tables.

## 3. Keyword Rule Model Changes

`xt_article_inspect_keywords` should stop treating category as free text.

Recommended change:

- replace or deprecate the current free-form `category` usage
- add `category_id`
- expose `category_name` in list/detail responses

The UI must only select from managed categories for the current organization.

## 4. Article Center Boundary

The real article center is backed by the existing article source tables, not inspection result tables.

Primary source tables already present in the codebase:

- `xt_article`
- `xt_article_info`

This means:

- `/articles` lists real articles from `xt_article`
- `/articles/:id` loads real article detail and enriches it with inspection history
- inspection results remain output data from task execution, not the source of truth for article browsing

## 5. Inspection Output Boundary

Inspection output continues to be backed by existing inspect tables:

- `xt_article_inspect_tasks`
- `xt_article_inspect_task_keywords`
- `xt_article_inspect_results`
- `xt_article_inspect_result_hits`
- `xt_article_inspect_actions`
- `xt_article_inspect_operation_logs`
- `xt_article_inspect_field_change_logs`

These power `/tasks/:id/results` and associated batch operations.

## Backend API Design

## 1. Organization APIs

Add organization read APIs for the header switcher.

Recommended initial route:

- `GET /api/v1/article-inspect/orgs`

Response returns enabled organizations only, ordered by `sort` then `id`.

## 2. Category APIs

Add full CRUD for category management.

Recommended routes:

- `GET /api/v1/article-inspect/categories?orgid=<id>`
- `POST /api/v1/article-inspect/categories`
- `GET /api/v1/article-inspect/categories/{id}?orgid=<id>`
- `PUT /api/v1/article-inspect/categories/{id}`
- `PATCH /api/v1/article-inspect/categories/{id}/status`
- `DELETE /api/v1/article-inspect/categories/{id}?orgid=<id>`

All routes require organization scoping.

## 3. Keyword APIs

Keep the current keyword domain but update request and response shapes.

Requirements:

- create and update accept `category_id`
- list and detail return both `category_id` and `category_name`
- all queries remain organization-scoped

## 4. Task APIs

Task creation becomes narrower and more explicit.

Create request should keep:

- `orgid`
- `keyword_ids`
- optional note or remark if retained

Create request should remove:

- `article_id`
- `title_like`
- ad hoc article-scoping fields from the main form

Execution rule:

- backend scans all articles in current `orgid` with `state=9`

## 5. Task Results APIs

Task results move to a dedicated workspace.

The existing result services can be reused, but the frontend should consume them under a task-centric page model:

- task summary
- task-specific result list
- task-specific hit article view
- task-specific operation logs
- task-specific batch operations

If needed, add thin task-results query endpoints rather than forcing the frontend to reconstruct this from unrelated screens.

## 6. Article Center APIs

Introduce real article center APIs using `xt_article` and `xt_article_info`.

Recommended routes:

- `GET /api/v1/article-inspect/articles?orgid=<id>&page=<n>&page_size=<n>&state=<optional>&keyword=<optional>`
- `GET /api/v1/article-inspect/articles/{id}?orgid=<id>`

Detail responses can enrich with latest inspect summary, recent task references, and linked operation history.

## Frontend Shell Design

The shell should align closely with the interaction model from `shadcn-admin`:

- collapsible sidebar on the left
- fixed header on top
- sidebar trigger on the header left
- organization switcher on the header right
- user menu on the header right
- inset content frame with denser spacing and less decorative page chrome

This does not mean copying the demo app's route tree. It means reusing its shell behavior and page rhythm for this product's actual business pages.

### Header

Left side:

- sidebar trigger
- current page title
- optional lightweight breadcrumb

Right side:

- organization switcher
- current user menu
- logout entry

### Sidebar

Sections:

- `规则中心`
- `检测任务`
- `文稿中心`
- `操作日志`

Nested items only where the product actually has a grouped domain.

### Page Header Policy

Avoid large repeated hero headers. The fixed shell header carries global frame context. Pages should open directly into tools, filters, tables, and cards.

## Page-Level UX Design

## 1. Category Management Page

Purpose:

- manage organization-scoped keyword categories

Layout:

- page title and `新建分类` action
- filter strip for category name and status
- table with name, code, status, sort, update time, actions
- modal or drawer for create/edit

## 2. Keyword Rules Page

Purpose:

- manage rules under selected organization and category

Layout:

- page title and `新建规则` action
- filter strip for keyword name, category, match type, risk level, status
- table with keyword name, category, match type, risk level, suggested action, status, update time, actions
- create/edit modal with organization read-only and category dropdown

## 3. Task List Page

Purpose:

- start and monitor inspection jobs

Layout:

- page title and `新建任务`
- table with task number, organization, create time, status, selected rule count, scanned articles, hit articles, hit count, actions
- primary row action is `运行结果`

## 4. New Task Page

Purpose:

- launch inspection against all online articles in the current organization

Layout:

- read-only organization field showing current organization name
- searchable multi-select for keyword rules
- optional note field if retained
- explanatory helper text: scans all current organization articles where `state=9`

No article ID or title filters appear on the main task form.

## 5. Task Results Page `/tasks/:id/results`

Purpose:

- operate on outputs of one inspection task

Layout:

- top summary with back button, task number, status, scanned articles, hit articles, hit count, duration
- internal sections or tabs for:
  - `风险结果`
  - `命中文稿`
  - `批量处置`
  - `操作日志`

Actions available from rows or bulk toolbars:

- view article detail
- enter rectify
- offline article
- ignore or mark processed
- batch dispose

## 6. Real Article Center

Purpose:

- browse the actual article source of truth
- inspect article details and jump into inspection workflows

### Article List

Columns:

- title
- article ID
- state
- publish time
- latest inspect risk
- latest task
- action to view detail

### Article Detail

Sections:

- article basics
- body snapshot
- latest inspect summary
- hit records
- operation history
- primary actions to rectify, inspect latest task result, or offline

## 7. Global Logs Page

Purpose:

- audit operations across tasks and articles within the current organization

Filters:

- organization
- task number
- article ID
- action type
- operator
- time range

Rows can link back into task or article detail pages.

## Runtime and Developer Experience

## 1. Dev Proxy

The frontend currently sends `/api/...` requests to the Vite dev server and gets 404 responses because there is no proxy configuration.

The Vite server must proxy `/api` to the backend app during local development.

## 2. Queue Runtime

Queue infrastructure already exists in the repository, but local startup does not bring all required processes together.

Recommended local runtime support:

- server
- worker
- scheduler
- admin dev server

`make dev` should be expanded to start the whole stack for day-to-day development, while narrower commands remain available for targeted debugging.

## Non-Goals for This Redesign

This redesign does not include:

- full article publishing CRUD in the article center
- multi-organization authorization logic beyond reading and switching the current organization context from the new organization table
- blind 1:1 reproduction of every demo page from `shadcn-admin`
- preserving incorrect route semantics for compatibility if they conflict with the corrected product model

## Migration Notes

- Existing `orgid=100` assumptions in frontend pages and tests must be migrated to organization `29` and then centralized behind an organization context helper.
- Existing `/articles` tests and services that depend on result aggregation must be redirected into task-results semantics or renamed to an inspection-specific view.
- Existing keyword fixtures and request payloads must be updated to use managed categories instead of free-form category strings.
- Existing frontend forms must stop exposing raw organization IDs as numeric text inputs.

## Acceptance Criteria

This redesign is successful when all of the following are true:

- the shell matches the intended `shadcn-admin` structure closely enough that the app reads like a modern, collapsible admin workbench
- every inspect-facing page reads and writes data using the active organization context
- categories are managed data with their own page and APIs
- keyword rules choose categories from a controlled data source
- task creation only launches scans for all current-organization online articles (`state=9`)
- `风险结果` no longer appears as a top-level nav item
- `/tasks/:id/results` becomes the main risk-operation workspace
- `/articles` becomes the real article center backed by article source tables
- local frontend API calls work in dev through a Vite proxy
- local development can start the app together with queue execution services

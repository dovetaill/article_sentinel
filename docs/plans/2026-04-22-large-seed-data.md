# Large Seed Data Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add and import a larger MySQL seed dataset so the article inspection UI and API have enough local data for realistic testing.

**Architecture:** Keep the current small demo seed untouched and add a second SQL seed file for volume testing. The new file creates minimal runtime-compatible article tables if needed, deletes only fixed seed ID ranges, then inserts a broader mix of articles, tasks, results, hits, and audit history.

**Tech Stack:** MySQL 5.7 SQL, Docker MySQL container, Go runtime schema expectations, existing `articleinspect` admin UI flows.

---

### Task 1: Document the large-seed dataset contract

**Files:**
- Create: `/home/wwwroot/article_sentinel/.worktrees/large-seed-data/docs/plans/2026-04-22-large-seed-data-design.md`
- Create: `/home/wwwroot/article_sentinel/.worktrees/large-seed-data/docs/plans/2026-04-22-large-seed-data.md`

**Step 1: Write the design doc**

Capture:

- why the project needs a second seed file
- exact seeded ID ranges
- which tables must be created or populated
- the intended mix of statuses for UI testing

**Step 2: Save the implementation plan**

Document:

- the target SQL file path
- the import command
- the post-import verification queries

**Step 3: Sanity check the docs**

Run:

```bash
sed -n '1,220p' docs/plans/2026-04-22-large-seed-data-design.md
sed -n '1,260p' docs/plans/2026-04-22-large-seed-data.md
```

Expected: both files exist and describe the approved dataset shape.

### Task 2: Generate the large SQL seed file

**Files:**
- Create: `/home/wwwroot/article_sentinel/.worktrees/large-seed-data/scripts/article_inspection_seed_large.sql`
- Reference: `/home/wwwroot/article_sentinel/.worktrees/large-seed-data/scripts/article_inspection_seed.sql`
- Reference: `/home/wwwroot/article_sentinel/.worktrees/large-seed-data/internal/modules/articleinspect/model.go`

**Step 1: Define runtime-compatible article tables**

Write SQL that creates:

- `xt_article` with the columns used by the Go services
- `xt_article_info` with `article_id`, `body`, and `update_at`

Expected: importing the seed works even in a fresh local database where upstream article tables are missing.

**Step 2: Delete only the large-seed ranges**

Write delete statements for the fixed article, keyword, task, result, hit, action, operation log, and field change log ranges.

Expected: repeated imports are idempotent and leave unrelated data intact.

**Step 3: Insert realistic volume data**

Insert:

- 48 articles and 48 article bodies
- 12 keywords with varied scopes and risk levels
- 12 tasks covering `pending`, `running`, `success`, `partial_success`, and `failed`
- 36 results and enough hits, actions, and logs for detail views

Expected: local UI pages have enough rows for filtering and at least the result page can paginate.

### Task 3: Import the seed into local MySQL

**Files:**
- Modify: `/home/wwwroot/article_sentinel/.worktrees/large-seed-data/scripts/article_inspection_seed_large.sql`

**Step 1: Import through the local MySQL container**

Run:

```bash
docker exec -i article-sentinel-mysql mysql -uroot -p'Holmes64125135' article_sentinel < scripts/article_inspection_seed_large.sql
```

Expected: import succeeds without SQL errors.

**Step 2: Verify seeded counts**

Run:

```bash
docker exec article-sentinel-mysql mysql -uroot -p'Holmes64125135' -D article_sentinel -Nse "
SELECT 'xt_article', COUNT(*) FROM xt_article WHERE id BETWEEN 9100001 AND 9100048
UNION ALL
SELECT 'xt_article_info', COUNT(*) FROM xt_article_info WHERE article_id BETWEEN 9100001 AND 9100048
UNION ALL
SELECT 'xt_article_inspect_keywords', COUNT(*) FROM xt_article_inspect_keywords WHERE id BETWEEN 9101001 AND 9101012
UNION ALL
SELECT 'xt_article_inspect_tasks', COUNT(*) FROM xt_article_inspect_tasks WHERE id BETWEEN 9202001 AND 9202012
UNION ALL
SELECT 'xt_article_inspect_results', COUNT(*) FROM xt_article_inspect_results WHERE id BETWEEN 9300001 AND 9300036
UNION ALL
SELECT 'xt_article_inspect_result_hits', COUNT(*) FROM xt_article_inspect_result_hits WHERE id BETWEEN 9400001 AND 9499999
UNION ALL
SELECT 'xt_article_inspect_actions', COUNT(*) FROM xt_article_inspect_actions WHERE id BETWEEN 9500001 AND 9599999
UNION ALL
SELECT 'xt_article_inspect_operation_logs', COUNT(*) FROM xt_article_inspect_operation_logs WHERE id BETWEEN 9600001 AND 9699999
UNION ALL
SELECT 'xt_article_inspect_field_change_logs', COUNT(*) FROM xt_article_inspect_field_change_logs WHERE id BETWEEN 9700001 AND 9799999;
"
```

Expected: counts match the generated dataset volume.

**Step 3: Spot-check representative records**

Run:

```bash
docker exec article-sentinel-mysql mysql -uroot -p'Holmes64125135' -D article_sentinel -e "
SELECT id, title, state FROM xt_article WHERE id BETWEEN 9100001 AND 9100005 ORDER BY id;
SELECT id, task_no, status, hit_articles FROM xt_article_inspect_tasks WHERE id BETWEEN 9202001 AND 9202005 ORDER BY id;
SELECT id, article_id, risk_level, disposition_status FROM xt_article_inspect_results WHERE id BETWEEN 9300001 AND 9300008 ORDER BY id;
"
```

Expected: records show the intended variety of states, risks, and dispositions.

### Task 4: Keep the branch reviewable

**Files:**
- Create: `/home/wwwroot/article_sentinel/.worktrees/large-seed-data/scripts/article_inspection_seed_large.sql`
- Create: `/home/wwwroot/article_sentinel/.worktrees/large-seed-data/docs/plans/2026-04-22-large-seed-data-design.md`
- Create: `/home/wwwroot/article_sentinel/.worktrees/large-seed-data/docs/plans/2026-04-22-large-seed-data.md`

**Step 1: Review diff**

Run:

```bash
git status --short
git diff -- docs/plans/2026-04-22-large-seed-data-design.md docs/plans/2026-04-22-large-seed-data.md scripts/article_inspection_seed_large.sql
```

Expected: only the new docs and seed SQL appear in the branch for this task.

**Step 2: Commit when satisfied**

Run:

```bash
git add docs/plans/2026-04-22-large-seed-data-design.md docs/plans/2026-04-22-large-seed-data.md scripts/article_inspection_seed_large.sql
git commit -m "chore: add large article inspection seed data"
```

Expected: the branch contains a single focused commit for the large local dataset.

# Large Seed Data Design

## Goal

Create a local-only MySQL seed file that populates enough article inspection data to exercise the admin UI and API without relying on upstream production tables.

## Scope

- Add a new SQL seed file dedicated to larger local test data.
- Create minimal `xt_article` and `xt_article_info` tables if they do not already exist.
- Seed enough articles, keywords, tasks, results, hits, actions, operation logs, and field change logs for paging and filter testing.
- Keep all inserted records inside fixed ID ranges so repeated imports are safe and existing user data is left alone.

## Recommended Approach

Use a separate file `scripts/article_inspection_seed_large.sql` rather than expanding the small demo seed. The existing seed stays easy to understand, while the new file focuses on volume, pagination, and state variety.

## Data Shape

- `xt_article`: 48 articles with mixed states, timestamps, titles, summaries, and keywords.
- `xt_article_info`: matching body HTML for every seeded article.
- `xt_article_inspect_keywords`: 12 rules covering different scopes, risk levels, and suggested actions.
- `xt_article_inspect_tasks`: 12 tasks across `pending`, `running`, `success`, `partial_success`, and `failed`.
- `xt_article_inspect_results`: 36 results with mixed `pending`, `ignored`, `processed`, `offlined`, and `republished` dispositions.
- `xt_article_inspect_result_hits`: multiple hits per result so detail pages are non-empty.
- `xt_article_inspect_actions`, `xt_article_inspect_operation_logs`, `xt_article_inspect_field_change_logs`: enough history to validate task detail, result detail, and log pages.

## Table Strategy

The runtime code expects a local `xt_article` table with datetime `publish_at_time` and `update_at`, and a local `xt_article_info` table keyed by `article_id`. The large seed will create that minimal runtime-compatible schema only when those tables are absent.

## Safety Rules

- Use fixed ranges only:
  - articles: `9100001-9100048`
  - keywords: `9101001-9101012`
  - tasks: `9202001-9202012`
  - results: `9300001-9300036`
  - hits/actions/logs: `94xxxxxx-97xxxxxx`
- Delete only those ranges before re-inserting.
- Leave any non-seed rows untouched.

## Import and Verification

Import through the local MySQL container with:

```bash
docker exec -i article-sentinel-mysql mysql -uroot -p'${MYSQL_ROOT_PASSWORD}' article_sentinel < scripts/article_inspection_seed_large.sql
```

Then verify row counts and spot-check sample records in the seeded ranges.

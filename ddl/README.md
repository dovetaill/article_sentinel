## Source Article DDL

This directory keeps the original upstream article table definitions as reference inputs
for the inspection module.

- `xt_article.sql`: source article metadata table
- `xt_article_info.sql`: source article body/details table

Notes:

- These files are reference snapshots, not project migrations.
- The executable schema for the inspection module remains in `migrations/20260420_01_article_inspection.sql`.
- When the upstream article schema changes, update these snapshots in place so the repo keeps a trackable history.

## Source Article DDL

This directory keeps the original upstream article table definitions as reference inputs
for the inspection module.

- `xt_article.sql`: source article metadata table
- `xt_article_info.sql`: source article body/details table
- `xt_article_inspection.sql`: article inspection MySQL schema snapshot

Notes:

- These files are reference snapshots, not project migrations.
- The executable schema for the inspection module remains in `migrations/20260420_01_article_inspection.sql`.
- `xt_article_inspection.sql` is a tracked reference copy of that migration for teams that want all MySQL DDL snapshots under one directory.
- When the upstream article schema changes, update these snapshots in place so the repo keeps a trackable history.

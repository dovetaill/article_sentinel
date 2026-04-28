SET @has_category_code_index := (
  SELECT COUNT(*)
  FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'xt_article_inspect_categories'
    AND index_name = 'uk_org_code'
);

SET @drop_category_code_index := IF(
  @has_category_code_index > 0,
  'ALTER TABLE `xt_article_inspect_categories` DROP INDEX `uk_org_code`',
  'SELECT 1'
);

PREPARE article_inspect_drop_category_code_index FROM @drop_category_code_index;
EXECUTE article_inspect_drop_category_code_index;
DEALLOCATE PREPARE article_inspect_drop_category_code_index;

SET @has_category_code_column := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'xt_article_inspect_categories'
    AND column_name = 'code'
);

SET @drop_category_code_column := IF(
  @has_category_code_column > 0,
  'ALTER TABLE `xt_article_inspect_categories` DROP COLUMN `code`',
  'SELECT 1'
);

PREPARE article_inspect_drop_category_code_column FROM @drop_category_code_column;
EXECUTE article_inspect_drop_category_code_column;
DEALLOCATE PREPARE article_inspect_drop_category_code_column;

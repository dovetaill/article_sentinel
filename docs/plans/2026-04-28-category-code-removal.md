# Category Code Removal Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Remove the unused category `code` field from the admin UI, backend category APIs, and database schema so categories are managed only by `id`, `name`, `enabled`, and `sort`.

**Architecture:** The change is a hard removal across three seams: React admin category management, Go article-inspection category contracts, and MySQL migration SQL. The safest path is to drive each seam with failing tests first, then make the smallest code changes needed to remove `code` while preserving category behavior based on `category_id`.

**Tech Stack:** React 18, Ant Design Pro, TypeScript, Vitest, Go, GORM, Huma, raw SQL migrations, SQLite-backed Go tests

---

### Task 1: Remove category code from the admin UI contract

**Files:**
- Modify: `web/admin/src/pages/categories/index.test.tsx`
- Modify: `web/admin/src/pages/categories/index.tsx`
- Modify: `web/admin/src/services/categories.ts`
- Modify: `web/admin/src/pages/keywords/index.test.tsx`

**Step 1: Write the failing test**

Update `web/admin/src/pages/categories/index.test.tsx` so the create/edit flow expects no `分类编码` field and no `code` property in the mutation payload.

```tsx
expect(screen.queryByText('分类编码')).not.toBeInTheDocument();

await waitFor(() => {
  expect(mockedCreateCategory).toHaveBeenCalledWith(expect.objectContaining({
    orgid: 29,
    name: '高频违规',
    enabled: true,
    sort: 20,
  }));
  expect(mockedCreateCategory).not.toHaveBeenCalledWith(expect.objectContaining({
    code: expect.anything(),
  }));
});
```

Also trim `code` from the mocked category objects in `web/admin/src/pages/keywords/index.test.tsx` so the tests describe the target contract.

**Step 2: Run test to verify it fails**

Run: `npm test -- src/pages/categories/index.test.tsx src/pages/keywords/index.test.tsx`

Expected: FAIL because the page still renders `分类编码` and still submits `code`.

**Step 3: Write minimal implementation**

Remove `code` from the category page types, table columns, modal form, and payload construction in `web/admin/src/pages/categories/index.tsx`, then remove `code` from the category service types in `web/admin/src/services/categories.ts`.

```ts
type CategoryFormValues = {
  org_name: string;
  name: string;
  enabled: boolean;
  sort?: number;
};

const payload: CategoryMutationInput = {
  orgid: currentOrgId,
  name: values.name,
  enabled: values.enabled,
  sort: values.sort ?? 0,
};
```

**Step 4: Run test to verify it passes**

Run: `npm test -- src/pages/categories/index.test.tsx src/pages/keywords/index.test.tsx`

Expected: PASS for both files.

**Step 5: Commit**

```bash
git add web/admin/src/pages/categories/index.test.tsx \
  web/admin/src/pages/categories/index.tsx \
  web/admin/src/services/categories.ts \
  web/admin/src/pages/keywords/index.test.tsx
git commit -m "refactor: remove category code from admin ui"
```

### Task 2: Remove category code from backend handlers, DTOs, and validation

**Files:**
- Modify: `internal/modules/articleinspect/articleinspect_test.go`
- Modify: `internal/modules/articleinspect/handler.go`
- Modify: `internal/modules/articleinspect/dto_categories.go`
- Modify: `internal/modules/articleinspect/service_categories.go`
- Modify: `internal/modules/articleinspect/repository_categories.go`
- Modify: `internal/modules/articleinspect/model.go`

**Step 1: Write the failing test**

Update `TestHandlerOrgCategoryAndArticleCenterContracts` in `internal/modules/articleinspect/articleinspect_test.go` so category create and update requests omit `code`, and add assertions that category response items no longer include a `code` key.

```go
created := sendArticleInspectJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/categories", map[string]any{
	"orgid":   29,
	"name":    "新增分类",
	"enabled": true,
	"sort":    10,
})

item := articleInspectDataMap(t, created.envelope.Data)
if _, ok := item["code"]; ok {
	t.Fatalf("category payload = %#v, do not want code", item)
}
```

Keep the missing-`orgid` coverage, but remove `code` from those request bodies because `orgid` should remain the only required field gate under test.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/articleinspect -run TestHandlerOrgCategoryAndArticleCenterContracts`

Expected: FAIL because the current handler/service path still requires `code` and still returns it in DTOs.

**Step 3: Write minimal implementation**

Remove `code` from the request body, DTO, normalization, repository update map, name search, and GORM model.

```go
type CreateCategoryInput struct {
	OrgID   uint64
	Name    string
	Enabled bool
	Sort    int64
}

func normalizeCategoryInput(input CreateCategoryInput) (*CreateCategoryInput, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidCategoryInput)
	}
	return &CreateCategoryInput{
		OrgID:   input.OrgID,
		Name:    name,
		Enabled: input.Enabled,
		Sort:    input.Sort,
	}, nil
}
```

Adjust category list search to:

```go
query = query.Where("name LIKE ?", like)
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/articleinspect -run TestHandlerOrgCategoryAndArticleCenterContracts`

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/articleinspect_test.go \
  internal/modules/articleinspect/handler.go \
  internal/modules/articleinspect/dto_categories.go \
  internal/modules/articleinspect/service_categories.go \
  internal/modules/articleinspect/repository_categories.go \
  internal/modules/articleinspect/model.go
git commit -m "refactor: remove category code from api contract"
```

### Task 3: Remove category code from migrations and test fixtures

**Files:**
- Modify: `internal/modules/articleinspect/articleinspect_test.go`
- Modify: `migrations/20260420_01_article_inspection.sql`
- Create: `migrations/20260428_01_drop_category_code.sql`

**Step 1: Write the failing test**

Extend the migration-related coverage in `internal/modules/articleinspect/articleinspect_test.go` so it asserts:

- the base migration no longer declares `` `code` VARCHAR(64) NOT NULL ``
- the base migration no longer declares `uk_org_code`
- the follow-up drop migration file exists and includes a guarded drop for both the index and the column
- the SQLite helper schema in `seedOrgCategoryFixtures` creates and inserts categories without `code`

```go
if strings.Contains(text, "`code` VARCHAR(64) NOT NULL") {
	t.Fatalf("base migration still contains category code column")
}
if strings.Contains(text, "uk_org_code") {
	t.Fatalf("base migration still contains uk_org_code")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/modules/articleinspect -run 'TestMigrationFileContainsInspectionTables|TestHandlerOrgCategoryAndArticleCenterContracts'`

Expected: FAIL because the current migration SQL and test fixtures still define `code`.

**Step 3: Write minimal implementation**

Update the base migration and seed rows to remove `code`, then add an idempotent follow-up migration for existing databases.

```sql
ALTER TABLE `xt_article_inspect_categories`
  DROP INDEX `uk_org_code`,
  DROP COLUMN `code`;
```

If MySQL compatibility requires guards, implement them with `information_schema` checks and prepared statements so repeated `migrate` runs stay safe.

Mirror the schema change inside `seedOrgCategoryFixtures` so the SQLite-backed tests use the same category shape as production.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/modules/articleinspect -run 'TestMigrationFileContainsInspectionTables|TestHandlerOrgCategoryAndArticleCenterContracts'`

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/modules/articleinspect/articleinspect_test.go \
  migrations/20260420_01_article_inspection.sql \
  migrations/20260428_01_drop_category_code.sql
git commit -m "refactor: drop category code schema"
```

### Task 4: Final verification and cleanup

**Files:**
- Review: `docs/plans/2026-04-28-category-code-removal-design.md`
- Review: `docs/plans/2026-04-28-category-code-removal.md`
- Verify: touched files from Tasks 1-3

**Step 1: Run focused frontend verification**

Run: `npm test -- src/pages/categories/index.test.tsx src/pages/keywords/index.test.tsx`

Expected: PASS.

**Step 2: Run focused backend verification**

Run: `go test ./internal/modules/articleinspect`

Expected: PASS.

**Step 3: Run admin build verification**

Run: `npm run build`

Expected: PASS with no TypeScript errors in the admin app.

**Step 4: Inspect the final diff**

Run: `git diff --stat HEAD~3..HEAD`

Expected: only category-code-removal changes plus the already-approved plan/design docs.

**Step 5: Commit**

```bash
git add web/admin/src/pages/categories/index.tsx \
  web/admin/src/pages/categories/index.test.tsx \
  web/admin/src/pages/keywords/index.test.tsx \
  web/admin/src/services/categories.ts \
  internal/modules/articleinspect/model.go \
  internal/modules/articleinspect/dto_categories.go \
  internal/modules/articleinspect/service_categories.go \
  internal/modules/articleinspect/repository_categories.go \
  internal/modules/articleinspect/handler.go \
  internal/modules/articleinspect/articleinspect_test.go \
  migrations/20260420_01_article_inspection.sql \
  migrations/20260428_01_drop_category_code.sql
git commit -m "refactor: remove unused category code"
```

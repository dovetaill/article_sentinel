package rules

import (
	"net/http"
	"reflect"
	"sort"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	testutil "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
	"gorm.io/gorm"
)

func newRulesHandler(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	group := huma.NewGroup(api, "/api/v1/article-inspect")
	group.UseSimpleModifier(func(op *huma.Operation) {
		op.SkipValidateParams = true
	})
	RegisterCategoryRoutes(group, NewCategoryService(NewCategoryRepository(db)))
	RegisterKeywordRoutes(group, NewKeywordService(NewKeywordRepository(db)))
	return mux
}

func TestOrgListReturnsSeededOrg(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	handler := newRulesHandler(t, db)

	result := testutil.SendRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/orgs", nil)
	if result.Status != http.StatusOK {
		t.Fatalf("list orgs status = %d, want %d", result.Status, http.StatusOK)
	}

	data := testutil.DataMap(t, result.Envelope.Data)
	items := testutil.ListField(t, data, "items")
	if len(items) == 0 {
		t.Fatal("list orgs items = empty, want seeded org")
	}

	first := testutil.DataMap(t, items[0])
	if testutil.Uint64Field(t, first, "id") != 29 {
		t.Fatalf("first org id = %d, want %d", testutil.Uint64Field(t, first, "id"), 29)
	}
	if testutil.StringField(t, first, "name") != "一县一端" {
		t.Fatalf("first org name = %q, want %q", testutil.StringField(t, first, "name"), "一县一端")
	}
	if testutil.Uint64Field(t, first, "cate_id") != 0 {
		t.Fatalf("first org cate_id = %d, want %d", testutil.Uint64Field(t, first, "cate_id"), 0)
	}
	if _, ok := first["cateid"]; ok {
		t.Fatalf("first org keys = %#v, do not want cateid", first)
	}
}

func TestCategoryListScopedByOrg(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	handler := newRulesHandler(t, db)

	result := testutil.SendRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/categories?orgid=29&page=1&page_size=20", nil)
	if result.Status != http.StatusOK {
		t.Fatalf("list categories status = %d, want %d", result.Status, http.StatusOK)
	}

	data := testutil.DataMap(t, result.Envelope.Data)
	if total := testutil.NumberField(t, data, "total"); total != 2 {
		t.Fatalf("list categories total = %v, want %d", total, 2)
	}

	items := testutil.ListField(t, data, "items")
	gotIDs := make([]uint64, 0, len(items))
	for _, raw := range items {
		item := testutil.DataMap(t, raw)
		if testutil.Uint64Field(t, item, "orgid") != 29 {
			t.Fatalf("category orgid = %d, want %d", testutil.Uint64Field(t, item, "orgid"), 29)
		}
		if _, ok := item["code"]; ok {
			t.Fatalf("category payload = %#v, do not want code", item)
		}
		if _, ok := item["created_at"]; !ok {
			t.Fatalf("category keys = %#v, want created_at", item)
		}
		if _, ok := item["create_at"]; ok {
			t.Fatalf("category keys = %#v, do not want create_at", item)
		}
		gotIDs = append(gotIDs, testutil.Uint64Field(t, item, "id"))
	}
	sort.Slice(gotIDs, func(i, j int) bool { return gotIDs[i] < gotIDs[j] })
	if !reflect.DeepEqual(gotIDs, []uint64{501, 502}) {
		t.Fatalf("category ids = %#v, want %#v", gotIDs, []uint64{501, 502})
	}
}

func TestCategoryRoutesRejectAnonymousMutation(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	handler := newRulesHandler(t, db)

	tests := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{name: "create", method: http.MethodPost, path: "/api/v1/article-inspect/categories", body: map[string]any{"name": "新增分类", "enabled": true}},
		{name: "detail", method: http.MethodGet, path: "/api/v1/article-inspect/categories/501"},
		{name: "update", method: http.MethodPut, path: "/api/v1/article-inspect/categories/501", body: map[string]any{"name": "分类更新", "enabled": true}},
		{name: "delete", method: http.MethodDelete, path: "/api/v1/article-inspect/categories/501"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result testutil.HTTPResult
			if tt.body != nil {
				result = testutil.SendJSONRequestWithOptions(t, handler, tt.method, tt.path, tt.body, testutil.RequestOptions{Anonymous: true})
			} else {
				result = testutil.SendRequestWithOptions(t, handler, tt.method, tt.path, nil, testutil.RequestOptions{Anonymous: true})
			}
			if result.Status != http.StatusUnauthorized {
				t.Fatalf("%s status = %d, want %d", tt.name, result.Status, http.StatusUnauthorized)
			}
			if result.Envelope.Code != http.StatusUnauthorized {
				t.Fatalf("%s envelope = %+v, want unauthorized code", tt.name, result.Envelope)
			}
		})
	}
}

func TestCategoryRoutesUseSnakeCasePayloadWithoutCode(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedOrgCategoryFixtures(t, db)
	handler := newRulesHandler(t, db)

	created := testutil.SendJSONRequest(t, handler, http.MethodPost, "/api/v1/article-inspect/categories", map[string]any{
		"orgid":   29,
		"name":    "新增分类",
		"enabled": true,
		"sort":    10,
	})
	if created.Status != http.StatusCreated {
		t.Fatalf("create category status = %d, want %d", created.Status, http.StatusCreated)
	}

	createdData := testutil.DataMap(t, created.Envelope.Data)
	if testutil.StringField(t, createdData, "name") != "新增分类" {
		t.Fatalf("create category name = %q, want %q", testutil.StringField(t, createdData, "name"), "新增分类")
	}
	if _, ok := createdData["code"]; ok {
		t.Fatalf("category payload = %#v, do not want code", createdData)
	}

	categoryID := testutil.Uint64String(t, createdData["id"])
	updated := testutil.SendJSONRequest(t, handler, http.MethodPut, "/api/v1/article-inspect/categories/"+categoryID, map[string]any{
		"orgid":   29,
		"name":    "新增分类-更新",
		"enabled": false,
		"sort":    15,
	})
	if updated.Status != http.StatusOK {
		t.Fatalf("update category status = %d, want %d", updated.Status, http.StatusOK)
	}

	updatedData := testutil.DataMap(t, updated.Envelope.Data)
	if testutil.StringField(t, updatedData, "name") != "新增分类-更新" {
		t.Fatalf("update category name = %q, want %q", testutil.StringField(t, updatedData, "name"), "新增分类-更新")
	}
	if _, ok := updatedData["code"]; ok {
		t.Fatalf("category payload = %#v, do not want code", updatedData)
	}
}

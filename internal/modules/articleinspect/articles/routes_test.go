package articles

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

func newArticlesHandler(t *testing.T, db *gorm.DB) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	group := huma.NewGroup(api, "/api/v1/article-inspect")
	group.UseSimpleModifier(func(op *huma.Operation) {
		op.SkipValidateParams = true
	})
	RegisterArticleRoutes(group, NewArticleService(NewArticleRepository(db)))
	return mux
}

func TestArticleRoutesListAndDetailContracts(t *testing.T) {
	db := testutil.NewArticleInspectTestDB(t)
	testutil.SeedArticleCenterFixtures(t, db)
	handler := newArticlesHandler(t, db)

	t.Run("article list endpoint reads real articles", func(t *testing.T) {
		result := testutil.SendRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/articles?orgid=29&page=1&page_size=20", nil)
		if result.Status != http.StatusOK {
			t.Fatalf("list articles status = %d, want %d", result.Status, http.StatusOK)
		}

		data := testutil.DataMap(t, result.Envelope.Data)
		if total := testutil.NumberField(t, data, "total"); total != 2 {
			t.Fatalf("list articles total = %v, want %d", total, 2)
		}

		items := testutil.ListField(t, data, "items")
		gotIDs := make([]uint64, 0, len(items))
		for _, raw := range items {
			item := testutil.DataMap(t, raw)
			gotIDs = append(gotIDs, testutil.Uint64Field(t, item, "id"))
		}
		sort.Slice(gotIDs, func(i, j int) bool { return gotIDs[i] < gotIDs[j] })
		if !reflect.DeepEqual(gotIDs, []uint64{9001, 9002}) {
			t.Fatalf("article ids = %#v, want %#v", gotIDs, []uint64{9001, 9002})
		}
	})

	t.Run("article list endpoint filters by title like", func(t *testing.T) {
		result := testutil.SendRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/articles?orgid=29&page=1&page_size=20&title=%E8%A6%81%E9%97%BB%E4%B8%80", nil)
		if result.Status != http.StatusOK {
			t.Fatalf("list articles by title status = %d, want %d", result.Status, http.StatusOK)
		}

		data := testutil.DataMap(t, result.Envelope.Data)
		items := testutil.ListField(t, data, "items")
		if len(items) != 1 {
			t.Fatalf("title filtered items len = %d, want %d", len(items), 1)
		}
		item := testutil.DataMap(t, items[0])
		if testutil.Uint64Field(t, item, "id") != 9001 {
			t.Fatalf("title filtered id = %d, want %d", testutil.Uint64Field(t, item, "id"), 9001)
		}
	})

	t.Run("article list endpoint filters by article id", func(t *testing.T) {
		result := testutil.SendRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/articles?orgid=29&page=1&page_size=20&article_id=9002", nil)
		if result.Status != http.StatusOK {
			t.Fatalf("list articles by article id status = %d, want %d", result.Status, http.StatusOK)
		}

		data := testutil.DataMap(t, result.Envelope.Data)
		items := testutil.ListField(t, data, "items")
		if len(items) != 1 {
			t.Fatalf("article id filtered items len = %d, want %d", len(items), 1)
		}
		item := testutil.DataMap(t, items[0])
		if testutil.Uint64Field(t, item, "id") != 9002 {
			t.Fatalf("article id filtered id = %d, want %d", testutil.Uint64Field(t, item, "id"), 9002)
		}
	})

	t.Run("article list endpoint still supports exact state filtering", func(t *testing.T) {
		result := testutil.SendRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/articles?orgid=29&page=1&page_size=20&state=9", nil)
		if result.Status != http.StatusOK {
			t.Fatalf("list articles by state status = %d, want %d", result.Status, http.StatusOK)
		}

		data := testutil.DataMap(t, result.Envelope.Data)
		items := testutil.ListField(t, data, "items")
		if len(items) != 1 {
			t.Fatalf("state filtered items len = %d, want %d", len(items), 1)
		}
		item := testutil.DataMap(t, items[0])
		if testutil.Uint64Field(t, item, "id") != 9001 {
			t.Fatalf("state filtered id = %d, want %d", testutil.Uint64Field(t, item, "id"), 9001)
		}
	})

	t.Run("article detail endpoint includes article data and latest inspect summary", func(t *testing.T) {
		result := testutil.SendRequest(t, handler, http.MethodGet, "/api/v1/article-inspect/articles/9001?orgid=29", nil)
		if result.Status != http.StatusOK {
			t.Fatalf("get article detail status = %d, want %d", result.Status, http.StatusOK)
		}

		data := testutil.DataMap(t, result.Envelope.Data)
		if testutil.Uint64Field(t, data, "id") != 9001 {
			t.Fatalf("article detail id = %d, want %d", testutil.Uint64Field(t, data, "id"), 9001)
		}
		if testutil.StringField(t, data, "title") != "县域要闻一" {
			t.Fatalf("article detail title = %q, want %q", testutil.StringField(t, data, "title"), "县域要闻一")
		}
		if testutil.StringField(t, data, "body") != "<p>real body one</p>" {
			t.Fatalf("article detail body = %q, want %q", testutil.StringField(t, data, "body"), "<p>real body one</p>")
		}
		if testutil.StringField(t, data, "thumbnail") != "https://example.com/article-9001.png" {
			t.Fatalf("article detail thumbnail = %q, want %q", testutil.StringField(t, data, "thumbnail"), "https://example.com/article-9001.png")
		}
		if testutil.Uint64Field(t, data, "latest_task_id") != 702 {
			t.Fatalf("article detail latest_task_id = %d, want %d", testutil.Uint64Field(t, data, "latest_task_id"), 702)
		}
	})
}

package articleinspect

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/dovetaill/article-sentinel/internal/api/response"
	"github.com/dovetaill/article-sentinel/internal/identity"
	"github.com/dovetaill/article-sentinel/internal/modules/articleinspect/testutil"
	queuetasks "github.com/dovetaill/article-sentinel/internal/queue/tasks"
	"gorm.io/gorm"
)

func newArticleInspectTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testutil.NewArticleInspectTestDB(t)
}

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	return testutil.MustTime(t, value)
}

func seedCandidateArticles(t *testing.T, db *gorm.DB) {
	t.Helper()
	testutil.SeedCandidateArticles(t, db)
}

func seedLifecycleArticles(t *testing.T, db *gorm.DB) {
	t.Helper()
	testutil.SeedLifecycleArticles(t, db)
}

func seedOrgCategoryFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()
	testutil.SeedOrgCategoryFixtures(t, db)
}

func seedArticleCenterFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()
	testutil.SeedArticleCenterFixtures(t, db)
}

func seedQueryFixtures(t *testing.T, db *gorm.DB) {
	t.Helper()
	testutil.SeedQueryFixtures(t, db)
}

type articleInspectHTTPResult struct {
	status   int
	envelope response.Envelope
}

type articleInspectRequestOptions struct {
	Anonymous bool
	Session   *identity.AdminSession
}

type articleInspectTaskDispatcherStub struct {
	payloads []queuetasks.ArticleInspectTaskPayload
	err      error
}

func (s *articleInspectTaskDispatcherStub) DispatchArticleInspectTask(ctx context.Context, payload queuetasks.ArticleInspectTaskPayload) error {
	_ = ctx
	s.payloads = append(s.payloads, payload)
	return s.err
}

func newArticleInspectHandler(t *testing.T, db *gorm.DB, dispatcher TaskDispatcher) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("Test API", "1.0.0"))
	RegisterRoutes(api, NewRoutes(db, dispatcher))
	return mux
}

func sendArticleInspectJSONRequest(t *testing.T, handler http.Handler, method, path string, body any) articleInspectHTTPResult {
	t.Helper()
	result := testutil.SendJSONRequest(t, handler, method, path, body)
	return articleInspectHTTPResult{status: result.Status, envelope: result.Envelope}
}

func sendArticleInspectJSONRequestWithOptions(t *testing.T, handler http.Handler, method, path string, body any, options articleInspectRequestOptions) articleInspectHTTPResult {
	t.Helper()
	result := testutil.SendJSONRequestWithOptions(t, handler, method, path, body, testutil.RequestOptions{Anonymous: options.Anonymous, Session: options.Session})
	return articleInspectHTTPResult{status: result.Status, envelope: result.Envelope}
}

func sendArticleInspectRequest(t *testing.T, handler http.Handler, method, path string, body io.Reader) articleInspectHTTPResult {
	t.Helper()
	result := testutil.SendRequest(t, handler, method, path, body)
	return articleInspectHTTPResult{status: result.Status, envelope: result.Envelope}
}

func sendArticleInspectRequestWithOptions(t *testing.T, handler http.Handler, method, path string, body io.Reader, options articleInspectRequestOptions) articleInspectHTTPResult {
	t.Helper()
	result := testutil.SendRequestWithOptions(t, handler, method, path, body, testutil.RequestOptions{Anonymous: options.Anonymous, Session: options.Session})
	return articleInspectHTTPResult{status: result.Status, envelope: result.Envelope}
}

func articleInspectDataMap(t *testing.T, value any) map[string]any {
	t.Helper()
	return testutil.DataMap(t, value)
}

func articleInspectListField(t *testing.T, m map[string]any, key string) []any {
	t.Helper()
	return testutil.ListField(t, m, key)
}

func articleInspectNumberField(t *testing.T, m map[string]any, key string) float64 {
	t.Helper()
	return testutil.NumberField(t, m, key)
}

func articleInspectStringField(t *testing.T, m map[string]any, key string) string {
	t.Helper()
	return testutil.StringField(t, m, key)
}

func articleInspectUint64Field(t *testing.T, m map[string]any, key string) uint64 {
	t.Helper()
	return testutil.Uint64Field(t, m, key)
}

func articleInspectUint64String(t *testing.T, value any) string {
	t.Helper()
	return testutil.Uint64String(t, value)
}

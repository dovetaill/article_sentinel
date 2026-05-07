package middleware

import (
	"net/http"
	"strings"

	"github.com/dovetaill/go-auth-demo/internal/identity"
)

type DocumentationPaths struct {
	OpenAPIPath string
	DocsPath    string
	SchemasPath string
}

type DocumentationOption func(*documentationProtectionConfig)

func WithDocumentationPaths(paths DocumentationPaths) DocumentationOption {
	return func(cfg *documentationProtectionConfig) {
		cfg.apply(paths)
	}
}

func ProtectDocumentation(enabled bool, options ...DocumentationOption) Middleware {
	cfg := newDocumentationProtectionConfig()
	for _, option := range options {
		if option != nil {
			option(&cfg)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if enabled && cfg.contains(r.URL.Path) {
				if _, ok := identity.ActorFromContext(r.Context()); !ok {
					writeAuthError(w, http.StatusUnauthorized, "unauthorized")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

type documentationProtectionConfig struct {
	openAPIBase string
	docsPath    string
	schemasPath string
}

func newDocumentationProtectionConfig() documentationProtectionConfig {
	cfg := documentationProtectionConfig{}
	cfg.apply(DocumentationPaths{
		OpenAPIPath: "/openapi",
		DocsPath:    "/docs",
		SchemasPath: "/schemas",
	})
	return cfg
}

func (c *documentationProtectionConfig) apply(paths DocumentationPaths) {
	if c == nil {
		return
	}
	if openAPIBase := NormalizeOpenAPIPath(paths.OpenAPIPath); openAPIBase != "" {
		c.openAPIBase = openAPIBase
	}
	if docsPath := normalizeDocumentationPath(paths.DocsPath); docsPath != "" {
		c.docsPath = docsPath
	}
	if schemasPath := normalizeDocumentationPath(paths.SchemasPath); schemasPath != "" {
		c.schemasPath = schemasPath
	}
}

func (c documentationProtectionConfig) contains(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	if path == c.docsPath || strings.HasPrefix(path, c.docsPath+"/") {
		return true
	}
	if c.schemasPath != "" && strings.HasPrefix(path, c.schemasPath+"/") {
		return true
	}
	if c.openAPIBase == "" {
		return false
	}
	for _, variant := range []string{
		c.openAPIBase + ".json",
		c.openAPIBase + ".yaml",
		c.openAPIBase + "-3.0.json",
		c.openAPIBase + "-3.0.yaml",
	} {
		if path == variant {
			return true
		}
	}
	return false
}

// NormalizeOpenAPIPath converts a configured OpenAPI path to the base path
// Huma should use before appending .json/.yaml variants.
func NormalizeOpenAPIPath(path string) string {
	path = normalizeDocumentationPath(path)
	if path == "" {
		return "/openapi"
	}
	for _, suffix := range []string{".json", ".yaml"} {
		if strings.HasSuffix(path, suffix) {
			return strings.TrimSuffix(path, suffix)
		}
	}
	return path
}

func normalizeDocumentationPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if path != "/" {
		path = strings.TrimRight(path, "/")
	}
	return path
}

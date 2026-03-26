package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSPAFrontendServesAssetFile(t *testing.T) {
	root := writeFrontendFixture(t)
	router := gin.New()
	registerSPAFrontend(router, root)

	request := httptest.NewRequest(http.MethodGet, "/style.css", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 for asset, got %d", response.Code)
	}

	if got := response.Body.String(); got != "body { color: red; }" {
		t.Fatalf("unexpected asset body: %q", got)
	}
}

func TestSPAFrontendFallsBackToIndexForHTMLRequests(t *testing.T) {
	root := writeFrontendFixture(t)
	router := gin.New()
	registerSPAFrontend(router, root)

	request := httptest.NewRequest(http.MethodGet, "/overview", nil)
	request.Header.Set("Accept", "text/html,application/xhtml+xml")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 for html fallback, got %d", response.Code)
	}

	if got := response.Body.String(); got != "<!doctype html><title>app</title>" {
		t.Fatalf("unexpected index body: %q", got)
	}
}

func TestSPAFrontendDoesNotSwallowUnknownAPIRequests(t *testing.T) {
	root := writeFrontendFixture(t)
	router := gin.New()
	registerSPAFrontend(router, root)

	request := httptest.NewRequest(http.MethodGet, "/api/missing", nil)
	request.Header.Set("Accept", "text/html")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown api route, got %d", response.Code)
	}
}

func TestSPAFrontendReturns404ForUnknownNonHTMLRequests(t *testing.T) {
	root := writeFrontendFixture(t)
	router := gin.New()
	registerSPAFrontend(router, root)

	request := httptest.NewRequest(http.MethodGet, "/missing.json", nil)
	request.Header.Set("Accept", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown non-html request, got %d", response.Code)
	}
}

func writeFrontendFixture(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<!doctype html><title>app</title>"), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "style.css"), []byte("body { color: red; }"), 0o644); err != nil {
		t.Fatalf("write style.css: %v", err)
	}

	return root
}

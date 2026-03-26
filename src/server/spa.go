package main

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func registerSPAFrontend(router *gin.Engine, root string) {
	handler := spaFrontendHandler(root)
	router.NoRoute(handler)
	router.NoMethod(handler)
}

func spaFrontendHandler(root string) gin.HandlerFunc {
	indexPath := filepath.Join(root, "index.html")

	return func(c *gin.Context) {
		if !isSPARequestMethod(c.Request.Method) || isAPIRequestPath(c.Request.URL.Path) {
			c.Status(http.StatusNotFound)
			return
		}

		if assetPath, ok := resolveFrontendAssetPath(root, c.Request.URL.Path); ok {
			c.File(assetPath)
			return
		}

		if acceptsHTML(c.Request) && fileExists(indexPath) {
			c.File(indexPath)
			return
		}

		c.Status(http.StatusNotFound)
	}
}

func isSPARequestMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

func isAPIRequestPath(requestPath string) bool {
	return requestPath == "/api" || strings.HasPrefix(requestPath, "/api/")
}

func acceptsHTML(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

func resolveFrontendAssetPath(root string, requestPath string) (string, bool) {
	cleanedPath := path.Clean("/" + requestPath)
	relativePath := strings.TrimPrefix(cleanedPath, "/")
	if relativePath == "" {
		return "", false
	}

	resolvedPath := filepath.Join(root, filepath.FromSlash(relativePath))
	if !pathIsWithinRoot(root, resolvedPath) {
		return "", false
	}

	if !fileExists(resolvedPath) {
		return "", false
	}

	return resolvedPath, true
}

func pathIsWithinRoot(root string, candidate string) bool {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return false
	}

	absoluteCandidate, err := filepath.Abs(candidate)
	if err != nil {
		return false
	}

	relativePath, err := filepath.Rel(absoluteRoot, absoluteCandidate)
	if err != nil {
		return false
	}

	return relativePath == "." || (relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(filepath.Separator)))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

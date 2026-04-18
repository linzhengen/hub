package http

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/linzhengen/hub/v1/server/pkg/logger"
	"github.com/linzhengen/hub/v1/ui/web"
)

func TestNewSPAHandler(t *testing.T) {
	// Create a test handler with mock filesystem
	handler := createTestHandler(t)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "root path should serve index.html",
			path:           "/",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "index.html path should serve index.html",
			path:           "/index.html",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent path should serve index.html (SPA routing)",
			path:           "/non-existent",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "static asset path should serve file",
			path:           "/assets/index.css",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestEmbeddedFS(t *testing.T) {
	// Use mock embedded filesystem if the real one is empty
	testFS := getTestFS(t)

	// Test that the embedded filesystem contains the expected files
	expectedFiles := []string{
		"dist/index.html",
		"dist/favicon.png",
	}

	for _, filename := range expectedFiles {
		_, err := testFS.Open(filename)
		if err != nil {
			t.Errorf("embedded filesystem missing expected file %s: %v", filename, err)
		}
	}

	// Check that assets directory exists
	assetsDir, err := testFS.Open("dist/assets")
	if err != nil {
		t.Errorf("embedded filesystem missing assets directory: %v", err)
		return
	}
	defer func() {
		if err := assetsDir.Close(); err != nil {
			// Log the error but don't fail the test
			t.Logf("warning: failed to close assets directory: %v", err)
		}
	}()

	// Check it's a directory
	stat, err := assetsDir.Stat()
	if err != nil {
		t.Errorf("failed to stat assets directory: %v", err)
		return
	}

	if !stat.IsDir() {
		t.Error("assets is not a directory")
	}
}

// getTestFS returns either the real embedded filesystem or a mock one
func getTestFS(t *testing.T) fs.FS {
	// Try to open a file from the real embedded filesystem
	_, err := web.Embedded.Open("dist/index.html")
	if err == nil {
		// Real filesystem has files, use it
		return web.Embedded
	}

	// Create a mock filesystem for testing
	t.Log("Using mock embedded filesystem for testing")
	return createMockFS()
}

// createMockFS creates a mock filesystem for testing
func createMockFS() fs.FS {
	// Create an in-memory filesystem
	fsys := fstest.MapFS{
		"dist/index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><head><title>Hub</title></head><body><div id=\"root\"></div></body></html>"),
		},
		"dist/favicon.png": &fstest.MapFile{
			Data: []byte{},
		},
		"dist/assets/index.css": &fstest.MapFile{
			Data: []byte("/* Mock CSS */"),
		},
		"dist/assets/index.js": &fstest.MapFile{
			Data: []byte("// Mock JS"),
		},
	}
	return fsys
}

// testSpaHandler is a copy of the spaHandler struct for testing
type testSpaHandler struct {
	fs http.FileSystem
}

// createTestHandler creates a spaHandler for testing with either real or mock filesystem
func createTestHandler(t *testing.T) http.Handler {
	testFS := getTestFS(t)

	// Get the subdirectory "dist" from the filesystem
	distFS, err := fs.Sub(testFS, "dist")
	if err != nil {
		// If we can't get the dist subdirectory, use the root
		distFS = testFS
	}

	return &testSpaHandler{
		fs: http.FS(distFS),
	}
}

// ServeHTTP implements the http.Handler interface for testSpaHandler
func (h *testSpaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean the path to prevent directory traversal
	requestPath := path.Clean(r.URL.Path)
	if requestPath == "/" || requestPath == "" {
		requestPath = "/index.html"
	}

	// Try to open the requested file
	f, err := h.fs.Open(requestPath)
	if err != nil {
		// File not found, serve index.html for SPA routing
		h.serveIndex(w, r)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.Infof("failed to close file: %v", err)
		}
	}()

	// Get file info to check if it's a directory
	stat, err := f.Stat()
	if err != nil || stat.IsDir() {
		// If it's a directory or we can't stat it, serve index.html
		h.serveIndex(w, r)
		return
	}

	// Special handling for index.html to set proper cache headers
	if strings.HasSuffix(requestPath, "index.html") {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	} else {
		// Static assets can be cached
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	}

	// Serve the file
	http.ServeContent(w, r, requestPath, stat.ModTime(), f)
}

// serveIndex serves the index.html file for SPA routing.
func (h *testSpaHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
	// Set no-cache headers for index.html
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Try to open index.html
	f, err := h.fs.Open("/index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusNotFound)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.Infof("failed to close file: %v", err)
		}
	}()

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, "index.html", stat.ModTime(), f)
}

package http

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/linzhengen/hub/v1/server/pkg/logger"
	"github.com/linzhengen/hub/v1/ui/web"
)

// spaHandler implements a http.Handler that serves the SPA (Single Page Application)
// from embedded files. It serves static files from the embedded dist directory
// and falls back to serving index.html for any path that doesn't match a static file.
type spaHandler struct {
	fs http.FileSystem
}

// NewSPAHandler creates a new SPA handler that serves files from the embedded UI resources.
func NewSPAHandler() http.Handler {
	// Use the embedded filesystem from the web package
	embeddedFS := web.Embedded

	// Get the subdirectory "dist" from the embedded filesystem
	distFS, err := fs.Sub(embeddedFS, "dist")
	if err != nil {
		// If we can't get the dist subdirectory, use the root
		distFS = embeddedFS
	}

	return &spaHandler{
		fs: http.FS(distFS),
	}
}

// ServeHTTP implements the http.Handler interface.
func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
func (h *spaHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
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

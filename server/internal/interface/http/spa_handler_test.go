package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linzhengen/hub/v1/ui/web"
)

func TestNewSPAHandler(t *testing.T) {
	handler := NewSPAHandler()

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
	// Test that the embedded filesystem contains the expected files
	expectedFiles := []string{
		"dist/index.html",
		"dist/favicon.png",
	}

	for _, filename := range expectedFiles {
		_, err := web.Embedded.Open(filename)
		if err != nil {
			t.Errorf("embedded filesystem missing expected file %s: %v", filename, err)
		}
	}

	// Check that assets directory exists
	assetsDir, err := web.Embedded.Open("dist/assets")
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

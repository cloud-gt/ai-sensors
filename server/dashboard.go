package server

import (
	"io/fs"
	"net/http"
	"strings"
)

// NewDashboardHandler creates an HTTP handler serving the embedded SPA.
// It serves static files directly and falls back to index.html for
// client-side routing.
func NewDashboardHandler(distFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		if path != "" {
			if f, err := distFS.Open(path); err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// SPA fallback: serve index.html for client-side routing
		r.URL.Path = "/index.html"
		fileServer.ServeHTTP(w, r)
	})
}

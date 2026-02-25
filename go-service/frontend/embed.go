package frontend

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var distFS embed.FS

// DistFS returns the embedded frontend filesystem rooted at "dist".
func DistFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}

// Handler returns an http.Handler that serves the embedded SPA.
// - Requests matching a real file are served directly (JS, CSS, images, etc.)
// - All other requests fall back to index.html (SPA client-side routing)
func Handler() http.Handler {
	sub, err := DistFS()
	if err != nil {
		panic("frontend: failed to create sub filesystem: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to open the requested file
		path := r.URL.Path
		if path == "/" {
			path = "index.html"
		} else if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}

		// Check if the file exists in the embedded FS
		if f, err := sub.Open(path); err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for SPA routing
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

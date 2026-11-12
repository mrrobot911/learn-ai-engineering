package handler

import (
	"net/http"
	"path"
	"strings"
)

// NewRouter serves the API under /api and static files from staticDir.
// An empty staticDir returns api unchanged so deployments without a frontend build keep working.
func NewRouter(api http.Handler, staticDir string) http.Handler {
	if staticDir == "" {
		return api
	}

	dir := http.Dir(staticDir)
	files := http.FileServer(dir)
	serveIndex := func(w http.ResponseWriter, r *http.Request) {
		file, err := dir.Open("/index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		http.ServeContent(w, r, "index.html", info.ModTime(), file)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/health":
			api.ServeHTTP(w, r)
		case r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/"):
			next := r.Clone(r.Context())
			next.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
			next.URL.RawPath = ""
			api.ServeHTTP(w, next)
			r.Pattern = next.Pattern
		default:
			if file, err := dir.Open(path.Clean("/" + r.URL.Path)); err == nil {
				file.Close()
				files.ServeHTTP(w, r)
				return
			}
			serveIndex(w, r)
		}
	})
}

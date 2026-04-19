package http

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// StaticHandler раздает файлы из embed.FS.
// Для SPA: если файл не найден — отдает index.html.
func StaticHandler(dist embed.FS, prefix string) http.Handler {
	sub, err := fs.Sub(dist, prefix)
	if err != nil {
		panic("invalid embed prefix: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws/") || r.URL.Path == "/healthz" {
			http.NotFound(w, r)
			return
		}

		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		if _, err := fs.Stat(sub, strings.TrimPrefix(path, "/")); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

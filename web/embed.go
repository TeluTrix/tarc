package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var files embed.FS

func Handler() http.Handler {
	staticFiles, err := fs.Sub(files, "dist")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(staticFiles))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(staticFiles, path); err != nil {
			r.URL.Path = "/index.html"
		}

		fileServer.ServeHTTP(w, r)
	})
}
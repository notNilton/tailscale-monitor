package server

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web
var webFiles embed.FS

// ServeStatic serves static files from embedded filesystem
func ServeStatic() http.Handler {
	// Get the web subdirectory from embedded filesystem
	webFS, err := fs.Sub(webFiles, "web")
	if err != nil {
		panic(err)
	}

	return http.FileServer(http.FS(webFS))
}

package ui

import (
	"embed"
	"net/http"
)

//go:embed index.html
var uiFS embed.FS

// Handler returns an HTTP handler serving the embedded UI
func Handler() http.Handler {
	return http.FileServer(http.FS(uiFS))
}

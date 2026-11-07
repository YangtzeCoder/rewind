package server

import (
	"bytes"
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed web
var frontendFiles embed.FS

func frontendFallback(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/api") || (c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead) {
		failure(c, errNotFound, "not found")
		return
	}
	dist, err := fs.Sub(frontendFiles, "web/dist")
	if err != nil {
		failure(c, errNotFound, "not found")
		return
	}
	name := strings.TrimPrefix(c.Request.URL.Path, "/")
	if name == "" {
		name = "index.html"
	}
	info, err := fs.Stat(dist, name)
	if err != nil || info.IsDir() {
		// Missing assets should return 404; client-side routes use the SPA entry.
		if path.Ext(name) != "" {
			http.NotFound(c.Writer, c.Request)
			return
		}
		name = "index.html"
	}
	content, err := fs.ReadFile(dist, name)
	if err != nil {
		failure(c, errNotFound, "not found")
		return
	}
	http.ServeContent(c.Writer, c.Request, name, infoModTime(dist, name), bytes.NewReader(content))
}

func infoModTime(files fs.FS, name string) (modTime time.Time) {
	if info, err := fs.Stat(files, name); err == nil {
		return info.ModTime()
	}
	return modTime
}

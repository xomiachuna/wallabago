package docs

import (
	"embed"
	"io/fs"
	"net/http"
)

var OpenAPI http.Handler

//go:embed static/*
var staticFiles embed.FS

func init() {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	OpenAPI = http.FileServerFS(sub)
}

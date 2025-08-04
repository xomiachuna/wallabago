package docs

import (
	"embed"
	"io/fs"
	"net/http"
)

type OpenAPI struct{}

func NewOpenAPI() *OpenAPI {
	return &OpenAPI{}
}

//go:embed static/*
var staticFiles embed.FS

func (o *OpenAPI) OpenAPIUI(w http.ResponseWriter, r *http.Request) {
	// Serve static/* folder
	//nolint:errcheck //this is guaranteed to be present
	sub, _ := fs.Sub(staticFiles, "static")
	http.FileServerFS(sub).ServeHTTP(w, r)
}

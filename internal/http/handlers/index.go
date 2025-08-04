package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/andriihomiak/wallabago/internal/http/constants"
)

// todo: use template/html
const indexPage = `<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="utf-8">
        <title>Wallabago</title>
    </head>
    <body>
        <h1>Wallabago</h1>
        <p>%s</p>
    </body>
</html>`

type WebUI struct{}

func NewWebUI() *WebUI {
	return &WebUI{}
}

func (s *WebUI) Index(w http.ResponseWriter, r *http.Request) {
	slog.DebugContext(r.Context(), "index", "pattern", r.Pattern)
	w.WriteHeader(http.StatusOK)
	w.Header().Set(constants.HeaderContentType, constants.MimeTextHTML)
	fmt.Fprintf(w, indexPage, time.Now().UTC().Format(time.Layout))
}

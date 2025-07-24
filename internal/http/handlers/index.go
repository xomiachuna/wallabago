package handlers

import (
	"fmt"
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

type Index struct{}

func (s *Index) Index(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set(constants.HeaderContentType, constants.MimeTextHTML)
	fmt.Fprintf(w, indexPage, time.Now().UTC().Format(time.Layout))
}

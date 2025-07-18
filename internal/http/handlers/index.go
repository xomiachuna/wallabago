package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/andriihomiak/wallabago/internal/http/constants"
)

const indexPage = `<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="utf-8">
        <title>Wallabago</title>
    </head>
    <body>
        <h1>Wallabago</h1>
    </body>
</html>`

func Index(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set(constants.HeaderContentType, constants.MimeTextHTML)
	time.Sleep(time.Second * 10)
	fmt.Fprint(w, indexPage)
}

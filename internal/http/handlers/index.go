package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/andriihomiak/wallabago/internal/http/constants"
	"github.com/pkg/errors"
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

func (s *TimeService) Index(w http.ResponseWriter, r *http.Request) {
	timestamp, err := s.querier.CurrentTimestamp(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set(constants.HeaderContentType, constants.MimeTextPlain)
		fmt.Fprintf(w, "Error: %s", errors.WithStack(err))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set(constants.HeaderContentType, constants.MimeTextHTML)
	fmt.Fprintf(w, indexPage, timestamp.UTC().Format(time.Layout))
}

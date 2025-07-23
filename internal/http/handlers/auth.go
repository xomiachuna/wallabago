package handlers

import (
	"fmt"
	"net/http"

	"github.com/andriihomiak/wallabago/internal/http/constants"
)

func TokenEndpoint(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set(constants.HeaderContentType, constants.MimeTextHTML)
	fmt.Fprintf(w, "url: %s", r.URL.String())
}

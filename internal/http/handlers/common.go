package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andriihomiak/wallabago/internal/http/constants"
	"github.com/pkg/errors"
)

func RespondOKJSON(w http.ResponseWriter, r *http.Request, body any) {
	RespondJSON(w, r, body, http.StatusOK)
}

func RespondJSON(w http.ResponseWriter, r *http.Request, body any, status int) {
	bodyContent, err := json.Marshal(body)
	if err != nil {
		RespondInternalErrorWithStack(w, r, err)
		return
	}
	w.WriteHeader(status)
	w.Header().Set(constants.HeaderContentType, constants.MimeApplicationJSON)
	//nolint:errcheck //todo
	w.Write(bodyContent)
}

func RespondInternalErrorWithStack(w http.ResponseWriter, _ *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set(constants.HeaderContentType, constants.MimeTextPlain)
	fmt.Fprint(w, errors.WithStack(err))
}

func RespondErrorPlain(w http.ResponseWriter, _ *http.Request, err error, status int) {
	w.WriteHeader(status)
	w.Header().Set(constants.HeaderContentType, constants.MimeTextPlain)
	fmt.Fprint(w, err)
}

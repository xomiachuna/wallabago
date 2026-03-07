package response

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
	w.Header().Set(constants.HeaderContentType, constants.MimeApplicationJSON)
	w.WriteHeader(status)
	//nolint:errcheck //todo
	w.Write(bodyContent)
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func RespondInternalErrorWithStack(w http.ResponseWriter, _ *http.Request, err error) {
	if errWithStack, ok := err.(stackTracer); ok {
		http.Error(w, fmt.Sprintf("error: %s\nstack:\n%+v", err.Error(), errWithStack.StackTrace()), http.StatusInternalServerError)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RespondErrorPlain(w http.ResponseWriter, _ *http.Request, err error, status int) {
	http.Error(w, err.Error(), status)
}

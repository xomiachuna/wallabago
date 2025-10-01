package handlers

import (
	"net/http"
	neturl "net/url"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/http/constants"
	"github.com/andriihomiak/wallabago/internal/http/middleware"
	"github.com/andriihomiak/wallabago/internal/http/response"
	"github.com/andriihomiak/wallabago/internal/managers"
	"github.com/pkg/errors"
)

type API struct {
	entryManager *managers.EntryManager
}

func NewAPI(
	entryManager *managers.EntryManager,
) *API {
	return &API{
		entryManager: entryManager,
	}
}

const (
	fieldURL = "url"
)

func requireAddEntryForm(r *http.Request) (*core.NewEntry, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}
	url, err := requiredPostFormField(r, fieldURL)
	if err != nil {
		return nil, err
	}
	parsedURL, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	return &core.NewEntry{
		URL: *parsedURL,
	}, nil
}

func (a *API) HandleAddEntry(w http.ResponseWriter, r *http.Request) {
	token := middleware.MustGetAccessToken(r)
	if r.Header.Get(constants.HeaderContentType) != constants.MimeApplicationXWWWFormURLEncoded {
		w.Header().Set(constants.HeaderAccept, constants.MimeApplicationXWWWFormURLEncoded)
		response.RespondErrorPlain(w, r, nil, http.StatusUnsupportedMediaType)
		return
	}
	entry, err := requireAddEntryForm(r)
	if err != nil {
		response.RespondErrorPlain(w, r, err, http.StatusBadRequest)
		return
	}
	result, err := a.entryManager.AddEntry(r.Context(), token, *entry)
	if err != nil {
		// Check if this is a retrieval error (user error) vs internal error
		var retrievalErr *core.RetrievalError
		if errors.As(err, &retrievalErr) {
			response.RespondErrorPlain(w, r, err, http.StatusBadRequest)
			return
		}
		response.RespondInternalErrorWithStack(w, r, err)
		return
	}
	response.RespondOKJSON(w, r, result)
}

func (a *API) HandleEntryExists(w http.ResponseWriter, r *http.Request) {
	token := middleware.MustGetAccessToken(r)
	url, err := requiredQueryField(r, fieldURL)
	if err != nil {
		response.RespondErrorPlain(w, r, err, http.StatusBadRequest)
		return
	}
	parsedURL, err := neturl.Parse(url)
	if err != nil {
		response.RespondErrorPlain(w, r, err, http.StatusBadRequest)
		return
	}
	result, err := a.entryManager.EntryExists(r.Context(), token, *parsedURL)
	if err != nil {
		response.RespondInternalErrorWithStack(w, r, err)
		return
	}

	response.RespondOKJSON(w, r, core.EntryExistence{
		Exists: result,
	})
}

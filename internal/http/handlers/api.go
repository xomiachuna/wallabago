package handlers

import (
	"fmt"
	"net/http"
	neturl "net/url"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/http/constants"
	"github.com/andriihomiak/wallabago/internal/http/middleware"
	"github.com/andriihomiak/wallabago/internal/http/response"
	"github.com/andriihomiak/wallabago/internal/managers"
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
	formFieldURL = "url"
)

func requireAddEntryForm(r *http.Request) (*core.NewEntry, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}
	url, err := requiredPostFormField(r, formFieldURL)
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

func (a *API) AddEntry(w http.ResponseWriter, r *http.Request) {
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
		response.RespondInternalErrorWithStack(w, r, err)
		return
	}
	response.RespondOKJSON(w, r, result)
}

func (a *API) EntryExists(w http.ResponseWriter, r *http.Request) {
	_ = middleware.MustGetAccessToken(r)
	response.RespondInternalErrorWithStack(w, r, fmt.Errorf("not implemented"))
}

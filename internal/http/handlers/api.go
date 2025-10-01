package handlers

import (
	"net/http"
	neturl "net/url"
	"strconv"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/http/constants"
	"github.com/andriihomiak/wallabago/internal/http/middleware"
	"github.com/andriihomiak/wallabago/internal/http/response"
	"github.com/andriihomiak/wallabago/internal/managers"
	"github.com/pkg/errors"
)

type API struct {
	entryManager    *managers.EntryManager
	identityManager *managers.IdentityManager
}

func NewAPI(
	entryManager *managers.EntryManager,
	identityManager *managers.IdentityManager,
) *API {
	return &API{
		entryManager:    entryManager,
		identityManager: identityManager,
	}
}

const (
	fieldURL      = "url"
	fieldUsername = "username"
	fieldIsAdmin  = "is_admin"
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

func (a *API) HandleGetEntry(w http.ResponseWriter, r *http.Request) {
	token := middleware.MustGetAccessToken(r)
	entryIDStr, err := requiredPathParam(r, "id")
	if err != nil {
		response.RespondErrorPlain(w, r, err, http.StatusBadRequest)
		return
	}
	entryID, err := strconv.ParseInt(entryIDStr, 10, 32)
	if err != nil {
		response.RespondErrorPlain(w, r, errors.Wrap(err, "invalid entry ID"), http.StatusBadRequest)
		return
	}
	entry, err := a.entryManager.GetEntry(r.Context(), token, int32(entryID))
	if err != nil {
		response.RespondInternalErrorWithStack(w, r, err)
		return
	}
	response.RespondOKJSON(w, r, entry)
}

func (a *API) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(constants.HeaderContentType) != constants.MimeApplicationXWWWFormURLEncoded {
		w.Header().Set(constants.HeaderAccept, constants.MimeApplicationXWWWFormURLEncoded)
		response.RespondErrorPlain(w, r, nil, http.StatusUnsupportedMediaType)
		return
	}

	username, err := requiredPostFormField(r, fieldUsername)
	if err != nil {
		response.RespondErrorPlain(w, r, err, http.StatusBadRequest)
		return
	}

	// Check if is_admin field is provided (optional, defaults to false)
	isAdmin := false
	if r.PostForm.Has(fieldIsAdmin) {
		isAdminStr := r.PostForm.Get(fieldIsAdmin)
		isAdmin, err = strconv.ParseBool(isAdminStr)
		if err != nil {
			response.RespondErrorPlain(w, r, errors.Wrap(err, "invalid is_admin value"), http.StatusBadRequest)
			return
		}
	}

	createdUser, err := a.identityManager.CreateUser(r.Context(), username, isAdmin)
	if err != nil {
		response.RespondInternalErrorWithStack(w, r, err)
		return
	}

	response.RespondOKJSON(w, r, createdUser)
}

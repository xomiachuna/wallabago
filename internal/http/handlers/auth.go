package handlers

import (
	"fmt"
	"net/http"

	"github.com/andriihomiak/wallabago/internal/auth"
	"github.com/andriihomiak/wallabago/internal/http/constants"
	"github.com/andriihomiak/wallabago/internal/managers"
	"github.com/pkg/errors"
)

type OAuth2Handler struct {
	manager managers.OAuth2Manager
}

const (
	OAuth2GrantType    = "grant_type"
	OAuth2ClientID     = "client_id"
	OAuth2ClientSecret = "client_secret"
	OAuth2Username     = "username"
	OAuth2Password     = "password"
	OAuth2Scope        = "scope"
	OAuth2RefreshToken = "refresh_token"
)

func requiredPostFormField(r *http.Request, key string) (string, error) {
	err := r.ParseForm()
	if err != nil {
		return "", errors.WithStack(err)
	}

	if !r.PostForm.Has(key) {
		return "", fmt.Errorf("required field: %s", key)
	}
	return r.PostForm.Get(key), nil
}

func (h *OAuth2Handler) handlePasswordFlow(w http.ResponseWriter, r *http.Request) {
	// todo: parse params into a struct
	clientID, requiredErr := requiredPostFormField(r, OAuth2ClientID)
	if requiredErr != nil {
		RespondErrorPlain(w, r, requiredErr, http.StatusBadRequest)
		return
	}
	clientSecret, requiredErr := requiredPostFormField(r, OAuth2ClientSecret)
	if requiredErr != nil {
		RespondErrorPlain(w, r, requiredErr, http.StatusBadRequest)
		return
	}
	username, requiredErr := requiredPostFormField(r, OAuth2Username)
	if requiredErr != nil {
		RespondErrorPlain(w, r, requiredErr, http.StatusBadRequest)
		return
	}

	password, requiredErr := requiredPostFormField(r, OAuth2Password)
	if requiredErr != nil {
		RespondErrorPlain(w, r, requiredErr, http.StatusBadRequest)
		return
	}

	token, err := h.manager.PasswordFlow(r.Context(), clientID, clientSecret, username, password)
	if err != nil {
		var authError auth.AuthError
		if errors.As(err, authError) {
			RespondJSON(w, r, authError, http.StatusUnauthorized)
			return
		}
		RespondInternalErrorWithStack(w, r, err)
	}
	RespondOKJSON(w, r, token)
}

func (h *OAuth2Handler) TokenEndpoint(w http.ResponseWriter, r *http.Request) {
	if mediaType := r.Header.Get(constants.HeaderContentType); mediaType != constants.MimeApplicationXWWWFormURLEncoded {
		RespondErrorPlain(w, r, fmt.Errorf("unsupported media type: %s", mediaType), http.StatusUnsupportedMediaType)
		return
	}

	grantType, err := requiredPostFormField(r, OAuth2GrantType)
	if err != nil {
		RespondErrorPlain(w, r, err, http.StatusBadRequest)
		return
	}

	switch grantType {
	case "":
		RespondErrorPlain(w, r, fmt.Errorf("required field: %s", OAuth2GrantType), http.StatusBadRequest)
		return
	case auth.GrantTypePassword:
		h.handlePasswordFlow(w, r)
		return
	default:
		RespondInternalErrorWithStack(w, r, fmt.Errorf("grant type '%s' is not implemented", grantType))
		return
	}
}

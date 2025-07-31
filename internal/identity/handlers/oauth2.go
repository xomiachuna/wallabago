package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/andriihomiak/wallabago/internal/http/constants"
	"github.com/andriihomiak/wallabago/internal/http/response"
	"github.com/andriihomiak/wallabago/internal/identity"
	"github.com/andriihomiak/wallabago/internal/identity/managers"
	"github.com/andriihomiak/wallabago/internal/identity/storage"
	"github.com/pkg/errors"
)

type OAuth2Handler struct {
	manager *managers.IdentityManager
}

func NewOAuth2Handler(
	manager *managers.IdentityManager,
) *OAuth2Handler {
	return &OAuth2Handler{
		manager: manager,
	}
}

func NewOAuth2HandlerFromDBPool(
	pool *sql.DB,
) *OAuth2Handler {
	return NewOAuth2Handler(
		managers.NewIdentityManager(
			storage.NewCodegenStorage(
				pool,
			),
		),
	)
}

func NewOAuth2HandlerFromStorage(
	sqlStorage storage.SQLStorage,
) *OAuth2Handler {
	return NewOAuth2Handler(
		managers.NewIdentityManager(
			sqlStorage,
		),
	)
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

func requiredPasswordFlowRequest(r *http.Request) (*identity.PasswordFlowRequest, error) {
	clientID, requiredErr := requiredPostFormField(r, OAuth2ClientID)
	if requiredErr != nil {
		return nil, requiredErr
	}
	clientSecret, requiredErr := requiredPostFormField(r, OAuth2ClientSecret)
	if requiredErr != nil {
		return nil, requiredErr
	}
	username, requiredErr := requiredPostFormField(r, OAuth2Username)
	if requiredErr != nil {
		return nil, requiredErr
	}

	password, requiredErr := requiredPostFormField(r, OAuth2Password)
	if requiredErr != nil {
		return nil, requiredErr
	}
	return &identity.PasswordFlowRequest{
		ClientID:     clientID,
		Username:     username,
		ClientSecret: clientSecret,
		Password:     password,
	}, nil
}

func (h *OAuth2Handler) handlePasswordFlow(w http.ResponseWriter, r *http.Request) {
	req, requiredFieldErr := requiredPasswordFlowRequest(r)
	if requiredFieldErr != nil {
		response.RespondErrorPlain(w, r, requiredFieldErr, http.StatusBadRequest)
		return
	}

	token, err := h.manager.PasswordFlow(r.Context(), *req)
	if err != nil {
		authError := &identity.AuthError{}
		if errors.As(err, &authError) {
			response.RespondJSON(w, r, authError, http.StatusUnauthorized)
			return
		}
		response.RespondInternalErrorWithStack(w, r, err)
	}
	response.RespondOKJSON(w, r, token)
}

func (h *OAuth2Handler) TokenEndpoint(w http.ResponseWriter, r *http.Request) {
	if mediaType := r.Header.Get(constants.HeaderContentType); mediaType != constants.MimeApplicationXWWWFormURLEncoded {
		response.RespondErrorPlain(w, r, fmt.Errorf("unsupported media type: '%s', expected: '%s'", mediaType, constants.MimeApplicationXWWWFormURLEncoded), http.StatusUnsupportedMediaType)
		return
	}

	grantType, err := requiredPostFormField(r, OAuth2GrantType)
	if err != nil {
		response.RespondErrorPlain(w, r, err, http.StatusBadRequest)
		return
	}

	switch grantType {
	case "":
		response.RespondErrorPlain(w, r, fmt.Errorf("required field: %s", OAuth2GrantType), http.StatusBadRequest)
		return
	case identity.GrantTypePassword:
		h.handlePasswordFlow(w, r)
		return
	default:
		response.RespondInternalErrorWithStack(w, r, fmt.Errorf("grant type '%s' is not implemented", grantType))
		return
	}
}

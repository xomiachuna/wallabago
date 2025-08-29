package handlers

import (
	"fmt"
	"net/http"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/http/constants"
	"github.com/andriihomiak/wallabago/internal/http/response"
	"github.com/andriihomiak/wallabago/internal/managers"
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

const (
	OAuth2GrantType    = "grant_type"
	OAuth2ClientID     = "client_id"
	OAuth2ClientSecret = "client_secret"
	OAuth2Username     = "username"
	OAuth2Password     = "password"
	OAuth2Scope        = "scope"
	OAuth2RefreshToken = "refresh_token"
)

func requiredPasswordFlowRequest(r *http.Request) (*core.PasswordFlowRequest, error) {
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
	return &core.PasswordFlowRequest{
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
		authError := &core.AuthError{}
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
	case core.GrantTypePassword:
		h.handlePasswordFlow(w, r)
		return
	default:
		response.RespondInternalErrorWithStack(w, r, fmt.Errorf("grant type '%s' is not implemented", grantType))
		return
	}
}

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/andriihomiak/wallabago/internal/core"
	"github.com/andriihomiak/wallabago/internal/http/constants"
	"github.com/andriihomiak/wallabago/internal/http/response"
	"github.com/andriihomiak/wallabago/internal/managers"
	"github.com/pkg/errors"
)

type OAuth2Middleware interface {
	Middleware
}

type oAuth2Middleware struct {
	identity *managers.IdentityManager
}

func NewOAuth2Middleware(identity *managers.IdentityManager) OAuth2Middleware {
	return &oAuth2Middleware{
		identity: identity,
	}
}

var _ OAuth2Middleware = (*oAuth2Middleware)(nil)

type contextAccessTokenKey struct{}

var ctxTokenKey = contextAccessTokenKey{}

func (m *oAuth2Middleware) withToken(r *http.Request, accessToken *core.AccessToken) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxTokenKey, accessToken))
}

func GetAccessToken(r *http.Request) core.AccessToken {
	token, ok := r.Context().Value(ctxTokenKey).(*core.AccessToken)
	if !ok || token == nil {
		panic("tried to access token when none is injected")
	}
	return *token
}

func (m *oAuth2Middleware) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get(constants.HeaderAuthorization)
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) != 2 || !strings.EqualFold(authHeaderParts[0], "bearer") {
			response.RespondErrorPlain(w, r, fmt.Errorf("bad authorization header: '%s'", authHeader), http.StatusUnauthorized)
			return
		}
		tokenPart := authHeaderParts[1]
		accessToken, err := m.identity.Authenticate(r.Context(), tokenPart)
		var authError *core.AuthError
		if errors.As(err, &authError) {
			response.RespondJSON(w, r, authError, http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, m.withToken(r, accessToken))
	})
}

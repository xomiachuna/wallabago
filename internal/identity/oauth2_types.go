package identity

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type TokenType string

const (
	TokenTypeBearer TokenType = "bearer"
)

type (
	ScopeName string
	Scope     string
)

const (
	ScopeEntries ScopeName = "entries"
)

var validScopeNames = []ScopeName{
	ScopeEntries,
}

func DefaultScope() *Scope {
	scope := Scope(ScopeEntries)
	return &scope
}

func NewScope(scopeNames ...ScopeName) (*Scope, error) {
	sb := strings.Builder{}
	for i, scopeName := range scopeNames {
		if !slices.Contains(validScopeNames, scopeName) {
			return nil, errors.WithStack(fmt.Errorf("invalid scope: %s", scopeName))
		}
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(string(scopeName))
	}
	scope := Scope(sb.String())
	return &scope, nil
}

func NewScopeFromString(scope string) (*Scope, error) {
	parts := strings.Split(scope, " ")
	names := make([]ScopeName, 0, len(parts))
	for _, part := range parts {
		names = append(names, ScopeName(part))
	}
	return NewScope(names...)
}

type JWT string

func NewJWT(claims map[string]any, key []byte) (*JWT, error) {
	refreshToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims(claims),
	).SignedString(key)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := JWT(refreshToken)
	return &result, nil
}

func NewRefreshToken(userID, clientID string, key []byte) (*RefreshToken, error) {
	tokenID := uuid.New().String()
	claims := map[string]any{
		"iat": time.Now().Unix(),
		// "exp": time.Now().Add(accessTokenExpiry).Unix(),
		"sub": userID,
		"aud": clientID,
		"jti": tokenID,
	}

	token, err := NewJWT(claims, key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &RefreshToken{Token: *token, ID: tokenID, ClientID: clientID, Revoked: false}, nil
}

type RefreshToken struct {
	Token    JWT
	ID       string
	ClientID string
	Revoked  bool
}

type AccessToken struct {
	Token            JWT       `json:"access_token"`
	ExpiresInSeconds int64     `json:"expires_in"`
	Scope            Scope     `json:"scope"`
	TokenType        TokenType `json:"token_type"`
	Revoked          bool      `json:"-"`
	ID               string    `json:"-"`
	ClientID         string    `json:"-"`
	UserID           string    `json:"-"`
	IssuedAt         time.Time `json:"-"`
}

func NewAccessToken(userID, clientID string, scope Scope, expiration time.Duration, key []byte) (*AccessToken, error) {
	tokenID := uuid.New()
	issuedAt := time.Now()
	claims := map[string]any{
		"iat": issuedAt.Unix(),
		"exp": issuedAt.Add(expiration).Unix(),
		"sub": userID,
		"aud": clientID,
		"jti": tokenID,
	}
	token, err := NewJWT(claims, key)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &AccessToken{
		Token:            *token,
		ExpiresInSeconds: int64(expiration.Seconds()),
		ID:               string(*token),
		Scope:            scope,
		Revoked:          false,
		TokenType:        TokenTypeBearer,
		ClientID:         clientID,
		IssuedAt:         issuedAt,
		UserID:           userID,
	}, nil
}

type AccessTokenResponse struct {
	AccessToken
	RefreshToken JWT `json:"refresh_token,omitempty"`
}

type GrantType string

const (
	GrantTypePassword     = "password"
	GrantTypeRefreshToken = "refresh_token"
)

type AuthErrorName string

const (
	AuthErrorInvaidRequest        = "invalid_request"
	AuthErrorInvalidClient        = "invalid_client"
	AuthErrorInvalidGrant         = "invalid_grant"
	AuthErrorUnauthorizedClient   = "unauthorized_client"
	AuthErrorUnsupportedGrantType = "unsupported_grant_type"
	AuthErrorInvalidScope         = "invalid_scope"
)

type AuthError struct {
	ErrorName        AuthErrorName `json:"error_name,omitempty"`
	ErrorDescription string        `json:"error_description,omitempty"`
	ErrorURI         string        `json:"error_uri,omitempty"`
}

func (e *AuthError) Error() string {
	return string(e.ErrorName)
}

//nolint:errcheck //only to make sure it implements error
var _ error = (*AuthError)(nil)

type RefreshTokenFlowRequest struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
}

type PasswordFlowRequest struct {
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	Scope        string
}

type UserInfo struct {
	ID           string
	Username     string
	Email        string
	PasswordHash []byte
}

type Client struct {
	ID     string
	Secret string
}

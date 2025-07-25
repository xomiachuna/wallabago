package auth

import (
	"fmt"
	"slices"
	"strings"

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

type JWT string

type TokenResponse struct {
	AccessToken  JWT
	ExpiresIn    uint
	RefreshToken JWT
	Scope        Scope
	TokenType    TokenType
}

type GrantType string

const (
	GrantTypePassword     = "password"
	GrantTypeRefreshToken = "refresh_token"
)

//nolint:revive //AuthError is described in RFC
type AuthError struct {
	ErrorName        string
	ErrorDescription string
	ErrorURI         string
}

func (e *AuthError) Error() string {
	return e.ErrorName
}

//nolint:errcheck //only to make sure it implements error
var _ error = (*AuthError)(nil)

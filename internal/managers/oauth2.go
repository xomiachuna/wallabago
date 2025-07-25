package managers

import (
	"context"
	"fmt"

	"github.com/andriihomiak/wallabago/internal/auth"
	"github.com/pkg/errors"
)

func NewOAuth2Manager() (*OAuth2Manager, error) {
	return &OAuth2Manager{}, nil
}

type OAuth2Manager struct{}

type RefreshTokenFlowRequest struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
}

//nolint:revive //todo
func (m *OAuth2Manager) PasswordFlow(ctx context.Context, clientId, clientSecret, username, password string) (*auth.TokenResponse, error) {
	return nil, errors.WithStack(fmt.Errorf("password flow not implemented yet"))
}

//nolint:revive //todo
func (m *OAuth2Manager) RefreshTokenFlow(ctx context.Context, args RefreshTokenFlowRequest) (*auth.TokenResponse, error) {
	return nil, errors.WithStack(fmt.Errorf("refresh token flow not implemented yet"))
}

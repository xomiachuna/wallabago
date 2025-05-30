package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/codes"
)

type User struct {
	Name    string
	IsAdmin bool
}

type Token struct{}

type Permission string

type BasicAuthnToken string

type AuthenticationEngine interface {
	BasicAuthn(context.Context, BasicAuthnToken) (User, error)
}

type HardcodedAuthnEngine struct {
	username, password string
}

type ErrInvalidCredentials struct{}

func (e *ErrInvalidCredentials) Error() string {
	return "Invalid credentials"
}

func (e *HardcodedAuthnEngine) BasicAuthn(ctx context.Context, token BasicAuthnToken) (User, error) {
	ctx, span := authTracer.Start(ctx, "basic-auth")
	defer span.End()
	decoded, err := base64.StdEncoding.DecodeString(string(token))
	if err != nil {
		err = fmt.Errorf("basic auth token decoding error: %w", err)
		span.SetStatus(codes.Error, err.Error())
		return User{}, err
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) != 2 {
		err = fmt.Errorf("basic token must have 2 parts delimited by ':'")
		span.SetStatus(codes.Error, err.Error())
		return User{}, err
	}
	if parts[0] == e.username && parts[1] == e.password {
		span.SetStatus(codes.Ok, "")
		return User{Name: parts[0], IsAdmin: true}, nil
	}
	err = &ErrInvalidCredentials{}
	span.SetStatus(codes.Error, err.Error())
	return User{}, err
}

type AuthorizationEngine interface {
	HasPermissions(context.Context, User, []Permission) (bool, error)
}

type AuthEngine interface {
	AuthenticationEngine
	AuthorizationEngine
}

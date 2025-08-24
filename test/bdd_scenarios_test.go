package bdd_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/cucumber/godog"
)

func givenThereIsADefaultClientBoootstrapped() error {
	// assumed to be true
	return nil
}

func givenThereIsAnAdminAccountBootstrapped() error {
	// assumed to be true
	return nil
}

func givenBootstrapAccountCredentialsAreValid() error {
	// passed in the default context
	return nil
}

type tokenResponseKey struct {}

type tokenResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn int `json:"expires_in"`
    RefreshToken string `json:"refresh_token"`
    Scope string `json:"scope"`
    TokenType string `json:"token_type"`
}

func authenthicateWithCredentialsViaClientCredentialsFlow(ctx context.Context, userCreds userCredentials, clientCreds clientCredentials) (context.Context, error) {
	tokenEndpoint, err := makeRequestUrl(ctx, "/oauth/v2/token")
	if err != nil {
		return ctx, err
	}
	client := http.Client{}
	resp, err := client.PostForm(tokenEndpoint, url.Values{
		"username":      []string{userCreds.username},
		"password":      []string{userCreds.password},
		"client_id":     []string{clientCreds.id},
		"client_secret": []string{clientCreds.secret},
		"grant_type":    []string{"password"},
	})
	if err != nil {
		return ctx, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, err
	}
	logger.DebugContext(ctx, "Received response", "statusCode", resp.StatusCode, "body", string(body), "headers", resp.Header)
    response := tokenResponse{}
    err = json.Unmarshal(body, &response)
	if err != nil {
		return ctx, err
	}
    return context.WithValue(ctx, tokenResponseKey{}, response), nil
}

func givenIAmAuthenticatedAsAdmin(ctx context.Context) (context.Context, error) {
	bootstrapCreds, ok := ctx.Value(bootstrapCredentialsKey{}).(userCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract bootstrap credentials")
	}

	bootstrapClient, ok := ctx.Value(bootstrapClientKey{}).(clientCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract bootstrap client")
	}
    return authenthicateWithCredentialsViaClientCredentialsFlow(ctx, bootstrapCreds, bootstrapClient)
}

func thenIAmPreventedFromDeletingTheAccount() error {
	return godog.ErrUndefined
}

func thenIAmSuccessfullyAuthenticatedAsAdmin(ctx context.Context) (context.Context, error) {
	_, ok := ctx.Value(tokenResponseKey{}).(tokenResponse)
	if !ok {
		return ctx, fmt.Errorf("unable to obtain token response")
	}
	return ctx, nil
}

func whenICreateANewAccount() error {
	return godog.ErrUndefined
}

func whenITryToDeleteAccount(which string) error {
	return godog.ErrUndefined
}

func makeRequestUrl(ctx context.Context, path string) (string, error) {
	addr, ok := ctx.Value(serverAddrKey{}).(string)
	if !ok {
		return "", fmt.Errorf("failed to extract server address from context")
	}
	return fmt.Sprintf("http://%s%s", addr, path), nil
}

func whenIUseBootstrapCredentialsToAuthenticate(ctx context.Context) (context.Context, error) {
	bootstrapCreds, ok := ctx.Value(bootstrapCredentialsKey{}).(userCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract bootstrap credentials")
	}

	bootstrapClient, ok := ctx.Value(bootstrapClientKey{}).(clientCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract bootstrap client")
	}
    return authenthicateWithCredentialsViaClientCredentialsFlow(ctx, bootstrapCreds, bootstrapClient)
}

func givenAnotherAccountExists() error {
	return godog.ErrUndefined
}

func thenAccountExistenceIsAsExpected() error {
	return godog.ErrUndefined
}

var logger *slog.Logger

func init(){
    logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
}

func InitializeScenario(ctx *godog.ScenarioContext, infra *testInfra) {
	ctx.Given(`there is a default client bootstrapped`, givenThereIsADefaultClientBoootstrapped)
	ctx.Given(`there is an admin account bootstrapped`, givenThereIsAnAdminAccountBootstrapped)
	ctx.Given(`bootstrap account credentials are valid`, givenBootstrapAccountCredentialsAreValid)
	ctx.Given(`I am authenticated as admin`, givenIAmAuthenticatedAsAdmin)
	ctx.Given(`there exists another (user|admin) account`, givenAnotherAccountExists)

	ctx.When(`I use bootstrap credentials to authenticate`, whenIUseBootstrapCredentialsToAuthenticate)
	ctx.When(`I create a new (user|admin) account`, whenICreateANewAccount)
	ctx.When(`I (?:try to )?delete (my|bootstrapped admin|that) account`, whenITryToDeleteAccount)

	ctx.Then(`I am successfully authenticated as admin`, thenIAmSuccessfullyAuthenticatedAsAdmin)
	ctx.Then(`I am prevented from deleting the account`, thenIAmPreventedFromDeletingTheAccount)
	ctx.Then(`(my|bootstrapped admin|admin|user) account (exists|still exists|no longer exists)`, thenAccountExistenceIsAsExpected)
}

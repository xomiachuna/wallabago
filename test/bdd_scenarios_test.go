package bdd_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

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

func givenIAmAuthenticatedAsAdmin() error {
	return godog.ErrPending
}

func thenIAmPreventedFromDeletingTheAccount() error {
	return godog.ErrPending
}

func thenIAmSuccessfullyAuthenticatedAsAdmin(ctx context.Context) (context.Context, error) {
	response, ok := ctx.Value(authenticationKey{}).(authenticationResponse)
	if !ok {
		return ctx, fmt.Errorf("failed to extract auth response from context")
	}
	if response.statusCode != http.StatusOK {
		return ctx, fmt.Errorf("received non-200 status code: %d", response.statusCode)
	}
	return ctx, nil
}

func whenICreateANewAccount() error {
	return godog.ErrPending
}

func whenITryToDeleteAccount(which string) error {
	return godog.ErrPending
}

func makeRequestUrl(ctx context.Context, path string) (string, error) {
	addr, ok := ctx.Value(serverAddrKey{}).(string)
	if !ok {
		return "", fmt.Errorf("failed to extract server address from context")
	}
	return fmt.Sprintf("http://%s%s", addr, path), nil
}

type (
	authenticationKey      struct{}
	authenticationResponse struct {
		statusCode int
		body       []byte
	}
)

func whenIUseBootstrapCredentialsToAuthenticate(ctx context.Context) (context.Context, error) {
	tokenEndpoint, err := makeRequestUrl(ctx, "/oauth/v2/token")
	if err != nil {
		return ctx, err
	}

	bootstrapCreds, ok := ctx.Value(bootstrapCredentialsKey{}).(userCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract bootstrap credentials")
	}

	bootstrapClient, ok := ctx.Value(bootstrapClientKey{}).(clientCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract bootstrap client")
	}

	client := http.Client{}

	resp, err := client.PostForm(tokenEndpoint, url.Values{
		"username":      []string{bootstrapCreds.username},
		"password":      []string{bootstrapCreds.password},
		"client_id":     []string{bootstrapClient.id},
		"client_secret": []string{bootstrapClient.secret},
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
	slog.DebugContext(ctx, "Received response", "statusCode", resp.StatusCode, "body", string(body), "headers", resp.Header)
	return context.WithValue(ctx, authenticationKey{}, authenticationResponse{
		statusCode: resp.StatusCode,
		body:       body,
	}), nil
}

func givenAnotherAccountExists() error {
	return godog.ErrPending
}

func thenAccountExistenceIsAsExpected() error {
	return godog.ErrPending
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

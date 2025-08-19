package bdd_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/cucumber/godog"
)

type identityManagerKey struct{}

type (
	clientIdKey     struct{}
	clientSecretKey struct{}
	usernameKey     struct{}
	passwordKey     struct{}
)

func theFollowingUsersExist(ctx context.Context, data *godog.Table) (context.Context, error) {
	return ctx, godog.ErrPending
}

func theFollowingClientsExist(ctx context.Context, data *godog.Table) (context.Context, error) {
	return ctx, godog.ErrPending
}

func setClientId(ctx context.Context, id string) (context.Context, error) {
	return ctx, godog.ErrPending
}

func setClientSecret(ctx context.Context, id string) (context.Context, error) {
	return ctx, godog.ErrPending
}

func setUsername(ctx context.Context, id string) (context.Context, error) {
	return ctx, godog.ErrPending
}

func setPassword(ctx context.Context, id string) (context.Context, error) {
	return ctx, godog.ErrPending
}

func tokenIsRequestedWithClientCredentialsFlow(ctx context.Context) (context.Context, error) {
	return ctx, godog.ErrPending
}

func refreshTokenShouldBeReturned(ctx context.Context) (context.Context, error) {
	return ctx, godog.ErrPending
}

func noErrorShouldBeReturned(ctx context.Context) (context.Context, error) {
	return ctx, godog.ErrPending
}

func setupInfra() {
	// start containers
	slog.Info("Started containers")
}

func tearDownInfra() {
	// stop containers
	slog.Info("Stopped containers")
}

func initSuite(tsc *godog.TestSuiteContext) {
	tsc.BeforeSuite(setupInfra)
	tsc.AfterSuite(tearDownInfra)
}

func initScenarios(ctx *godog.ScenarioContext) {
	// background
	ctx.Step(`^the following clients exist:$`, theFollowingClientsExist)
	ctx.Step(`^the following users exist:$`, theFollowingUsersExist)
	// given
	ctx.Step(`^client id is "([^"]*)"$`, setClientId)
	ctx.Step(`^client secret is "([^"]*)"$`, setClientSecret)
	ctx.Step(`^password "([^"]*)"$`, setPassword)
	ctx.Step(`^username is "([^"]*)"$`, setUsername)
	// events
	ctx.Step(`^token is requested with client credentials flow$`, tokenIsRequestedWithClientCredentialsFlow)
	// assertions

	ctx.Step(`^no error should be returned$`, noErrorShouldBeReturned)
	ctx.Step(`^refresh token should be returned$`, refreshTokenShouldBeReturned)
}

func TestScenarios(t *testing.T) {
	suite := godog.TestSuite{
		TestSuiteInitializer: initSuite,
		ScenarioInitializer:  initScenarios,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features"},
			TestingT: t,
			Output:   os.Stderr,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("failed to run feature tests")
	}
}

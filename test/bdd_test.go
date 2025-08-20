package bdd_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/andriihomiak/wallabago/internal/app"
	"github.com/andriihomiak/wallabago/internal/http"
	"github.com/cucumber/godog"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type (
	clientIDKey     struct{}
	clientSecretKey struct{}
	usernameKey     struct{}
	passwordKey     struct{}

	identityManagerKey struct{}
)

type errNotImplemented struct{}

func (e errNotImplemented) Error() string {
	return "not implemented"
}

func theFollowingUsersExist(ctx context.Context, data *godog.Table) (context.Context, error) {
	return ctx, errNotImplemented{}
}

func theFollowingClientsExist(ctx context.Context, data *godog.Table) (context.Context, error) {
	return ctx, errNotImplemented{}
}

func setClientId(ctx context.Context, id string) (context.Context, error) {
	return ctx, errNotImplemented{}
}

func setClientSecret(ctx context.Context, id string) (context.Context, error) {
	return ctx, errNotImplemented{}
}

func setUsername(ctx context.Context, id string) (context.Context, error) {
	return ctx, errNotImplemented{}
}

func setPassword(ctx context.Context, id string) (context.Context, error) {
	return ctx, errNotImplemented{}
}

func tokenIsRequestedWithClientCredentialsFlow(ctx context.Context) (context.Context, error) {
	return ctx, errNotImplemented{}
}

func refreshTokenShouldBeReturned(ctx context.Context) (context.Context, error) {
	return ctx, errNotImplemented{}
}

func noErrorShouldBeReturned(ctx context.Context) (context.Context, error) {
	return ctx, errNotImplemented{}
}

type testInfra struct {
	pg            *postgres.PostgresContainer
	server        *http.Server
	cancelContext context.CancelFunc

	teminationMu sync.Mutex
	terminated   bool
}

func newTestInfra() *testInfra {
	return &testInfra{}
}

func (ti *testInfra) setup(ctx context.Context, cancelContext context.CancelFunc) error {
	ti.cancelContext = cancelContext
	slog.Info("Starting postgres in a testcontainer")
	pgContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithDatabase("bdd-test-database"),
		postgres.WithUsername("bdd-test-username"),
		postgres.WithPassword("bdd-test-password"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return err
	}
	ti.pg = pgContainer
	connString, err := pgContainer.ConnectionString(ctx, "sslmode=disable", "application_name=bdd_test")
	if err != nil {
		return err
	}
	slog.Info("Postgres started", "conn", connString)

	slog.Info("Running migrations")
	err = fmt.Errorf("migrations not implemented")
	return err

	server, err := http.NewServer(context.Background(), app.Config{
		Addr:                   ":29999",
		InstrumentationEnabled: false,
		DBConnectionString:     connString,
	})
	if err != nil {
		return err
	}
	ti.server = server
	go func() {
		server.Start(ctx)
	}()
	// todo - check for readiness?
	return nil
}

func (ti *testInfra) teardown(cause error) error {
	ti.teminationMu.Lock()
	defer ti.teminationMu.Unlock()
	if ti.terminated {
		return nil
	}
	slog.Info("Tearing down test infra", "cause", cause)

	if ti.cancelContext != nil {
		slog.Info("Cancelling context")
		ti.cancelContext()
		slog.Info("Context cancelled")
	} else {
		slog.Warn("Unable to cancel context - no cancellation func available")
	}

	if ti.pg != nil {
		err := ti.pg.Terminate(context.TODO())
		if err != nil {
			slog.Error("Failed to teminate pg container", "err", err)
			return err
		}
		slog.Info("pg container terminated")
	} else {
		slog.Warn("Unable to ternminate pg: nothing to terminate")
	}
	ti.terminated = true
	return nil
}

func initSuite(tsc *godog.TestSuiteContext) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	infra := newTestInfra()
	tsc.BeforeSuite(func() {
		err := infra.setup(ctx, cancelFunc)
		if err != nil {
			infra.teardown(err)
		}
	})
	tsc.AfterSuite(func() { infra.teardown(nil) })
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

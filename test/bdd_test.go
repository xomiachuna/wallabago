package bdd_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	nethttp "net/http"

	"github.com/andriihomiak/wallabago/internal/app"
	"github.com/andriihomiak/wallabago/internal/http"
	"github.com/cucumber/godog"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testInfra struct {
	cancelContext context.CancelFunc
	stack         *compose.DockerCompose
	server        *http.Server
}

func newTestInfra() *testInfra {
	return &testInfra{}
}

func (ti *testInfra) setup(ctx context.Context, cancelContext context.CancelFunc) error {
	ti.cancelContext = cancelContext
	// run compose
	stack, err := compose.NewDockerComposeWith(
		compose.WithStackFiles("../deployments/docker-compose/docker-compose.yaml"),
	)
	if err != nil {
		return err
	}
	ti.stack = stack
	err = stack.WaitForService("migrations", wait.ForExit()).Up(ctx, compose.RunServices("postgres", "migrations"))
	if err != nil {
		return err
	}

	addr := "0.0.0.0:29999"

	server, err := http.NewServer(context.Background(), app.Config{
		Addr:                   addr,
		InstrumentationEnabled: false,
		DBConnectionString:     "postgresql://wallabago-api:wallabago@localhost:25432/wallabago-db?sslmode=disable&application_name=wallabago-api-client",
		BootstrapClientID:      "web",
		BootstrapClientSecret:  "web",
		BootstrapAdminPassword: "admin",
		BootstrapAdminUsername: "admin",
		BootstrapAdminEmail:    "admin@admin.co",
	})
	if err != nil {
		return err
	}

	ti.server = server

	go func() {
		server.Start(ctx)
	}()

	startupCtx, cancelFunc := context.WithTimeout(ctx, time.Millisecond*15000)
	defer cancelFunc()

	ready := make(chan struct{})
	go func() {
		client := nethttp.Client{}
		// get index
		url := fmt.Sprintf("http://%s/", addr)
		for {
			resp, err := client.Get(url)
			if err != nil {
				slog.Debug("Error while checking server health", "err", err)
				time.Sleep(time.Millisecond * 1000)
				continue
			}
			defer resp.Body.Close()
			slog.Debug("Server is online")
			break
		}
		ready <- struct{}{}
	}()
	select {
	case <-ready:
		return nil
	case <-startupCtx.Done():
		return errors.Wrap(startupCtx.Err(), "readiness probe failed")
	}
}

func (ti *testInfra) teardown(cause error) error {
	slog.Info("Tearing down test infra", "cause", cause)

	slog.Info("Cancelling context")
	ti.cancelContext()
	slog.Info("Context cancelled")

	err := ti.stack.Down(context.TODO())

	slog.Info("Infra tear down finished")
	return err
}

type serverAddrKey struct{}

type bootstrapCredentialsKey struct{}

type userCredentials struct {
	username, password string
}

type bootstrapClientKey struct{}

type clientCredentials struct {
	id     string
	secret string
}

func TestBDDScenarios(t *testing.T) {
	infraCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	infra := newTestInfra()
	err := infra.setup(infraCtx, cancelFunc)
	if err != nil {
		infra.teardown(err)
	}
	defer infra.teardown(nil)

	ctx := context.Background()
	ctx = context.WithValue(ctx, serverAddrKey{}, infra.server.App().Addr())
	ctx = context.WithValue(ctx, bootstrapCredentialsKey{}, userCredentials{
		username: infra.server.App().Config().BootstrapAdminUsername,
		password: infra.server.App().Config().BootstrapAdminPassword,
	})
	ctx = context.WithValue(ctx, bootstrapClientKey{}, clientCredentials{
		id:     infra.server.App().Config().BootstrapClientID,
		secret: infra.server.App().Config().BootstrapClientSecret,
	})

	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) { InitializeScenario(sc, infra) },
		Options: &godog.Options{
			Format:         "pretty",
			Paths:          []string{"../features"},
			TestingT:       t,
			Output:         os.Stderr,
			Strict:         true,
			DefaultContext: ctx,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("failed to run feature tests")
	}
	t.Log("BDD suite finished")
}

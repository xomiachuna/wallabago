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
	})
	if err != nil {
		return err
	}
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

func initScenarios(ctx *godog.ScenarioContext) {
	// background
	// given
	// events
	// assertions
}

func TestBDDScenarios(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	infra := newTestInfra()
	err := infra.setup(ctx, cancelFunc)
	if err != nil {
		infra.teardown(err)
	}
	defer infra.teardown(nil)
	suite := godog.TestSuite{
		ScenarioInitializer: initScenarios,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features"},
			TestingT: t,
			Output:   os.Stderr,
			Strict:   true,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("failed to run feature tests")
	}
	t.Log("BDD suite finished")
}

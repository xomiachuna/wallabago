package bdd_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	nethttp "net/http"

	"github.com/andriihomiak/wallabago/internal/app"
	"github.com/andriihomiak/wallabago/internal/http"
	"github.com/cucumber/godog"
	"github.com/pkg/errors"
)

type testInfra struct {
	cancelContext context.CancelFunc
	server        *http.Server
	dbPath        string
}

func newTestInfra() *testInfra {
	return &testInfra{}
}

func (ti *testInfra) setup(ctx context.Context, cancelContext context.CancelFunc) error {
	ti.cancelContext = cancelContext

	// Create temporary database for testing
	tempDir := os.TempDir()
	dbPath := filepath.Join(tempDir, fmt.Sprintf("wallabago-test-%d.db", time.Now().UnixNano()))
	ti.dbPath = dbPath

	addr := "0.0.0.0:29999"

	server, err := http.NewServer(ctx, app.Config{
		Addr:                   addr,
		InstrumentationEnabled: false,
		DBPath:                 dbPath,
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
		err := server.Start(ctx)
		if err != nil {
			slog.Warn("Server exited", "cause", err.Error())
		}
	}()

	startupCtx, cancelFunc := context.WithTimeout(ctx, time.Millisecond*15000)
	defer cancelFunc()

	ready := make(chan struct{})
	client := nethttp.Client{}
	url := fmt.Sprintf("http://%s/", addr)
	req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodGet, url, nethttp.NoBody)
	if err != nil {
		return errors.WithStack(err)
	}
	go func() {
		// get index
		for {
			resp, err := client.Do(req)
			if err != nil {
				slog.Debug("Error while checking server health", "err", err)
				time.Sleep(time.Millisecond * 1000)
				continue
			}
			resp.Body.Close()
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

func (ti *testInfra) teardown(cause error) {
	slog.Info("Tearing down test infra", "cause", cause)

	slog.Info("Cancelling context")
	ti.cancelContext()
	slog.Info("Context cancelled")

	// Clean up test database
	if ti.dbPath != "" {
		slog.Info("Removing test database", "path", ti.dbPath)
		err := os.Remove(ti.dbPath)
		if err != nil && !os.IsNotExist(err) {
			slog.Warn("Failed to remove test database", "err", err)
		}
		// Also remove WAL files if they exist
		os.Remove(ti.dbPath + "-shm")
		os.Remove(ti.dbPath + "-wal")
	}

	slog.Info("Infra tear down finished")
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
		t.Fatal(err)
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
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:         "pretty",
			Paths:          []string{"../features"},
			TestingT:       t,
			Output:         os.Stderr,
			Strict:         true,
			DefaultContext: ctx,
			StopOnFailure:  true,
			Concurrency:    1,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("bdd tests failed")
	}
	t.Log("BDD suite finished")
}

package instrumentation

import (
	"log/slog"
	"os"
	"runtime/debug"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func initLogger() {
	otelScopeName := "wallabago"
	otelScopeVersion := "0.0.0"
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		otelScopeName = buildInfo.Main.Path
		otelScopeVersion = buildInfo.Main.Version
	}
	otelHandler := otelslog.NewHandler(otelScopeName, otelslog.WithVersion(otelScopeVersion))
	stderrHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false})
	logger := slog.New(MultiHandler(otelHandler, stderrHandler))
	slog.SetDefault(logger)
}

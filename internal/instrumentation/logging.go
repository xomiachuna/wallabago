package instrumentation

import (
	"log/slog"
	"os"
	"runtime/debug"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

var Logger *slog.Logger

func InitLogger() *slog.Logger {
	otelScopeName := "wallabago"
	otelScopeVersion := "0.0.0"
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		otelScopeName = buildInfo.Main.Path
		otelScopeVersion = buildInfo.Main.Version
	}
	otelHandler := otelslog.NewHandler(otelScopeName, otelslog.WithVersion(otelScopeVersion))
	stderrHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false})
	Logger = slog.New(MultiHandler(otelHandler, stderrHandler))
	Logger.Debug("Created logger", "logger", Logger)
	slog.SetDefault(Logger)
	return Logger
}

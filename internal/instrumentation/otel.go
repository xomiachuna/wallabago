package instrumentation

import (
	"runtime/debug"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

func InitTracer() trace.Tracer {
	otelScopeName := "wallabago"
	otelScopeVersion := "0.0.0"
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		otelScopeName = buildInfo.Main.Path
		otelScopeVersion = buildInfo.Main.Version
	}
	Tracer = otel.Tracer(otelScopeName, trace.WithInstrumentationVersion(otelScopeVersion))
	return Tracer
}

package instrumentation

import (
	"context"
	"runtime/debug"

	"go.opentelemetry.io/otel"

	stderrors "errors"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otellog "go.opentelemetry.io/otel/log"
	otelgloballog "go.opentelemetry.io/otel/log/global"
	otelpropagation "go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var Tracer oteltrace.Tracer

func initTracer() {
	otelScopeName := "wallabago"
	otelScopeVersion := "0.0.0"
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		otelScopeName = buildInfo.Main.Path
		otelScopeVersion = buildInfo.Main.Version
	}
	Tracer = otel.Tracer(otelScopeName, oteltrace.WithInstrumentationVersion(otelScopeVersion))
}

func init() {
	initTracer()
}

func newPropagator() otelpropagation.TextMapPropagator {
	return otelpropagation.NewCompositeTextMapPropagator(
		otelpropagation.Baggage{},
		otelpropagation.TraceContext{},
	)
}

func newResource(ctx context.Context) (*sdkresource.Resource, error) {
	baseResource, err := sdkresource.Merge(
		sdkresource.Default(),
		sdkresource.NewWithAttributes(
			otelsemconv.SchemaURL,
			otelsemconv.ServiceName("wallabago-api"),
		),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	extendedInfo, err := sdkresource.New(
		ctx,
		sdkresource.WithOS(),
		sdkresource.WithContainer(),
		sdkresource.WithContainerID(),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	merged, err := sdkresource.Merge(extendedInfo, baseResource)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return merged, nil
}

func newTracerProvider(ctx context.Context, resource *sdkresource.Resource) (*sdktrace.TracerProvider, error) {
	remoteExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(remoteExporter),
	), nil
}

func newMeterProvider(ctx context.Context, resource *sdkresource.Resource) (*sdkmetric.MeterProvider, error) {
	remoteExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(remoteExporter)),
	), nil
}

// scopeInjectingSpanProcessor is a [sdklog.Processor]
// that injects [instrumentation.Scope] via [attribute]s using [otelsemconv].
type scopeInjectingLogProcessor struct{}

func (lp *scopeInjectingLogProcessor) OnEmit(_ context.Context, record *sdklog.Record) error {
	scope := record.InstrumentationScope()
	record.SetAttributes(
		otellog.String(string(otelsemconv.OTelScopeNameKey), scope.Name),
		otellog.String(string(otelsemconv.OTelScopeVersionKey), scope.Version),
	)
	return nil
}
func (lp *scopeInjectingLogProcessor) Shutdown(_ context.Context) error   { return nil }
func (lp *scopeInjectingLogProcessor) ForceFlush(_ context.Context) error { return nil }

func newLoggerProvider(ctx context.Context, resource *sdkresource.Resource) (*sdklog.LoggerProvider, error) {
	remoteExporter, err := otlploggrpc.New(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return sdklog.NewLoggerProvider(
		sdklog.WithResource(resource),
		sdklog.WithProcessor(&scopeInjectingLogProcessor{}),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(remoteExporter)),
	), nil
}

func SetupOtelSDK(ctx context.Context) (shutdownOtel func(ctx context.Context) error, err error) {
	// holds the callbacks necessary for proper otel shutdown
	doOtelShutdownCallbacks := make([]func(context.Context) error, 0)

	// properly shutdown everything set up by otel
	// by calling every callback
	shutdownOtel = func(ctx context.Context) error {
		var shutdownErr error
		for i, callback := range doOtelShutdownCallbacks {
			if err := callback(ctx); err != nil {
				shutdownErr = errors.Wrapf(err, "error when calling cancel callback #%d", i)
			}
		}
		return shutdownErr
	}

	propagator := newPropagator()
	otel.SetTextMapPropagator(propagator)

	resource, err := newResource(ctx)
	if err != nil {
		return nil, stderrors.Join(err, shutdownOtel(ctx))
	}

	tracerProvider, err := newTracerProvider(ctx, resource)
	if err != nil {
		return nil, stderrors.Join(err, shutdownOtel(ctx))
	}
	otel.SetTracerProvider(tracerProvider)
	doOtelShutdownCallbacks = append(doOtelShutdownCallbacks, tracerProvider.Shutdown)

	loggerProvider, err := newLoggerProvider(ctx, resource)
	if err != nil {
		return nil, stderrors.Join(err, shutdownOtel(ctx))
	}
	otelgloballog.SetLoggerProvider(loggerProvider)
	doOtelShutdownCallbacks = append(doOtelShutdownCallbacks, loggerProvider.Shutdown)

	meterProvider, err := newMeterProvider(ctx, resource)
	if err != nil {
		return nil, stderrors.Join(err, shutdownOtel(ctx))
	}
	otel.SetMeterProvider(meterProvider)
	doOtelShutdownCallbacks = append(doOtelShutdownCallbacks, meterProvider.Shutdown)

	return shutdownOtel, nil
}

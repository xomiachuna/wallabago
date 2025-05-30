package main

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// setupOtelSdk sets up the trace, logs and metrics providers and a propagator
// and returns a callback that stops the otel machinery
func setupOtelSdk(ctx context.Context) (shutdownOtel func(ctx context.Context) error, err error) {
	var onShutdownCallbacks []func(context.Context) error
	shutdownOtel = func(ctx context.Context) error {
		var err error
		for _, cb := range onShutdownCallbacks {
			err = errors.Join(err, cb(ctx))
		}
		onShutdownCallbacks = nil
		return err
	}

	shutdownAndJoinErrors := func(err error) error {
		return errors.Join(err, shutdownOtel(ctx))
	}

	resource, err := newResource(ctx)
	if err != nil {
		err = shutdownAndJoinErrors(err)
		return
	}

	propagator := newPropagator()
	otel.SetTextMapPropagator(propagator)

	tracerProvider, err := newTracerProvider(ctx, resource)
	if err != nil {
		err = shutdownAndJoinErrors(err)
		return
	}
	otel.SetTracerProvider(tracerProvider)
	onShutdownCallbacks = append(onShutdownCallbacks, tracerProvider.Shutdown)

	loggerProvider, err := newLoggerProvider(ctx, resource)
	if err != nil {
		err = shutdownAndJoinErrors(err)
		return
	}
	global.SetLoggerProvider(loggerProvider)
	onShutdownCallbacks = append(onShutdownCallbacks, loggerProvider.Shutdown)

	meterProvider, err := newMeterProvider(ctx, resource)
	if err != nil {
		err = shutdownAndJoinErrors(err)
		return
	}
	// otel.SetMeterProvider(meterProvider)
	onShutdownCallbacks = append(onShutdownCallbacks, meterProvider.Shutdown)
	return
}

// scopeInjectingSpanProcessor is a [trace.SpanProcessor]
// that injects [instrumentation.Scope] via [attribute]s using [semconv]
type scopeInjectingSpanProcessor struct{}

func (sp *scopeInjectingSpanProcessor) OnStart(parent context.Context, s trace.ReadWriteSpan) {
	var scope instrumentation.Scope = s.InstrumentationScope()
	s.SetAttributes(
		semconv.OTelScopeName(scope.Name),
		semconv.OTelScopeVersion(scope.Version),
	)
}
func (sp *scopeInjectingSpanProcessor) OnEnd(s trace.ReadOnlySpan)           {}
func (sp *scopeInjectingSpanProcessor) Shutdown(ctx context.Context) error   { return nil }
func (sp *scopeInjectingSpanProcessor) ForceFlush(ctx context.Context) error { return nil }

// scopeInjectingSpanProcessor is a [log.Processor]
// that injects [instrumentation.Scope] via [attribute]s using [semconv]
type scopeInjectingLogProcessor struct{}

func (lp *scopeInjectingLogProcessor) OnEmit(ctx context.Context, record *log.Record) error {
	scope := record.InstrumentationScope()
	record.SetAttributes(
		otellog.String(string(semconv.OTelScopeNameKey), scope.Name),
		otellog.String(string(semconv.OTelScopeVersionKey), scope.Version),
	)
	return nil
}
func (lp *scopeInjectingLogProcessor) Shutdown(ctx context.Context) error   { return nil }
func (lp *scopeInjectingLogProcessor) ForceFlush(ctx context.Context) error { return nil }

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider(ctx context.Context, resource *resource.Resource) (*trace.TracerProvider, error) {
	remoteExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc trace exporter: %w", err)
	}
	return trace.NewTracerProvider(
		trace.WithResource(resource),
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSpanProcessor(&scopeInjectingSpanProcessor{}),
		trace.WithBatcher(remoteExporter),
	), nil
}

func newResource(ctx context.Context) (*resource.Resource, error) {
    res, err :=  resource.Merge(
        resource.Default(),
        resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName("wallabago-api"),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create resource: %w", err)
    }
    extended, err := resource.New(
        ctx, 
        resource.WithOS(),
        resource.WithHost(),
        resource.WithProcessCommandArgs(),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create resource: %w", err)
    }
    return resource.Merge(res, extended)
}

func newLoggerProvider(ctx context.Context, resource *resource.Resource) (*log.LoggerProvider, error) {
	stdoutExporter, err := stdoutlog.New(
		stdoutlog.WithPrettyPrint(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout log exporter: %w", err)
	}

	remoteExporter, err := otlploggrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc log exporter: %w", err)
	}

	return log.NewLoggerProvider(
		log.WithResource(resource),
		log.WithProcessor(&scopeInjectingLogProcessor{}),
		log.WithProcessor(log.NewBatchProcessor(stdoutExporter)),
		log.WithProcessor(log.NewBatchProcessor(remoteExporter)),
	), nil
}

func newMeterProvider(ctx context.Context, resource *resource.Resource) (*metric.MeterProvider, error) {
	stdoutExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout metric exporter: %w", err)
	}

	remoteExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc metric exporter: %w", err)
	}

	return metric.NewMeterProvider(
		metric.WithResource(resource),
		metric.WithReader(metric.NewPeriodicReader(stdoutExporter)),
		metric.WithReader(metric.NewPeriodicReader(remoteExporter)),
	), nil

}

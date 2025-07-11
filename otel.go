package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"go.opentelemetry.io/contrib/bridges/otelslog"
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
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)


var otelScopeName = "xomiachuna.com/wallabago-api"
var otelScopeVersion = "0.0.1"
var tracer = otel.Tracer(otelScopeName, trace.WithInstrumentationVersion(otelScopeVersion))
    var logger = slog.New(MultiHandler(
	otelslog.NewHandler(
		otelScopeName,
		otelslog.WithVersion(otelScopeVersion),
	),
	slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}),
))

var authTracer = otel.Tracer(fmt.Sprintf("%s/auth", otelScopeName), trace.WithInstrumentationVersion(otelScopeVersion))

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

type ServerTimingHeaderExporter struct {
	mu            sync.RWMutex
	stopped       bool
	finishedSpans map[string]sdktrace.ReadOnlySpan
}

func (e *ServerTimingHeaderExporter) GetServerTimingHeaderValue(ctx context.Context) (string, error) {
	e.mu.RLock()
	if e.stopped {
		return "", fmt.Errorf("exporter stopped")
	}
	spanId := trace.SpanFromContext(ctx).SpanContext().SpanID().String()
	v, ok := e.finishedSpans[spanId]
	if !ok {
		return "", fmt.Errorf("no span found: %s", spanId)
	}
    desc := e.describeSpan(v)
    e.mu.RUnlock()
    e.mu.Lock()
    e.deleteChildenAndSelf(spanId)
    e.mu.Unlock()
	return desc, nil
}

func (e *ServerTimingHeaderExporter) deleteChildenAndSelf(spanId string) {
        for key, otherSpan := range e.finishedSpans {
            if otherSpan.Parent().SpanID().String() == spanId {
                e.deleteChildenAndSelf(key) 
            }
        }
        delete(e.finishedSpans, spanId)
    }

func (e *ServerTimingHeaderExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	e.mu.Lock()
	for _, span := range spans {
        spanId := span.SpanContext().SpanID().String()
        if span.Parent().IsValid() {
            e.finishedSpans[spanId] = span
        } else {
        }
	}
	defer e.mu.Unlock()
	return nil
}

func (e *ServerTimingHeaderExporter) describeSpan(span sdktrace.ReadOnlySpan) string {
    suffix := ""
    for key, otherSpan := range e.finishedSpans {
        if otherSpan.Parent().SpanID().String() == span.SpanContext().SpanID().String() {
            suffix = e.describeSpan(e.finishedSpans[key]) 
        }
    }
    durationMillis := float64(span.EndTime().Sub(span.StartTime()).Microseconds()) / 1000
    if suffix != "" {
        return fmt.Sprintf("%s;dur=%.3f, %s", span.Name(), durationMillis, suffix)
    } else {
        return fmt.Sprintf("%s;dur=%.3f", span.Name(), durationMillis)
    }
}

func (e *ServerTimingHeaderExporter) Shutdown(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.stopped = false
	e.finishedSpans = nil
	return nil
}

func NewServerTimingHeaderExporter() *ServerTimingHeaderExporter {
	return &ServerTimingHeaderExporter{
		finishedSpans: map[string]sdktrace.ReadOnlySpan{},
	}
}

var serverTimingExporter = NewServerTimingHeaderExporter()

// scopeInjectingSpanProcessor is a [sdktrace.SpanProcessor]
// that injects [instrumentation.Scope] via [attribute]s using [semconv]
type scopeInjectingSpanProcessor struct{}

func (sp *scopeInjectingSpanProcessor) OnStart(parent context.Context, s sdktrace.ReadWriteSpan) {
	var scope instrumentation.Scope = s.InstrumentationScope()
	s.SetAttributes(
		semconv.OTelScopeName(scope.Name),
		semconv.OTelScopeVersion(scope.Version),
	)
}
func (sp *scopeInjectingSpanProcessor) OnEnd(s sdktrace.ReadOnlySpan)        {}
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

func newTracerProvider(ctx context.Context, resource *resource.Resource) (*sdktrace.TracerProvider, error) {
	remoteExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc trace exporter: %w", err)
	}
	return sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(&scopeInjectingSpanProcessor{}),
		sdktrace.WithBatcher(remoteExporter),
		sdktrace.WithSyncer(serverTimingExporter),
	), nil
}

func newResource(ctx context.Context) (*resource.Resource, error) {
	res, err := resource.Merge(
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
	_, err := stdoutlog.New(
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
		// log.WithProcessor(log.NewBatchProcessor(stdoutExporter)),
		log.WithProcessor(log.NewBatchProcessor(remoteExporter)),
	), nil
}

func newMeterProvider(ctx context.Context, resource *resource.Resource) (*metric.MeterProvider, error) {
	_, err := stdoutmetric.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout metric exporter: %w", err)
	}

	remoteExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc metric exporter: %w", err)
	}

	return metric.NewMeterProvider(
		metric.WithResource(resource),
		// metric.WithReader(metric.NewPeriodicReader(stdoutExporter)),
		metric.WithReader(metric.NewPeriodicReader(remoteExporter)),
	), nil

}

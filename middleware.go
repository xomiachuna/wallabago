package main

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type middleware interface {
    wrap(http.Handler) http.Handler
}

type middlewareStack struct {
    stack []middleware
}

func (s *middlewareStack) wrap(next http.Handler) http.Handler {
    handler := next 
    for i := len(s.stack) - 1; i >= 0; i-- {
        handler = s.stack[i].wrap(handler)
    }
    return handler
}

type middlewareFunc struct {
    f func(http.Handler) http.Handler
}

func (m *middlewareFunc) wrap(next http.Handler) http.Handler {
    return m.f(next)
}

func MiddlewareFunc(f func(http.Handler) http.Handler) middleware {
    return &middlewareFunc{f: f}
}

func NewMiddlewareStack(middlewares ...middleware) *middlewareStack {
    return &middlewareStack{
        stack: middlewares,
    }
}

func setupMiddleware() middleware {
    otelMiddleware := otelhttp.NewMiddleware("handler", otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
        return r.Pattern
    }))
    return NewMiddlewareStack(
        MiddlewareFunc(otelMiddleware),
        serverTimingMiddleware(),
    )
}

func serverTimingMiddleware() middleware {
    return MiddlewareFunc(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, span := tracer.Start(r.Context(), "server-timing")
            w.Header().Set("Trailer", "Server-Timing")
            r = r.WithContext(ctx)
            next.ServeHTTP(w, r)
            span.End()
			timingHeader, err := serverTimingExporter.GetServerTimingHeaderValue(ctx)
			if err != nil {
				logger.DebugContext(ctx, "failed to create server-timing header", "cause", err)
			} else {
                logger.DebugContext(ctx, "timing header", "value", timingHeader)
				w.Header().Add("Server-Timing", timingHeader)
			}
        })
    })
}

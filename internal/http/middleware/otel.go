package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewOtelHTTPMiddleware() Func {
	return func(h http.Handler) http.Handler {
		return otelhttp.NewHandler(
			h,
			"",
			otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
				return r.RequestURI
			}),
			otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		)
	}
}

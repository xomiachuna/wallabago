package middleware

import (
	"log/slog"
	"net/http"
)

type loggingInterceptor struct {
	statusCode  int
	innerWriter http.ResponseWriter
}

var _ http.ResponseWriter = (*loggingInterceptor)(nil)

func (i *loggingInterceptor) Header() http.Header {
	return i.innerWriter.Header()
}

func (i *loggingInterceptor) Write(data []byte) (int, error) {
	return i.innerWriter.Write(data)
}

func (i *loggingInterceptor) WriteHeader(statusCode int) {
	i.statusCode = statusCode
	i.innerWriter.WriteHeader(statusCode)
}

var LoggingMiddleware = Func(func(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		interceptor := &loggingInterceptor{innerWriter: w}
		h.ServeHTTP(interceptor, r)
		slog.InfoContext(r.Context(), "request",
			"method", r.Method,
			"url", r.URL,
			"statusCode", interceptor.statusCode,
			"status", http.StatusText(interceptor.statusCode),
		)
	})
})

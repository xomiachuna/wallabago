package middleware

import (
	"fmt"
	"net/http"
)

// todo: consider intermediate writer
var PanicInterceptMiddleware = Func(func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, fmt.Sprintf("%s", err), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
})

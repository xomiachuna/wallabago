package middleware

import "net/http"

type Middleware interface {
	Wrap(handler http.Handler) http.Handler
}

// Chain represents a chain of middlewares being applied
// in sequential fashion.
type Chain struct {
	middlewares []Middleware
}

func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{
		middlewares: middlewares,
	}
}

func (c *Chain) Wrap(handler http.Handler) http.Handler {
	wrapped := handler
	for _, middleware := range c.middlewares {
		wrapped = middleware.Wrap(wrapped)
	}
	return wrapped
}

// ensure that Chain conforms to Middleware interface.
var _ Middleware = (*Chain)(nil)

// Func is an alias for a simple middleware function.
type Func func(http.Handler) http.Handler

func (mf Func) Wrap(handler http.Handler) http.Handler {
	return mf(handler)
}

// ensure that Func conforms to Middleware interface.
var _ Middleware = Func(nil)
